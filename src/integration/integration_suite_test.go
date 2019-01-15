package integration

import (
	"autoscaler/cf"
	"autoscaler/db"
	"autoscaler/metricscollector/testhelpers"
	"autoscaler/models"
	as_testhelpers "autoscaler/testhelpers"
	"bytes"

	"database/sql"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"mime/multipart"
	"net/http"
	"os"
	"path"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"
	"syscall"
	"testing"
	"time"

	"code.cloudfoundry.org/cfhttp"
	"code.cloudfoundry.org/go-loggregator/rpc/loggregator_v2"
	"code.cloudfoundry.org/lager"
	"github.com/cloudfoundry/sonde-go/events"
	"github.com/gogo/protobuf/proto"
	_ "github.com/lib/pq"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/onsi/gomega/gexec"
	"github.com/onsi/gomega/ghttp"
	"github.com/tedsuo/ifrit"
	"github.com/tedsuo/ifrit/ginkgomon"
	"github.com/tedsuo/ifrit/grouper"
)

type APIType uint8

const (
	INTERNAL APIType = iota
	PUBLIC
)

var (
	components               Components
	tmpDir                   string
	serviceBrokerConfPath    string
	apiServerConfPath        string
	schedulerConfPath        string
	metricsCollectorConfPath string
	eventGeneratorConfPath   string
	scalingEngineConfPath    string
	operatorConfPath         string
	metricsGatewayConfPath   string
	metricsServerConfPath    string
	brokerUserName           string = "username"
	brokerPassword           string = "password"
	brokerAuth               string
	dbUrl                    string
	LOGLEVEL                 string
	noaaPollingRegPath       = regexp.MustCompile(`^/apps/.*/containermetrics$`)
	noaaStreamingRegPath     = regexp.MustCompile(`^/apps/.*/stream$`)
	appSummaryRegPath        = regexp.MustCompile(`^/v2/apps/.*/summary$`)
	appInstanceRegPath       = regexp.MustCompile(`^/v2/apps/.*$`)
	checkUserSpaceRegPath    = regexp.MustCompile(`^/v2/users/.+/spaces.*$`)
	dbHelper                 *sql.DB
	fakeScheduler            *ghttp.Server
	fakeCCNOAAUAA            *ghttp.Server
	messagesToSend           chan []byte
	streamingDoneChan        chan bool
	emptyMessageChannel      chan []byte
	testUserId               string   = "testUserId"
	testUserScope            []string = []string{"cloud_controller.read", "cloud_controller.write", "password.write", "openid", "network.admin", "network.write", "uaa.user"}

	processMap map[string]ifrit.Process = map[string]ifrit.Process{}

	defaultHttpClientTimeout time.Duration = 10 * time.Second

	brokerApiHttpRequestTimeout              time.Duration = 10 * time.Second
	apiSchedulerHttpRequestTimeout           time.Duration = 10 * time.Second
	apiScalingEngineHttpRequestTimeout       time.Duration = 10 * time.Second
	apiMetricsCollectorHttpRequestTimeout    time.Duration = 10 * time.Second
	apiEventGeneratorHttpRequestTimeout      time.Duration = 10 * time.Second
	schedulerScalingEngineHttpRequestTimeout time.Duration = 10 * time.Second

	collectInterval           time.Duration = 1 * time.Second
	refreshInterval           time.Duration = 1 * time.Second
	saveInterval              time.Duration = 1 * time.Second
	aggregatorExecuteInterval time.Duration = 1 * time.Second
	policyPollerInterval      time.Duration = 1 * time.Second
	evaluationManagerInterval time.Duration = 1 * time.Second
	breachDurationSecs        int           = 5

	httpClient             *http.Client
	httpClientForPublicApi *http.Client
	logger                 lager.Logger

	testCertDir string = "../../test-certs"
)

func TestIntegration(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Integration Suite")
}

var _ = SynchronizedBeforeSuite(func() []byte {
	components = Components{
		Ports:       PreparePorts(),
		Executables: CompileTestedExecutables(),
	}
	payload, err := json.Marshal(&components)
	Expect(err).NotTo(HaveOccurred())

	dbUrl = os.Getenv("DBURL")
	if dbUrl == "" {
		Fail("environment variable $DBURL is not set")
	}

	dbHelper, err = sql.Open(db.PostgresDriverName, dbUrl)
	Expect(err).NotTo(HaveOccurred())

	clearDatabase()

	return payload
}, func(encodedBuiltArtifacts []byte) {
	err := json.Unmarshal(encodedBuiltArtifacts, &components)
	Expect(err).NotTo(HaveOccurred())
	components.Ports = PreparePorts()

	tmpDir, err = ioutil.TempDir("", "autoscaler")
	Expect(err).NotTo(HaveOccurred())

	dbUrl = os.Getenv("DBURL")
	dbHelper, err = sql.Open(db.PostgresDriverName, dbUrl)
	Expect(err).NotTo(HaveOccurred())

	LOGLEVEL = os.Getenv("LOGLEVEL")
	if LOGLEVEL == "" {
		LOGLEVEL = "info"
	}
})

var _ = SynchronizedAfterSuite(func() {
	if len(tmpDir) > 0 {
		os.RemoveAll(tmpDir)
	}
}, func() {

})

var _ = BeforeEach(func() {
	httpClient = cfhttp.NewClient()
	httpClientForPublicApi = cfhttp.NewClient()
	logger = lager.NewLogger("test")
	logger.RegisterSink(lager.NewWriterSink(GinkgoWriter, lager.DEBUG))
})

func CompileTestedExecutables() Executables {
	builtExecutables := Executables{}
	rootDir := os.Getenv("GOPATH")
	var err error
	builtExecutables[APIServer] = path.Join(rootDir, "api/index.js")
	builtExecutables[ServiceBroker] = path.Join(rootDir, "servicebroker/lib/index.js")
	builtExecutables[Scheduler] = path.Join(rootDir, "scheduler/target/scheduler-1.0-SNAPSHOT.war")

	builtExecutables[EventGenerator], err = gexec.BuildIn(rootDir, "autoscaler/eventgenerator/cmd/eventgenerator", "-race")
	Expect(err).NotTo(HaveOccurred())

	builtExecutables[MetricsCollector], err = gexec.BuildIn(rootDir, "autoscaler/metricscollector/cmd/metricscollector", "-race")
	Expect(err).NotTo(HaveOccurred())

	builtExecutables[ScalingEngine], err = gexec.BuildIn(rootDir, "autoscaler/scalingengine/cmd/scalingengine", "-race")
	Expect(err).NotTo(HaveOccurred())

	builtExecutables[Operator], err = gexec.BuildIn(rootDir, "autoscaler/operator/cmd/operator", "-race")
	Expect(err).NotTo(HaveOccurred())

	builtExecutables[MetricsGateway], err = gexec.BuildIn(rootDir, "autoscaler/metricsgateway/cmd/metricsgateway", "-race")
	Expect(err).NotTo(HaveOccurred())

	builtExecutables[MetricsServerHTTP], err = gexec.BuildIn(rootDir, "autoscaler/metricsserver/cmd/metricsserver", "-race")
	Expect(err).NotTo(HaveOccurred())

	return builtExecutables
}

func PreparePorts() Ports {
	return Ports{
		APIServer:             10000 + GinkgoParallelNode(),
		APIPublicServer:       12000 + GinkgoParallelNode(),
		ServiceBroker:         13000 + GinkgoParallelNode(),
		ServiceBrokerInternal: 14000 + GinkgoParallelNode(),
		Scheduler:             15000 + GinkgoParallelNode(),
		MetricsCollector:      16000 + GinkgoParallelNode(),
		MetricsServerHTTP:     20000 + GinkgoParallelNode(),
		MetricsServerWS:       21000 + GinkgoParallelNode(),
		EventGenerator:        17000 + GinkgoParallelNode(),
		ScalingEngine:         18000 + GinkgoParallelNode(),
	}
}

func startApiServer() *ginkgomon.Runner {
	runner := components.ApiServer(apiServerConfPath)
	processMap[APIServer] = ginkgomon.Invoke(grouper.NewOrdered(os.Interrupt, grouper.Members{
		{APIServer, runner},
	}))
	return runner
}

func startServiceBroker() *ginkgomon.Runner {
	runner := components.ServiceBroker(serviceBrokerConfPath)
	processMap[ServiceBroker] = ginkgomon.Invoke(grouper.NewOrdered(os.Interrupt, grouper.Members{
		{ServiceBroker, runner},
	}))
	return runner
}

func startScheduler() {
	processMap[Scheduler] = ginkgomon.Invoke(grouper.NewOrdered(os.Interrupt, grouper.Members{
		{Scheduler, components.Scheduler(schedulerConfPath)},
	}))
}

func startMetricsCollector() {
	processMap[MetricsCollector] = ginkgomon.Invoke(grouper.NewOrdered(os.Interrupt, grouper.Members{
		{MetricsCollector, components.MetricsCollector(metricsCollectorConfPath)},
	}))
}

func startEventGenerator() {
	processMap[EventGenerator] = ginkgomon.Invoke(grouper.NewOrdered(os.Interrupt, grouper.Members{
		{EventGenerator, components.EventGenerator(eventGeneratorConfPath)},
	}))
}

func startScalingEngine() {
	processMap[ScalingEngine] = ginkgomon.Invoke(grouper.NewOrdered(os.Interrupt, grouper.Members{
		{ScalingEngine, components.ScalingEngine(scalingEngineConfPath)},
	}))
}

func startOperator() {
	processMap[Operator] = ginkgomon.Invoke(grouper.NewOrdered(os.Interrupt, grouper.Members{
		{Operator, components.Operator(operatorConfPath)},
	}))
}

func startMetricsGateway() {
	processMap[MetricsGateway] = ginkgomon.Invoke(grouper.NewOrdered(os.Interrupt, grouper.Members{
		{MetricsGateway, components.MetricsGateway(metricsGatewayConfPath)},
	}))
}

func startMetricsServer() {
	processMap[MetricsServerHTTP] = ginkgomon.Invoke(grouper.NewOrdered(os.Interrupt, grouper.Members{
		{MetricsServerHTTP, components.MetricsServer(metricsServerConfPath)},
	}))
}

func stopApiServer() {
	ginkgomon.Kill(processMap[APIServer], 5*time.Second)
}
func stopScheduler() {
	ginkgomon.Kill(processMap[Scheduler], 5*time.Second)
}
func stopScalingEngine() {
	ginkgomon.Kill(processMap[ScalingEngine], 5*time.Second)
}
func stopMetricsCollector() {
	ginkgomon.Kill(processMap[MetricsCollector], 5*time.Second)
}
func stopEventGenerator() {
	ginkgomon.Kill(processMap[EventGenerator], 5*time.Second)
}
func stopServiceBroker() {
	ginkgomon.Kill(processMap[ServiceBroker], 5*time.Second)
}
func stopOperator() {
	ginkgomon.Kill(processMap[Operator], 5*time.Second)
}
func stopMetricsGateway() {
	ginkgomon.Kill(processMap[MetricsGateway], 5*time.Second)
}
func stopMetricsServer() {
	ginkgomon.Kill(processMap[MetricsServerHTTP], 5*time.Second)
}

func sendSigusr2Signal(component string) {
	process := processMap[component]
	if process != nil {
		process.Signal(syscall.SIGUSR2)
	}
}

func sendKillSignal(component string) {
	ginkgomon.Kill(processMap[component], 5*time.Second)
}

func stopAll() {
	for _, process := range processMap {
		if process == nil {
			continue
		}
		ginkgomon.Interrupt(process, 15*time.Second)
	}
}

func getRandomId() string {
	return strconv.FormatInt(time.Now().UnixNano(), 10)
}

func initializeHttpClient(certFileName string, keyFileName string, caCertFileName string, httpRequestTimeout time.Duration) {
	TLSConfig, err := cfhttp.NewTLSConfig(
		filepath.Join(testCertDir, certFileName),
		filepath.Join(testCertDir, keyFileName),
		filepath.Join(testCertDir, caCertFileName),
	)
	Expect(err).NotTo(HaveOccurred())
	httpClient.Transport.(*http.Transport).TLSClientConfig = TLSConfig
	httpClient.Timeout = httpRequestTimeout
}
func initializeHttpClientForPublicApi(certFileName string, keyFileName string, caCertFileName string, httpRequestTimeout time.Duration) {
	TLSConfig, err := cfhttp.NewTLSConfig(
		filepath.Join(testCertDir, certFileName),
		filepath.Join(testCertDir, keyFileName),
		filepath.Join(testCertDir, caCertFileName),
	)
	Expect(err).NotTo(HaveOccurred())
	httpClientForPublicApi.Transport.(*http.Transport).TLSClientConfig = TLSConfig
	httpClientForPublicApi.Timeout = httpRequestTimeout
}

func provisionServiceInstance(serviceInstanceId string, orgId string, spaceId string) (*http.Response, error) {
	req, err := http.NewRequest("PUT", fmt.Sprintf("https://127.0.0.1:%d/v2/service_instances/%s", components.Ports[ServiceBroker], serviceInstanceId), strings.NewReader(fmt.Sprintf(`{"organization_guid":"%s","space_guid":"%s"}`, orgId, spaceId)))
	Expect(err).NotTo(HaveOccurred())
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Basic "+brokerAuth)
	return httpClient.Do(req)
}

func deprovisionServiceInstance(serviceInstanceId string) (*http.Response, error) {
	req, err := http.NewRequest("DELETE", fmt.Sprintf("https://127.0.0.1:%d/v2/service_instances/%s", components.Ports[ServiceBroker], serviceInstanceId), strings.NewReader(""))
	Expect(err).NotTo(HaveOccurred())
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Basic "+brokerAuth)
	return httpClient.Do(req)
}

func bindService(bindingId string, appId string, serviceInstanceId string, policy []byte) (*http.Response, error) {
	var bindBody map[string]interface{}
	if policy != nil {
		rawParameters := json.RawMessage(policy)
		bindBody = map[string]interface{}{
			"app_guid":   appId,
			"parameters": &rawParameters,
		}
	} else {
		bindBody = map[string]interface{}{
			"app_guid": appId,
		}
	}

	body, err := json.Marshal(bindBody)
	req, err := http.NewRequest("PUT", fmt.Sprintf("https://127.0.0.1:%d/v2/service_instances/%s/service_bindings/%s", components.Ports[ServiceBroker], serviceInstanceId, bindingId), bytes.NewReader(body))
	Expect(err).NotTo(HaveOccurred())
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Basic "+brokerAuth)
	return httpClient.Do(req)
}

func unbindService(bindingId string, appId string, serviceInstanceId string) (*http.Response, error) {
	req, err := http.NewRequest("DELETE", fmt.Sprintf("https://127.0.0.1:%d/v2/service_instances/%s/service_bindings/%s", components.Ports[ServiceBroker], serviceInstanceId, bindingId), strings.NewReader(fmt.Sprintf(`{"app_guid":"%s"}`, appId)))
	Expect(err).NotTo(HaveOccurred())
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Basic "+brokerAuth)
	return httpClient.Do(req)
}

func provisionAndBind(serviceInstanceId string, orgId string, spaceId string, bindingId string, appId string, policy []byte) {
	resp, err := provisionServiceInstance(serviceInstanceId, orgId, spaceId)
	Expect(err).NotTo(HaveOccurred())
	Expect(resp.StatusCode).To(Equal(http.StatusCreated))
	resp.Body.Close()

	resp, err = bindService(bindingId, appId, serviceInstanceId, policy)
	Expect(err).NotTo(HaveOccurred())
	Expect(resp.StatusCode).To(Equal(http.StatusCreated))
	resp.Body.Close()
}
func unbindAndDeprovision(bindingId string, appId string, serviceInstanceId string) {
	resp, err := unbindService(bindingId, appId, serviceInstanceId)
	Expect(err).NotTo(HaveOccurred())
	Expect(resp.StatusCode).To(Equal(http.StatusOK))
	resp.Body.Close()

	resp, err = deprovisionServiceInstance(serviceInstanceId)
	Expect(err).NotTo(HaveOccurred())
	Expect(resp.StatusCode).To(Equal(http.StatusOK))
	resp.Body.Close()

}
func getPolicy(appId string, apiType APIType) (*http.Response, error) {
	var apiServerPort int
	var httpClientTmp *http.Client
	if apiType == INTERNAL {
		apiServerPort = components.Ports[APIServer]
		httpClientTmp = httpClient
	} else {
		apiServerPort = components.Ports[APIPublicServer]
		httpClientTmp = httpClientForPublicApi
	}
	req, err := http.NewRequest("GET", fmt.Sprintf("https://127.0.0.1:%d/v1/apps/%s/policy", apiServerPort, appId), nil)
	if apiType == PUBLIC {
		req.Header.Set("Authorization", "bearer fake-token")
	}
	Expect(err).NotTo(HaveOccurred())
	return httpClientTmp.Do(req)
}

func detachPolicy(appId string, apiType APIType) (*http.Response, error) {
	var apiServerPort int
	var httpClientTmp *http.Client
	if apiType == INTERNAL {
		apiServerPort = components.Ports[APIServer]
		httpClientTmp = httpClient
	} else {
		apiServerPort = components.Ports[APIPublicServer]
		httpClientTmp = httpClientForPublicApi
	}
	req, err := http.NewRequest("DELETE", fmt.Sprintf("https://127.0.0.1:%d/v1/apps/%s/policy", apiServerPort, appId), strings.NewReader(""))
	Expect(err).NotTo(HaveOccurred())
	req.Header.Set("Content-Type", "application/json")
	if apiType == PUBLIC {
		req.Header.Set("Authorization", "bearer fake-token")
	}
	return httpClientTmp.Do(req)
}

func attachPolicy(appId string, policy []byte, apiType APIType) (*http.Response, error) {
	var apiServerPort int
	var httpClientTmp *http.Client
	if apiType == INTERNAL {
		apiServerPort = components.Ports[APIServer]
		httpClientTmp = httpClient
	} else {
		apiServerPort = components.Ports[APIPublicServer]
		httpClientTmp = httpClientForPublicApi
	}

	req, err := http.NewRequest("PUT", fmt.Sprintf("https://127.0.0.1:%d/v1/apps/%s/policy", apiServerPort, appId), bytes.NewReader(policy))
	Expect(err).NotTo(HaveOccurred())
	req.Header.Set("Content-Type", "application/json")
	if apiType == PUBLIC {
		req.Header.Set("Authorization", "bearer fake-token")
	}
	return httpClientTmp.Do(req)
}

func getSchedules(appId string) (*http.Response, error) {
	req, err := http.NewRequest("GET", fmt.Sprintf("https://127.0.0.1:%d/v1/apps/%s/schedules", components.Ports[Scheduler], appId), strings.NewReader(""))
	Expect(err).NotTo(HaveOccurred())
	req.Header.Set("Content-Type", "application/json")
	return httpClient.Do(req)
}

func createSchedule(appId string, guid string, schedule string) (*http.Response, error) {
	req, err := http.NewRequest("PUT", fmt.Sprintf("https://127.0.0.1:%d/v1/apps/%s/schedules?guid=%s", components.Ports[Scheduler], appId, guid), bytes.NewReader([]byte(schedule)))
	if err != nil {
		panic(err)
	}
	Expect(err).NotTo(HaveOccurred())
	req.Header.Set("Content-Type", "application/json")
	return httpClient.Do(req)
}

func deleteSchedule(appId string) (*http.Response, error) {
	req, err := http.NewRequest("DELETE", fmt.Sprintf("https://127.0.0.1:%d/v1/apps/%s/schedules", components.Ports[Scheduler], appId), strings.NewReader(""))
	Expect(err).NotTo(HaveOccurred())
	req.Header.Set("Content-Type", "application/json")
	return httpClient.Do(req)
}

func getActiveSchedule(appId string) (*http.Response, error) {
	req, err := http.NewRequest("GET", fmt.Sprintf("https://127.0.0.1:%d/v1/apps/%s/active_schedules", components.Ports[ScalingEngine], appId), strings.NewReader(""))
	Expect(err).NotTo(HaveOccurred())
	req.Header.Set("Content-Type", "application/json")
	return httpClient.Do(req)
}

func activeScheduleExists(appId string) bool {
	resp, err := getActiveSchedule(appId)
	Expect(err).NotTo(HaveOccurred())

	return resp.StatusCode == http.StatusOK
}

func setPolicySpecificDateTime(policyByte []byte, start time.Duration, end time.Duration) string {
	timeZone := "GMT"
	location, _ := time.LoadLocation(timeZone)
	timeNowInTimeZone := time.Now().In(location)
	dateTimeFormat := "2006-01-02T15:04"
	startTime := timeNowInTimeZone.Add(start).Format(dateTimeFormat)
	endTime := timeNowInTimeZone.Add(end).Format(dateTimeFormat)

	return fmt.Sprintf(string(policyByte), timeZone, startTime, endTime)
}
func getScalingHistories(pathVariables []string, parameters map[string]string) (*http.Response, error) {
	var apiServerPort int
	var httpClientTmp *http.Client
	apiServerPort = components.Ports[APIPublicServer]
	httpClientTmp = httpClientForPublicApi

	url := "https://127.0.0.1:%d/v1/apps/%s/scaling_histories"
	if parameters != nil && len(parameters) > 0 {
		url += "?any=any"
		for paramName, paramValue := range parameters {
			url += "&" + paramName + "=" + paramValue
		}
	}
	req, err := http.NewRequest("GET", fmt.Sprintf(url, apiServerPort, pathVariables[0]), strings.NewReader(""))
	Expect(err).NotTo(HaveOccurred())
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "bearer fake-token")
	return httpClientTmp.Do(req)
}
func getAppInstanceMetrics(pathVariables []string, parameters map[string]string) (*http.Response, error) {
	var apiServerPort int
	var httpClientTmp *http.Client
	apiServerPort = components.Ports[APIPublicServer]
	httpClientTmp = httpClientForPublicApi
	url := "https://127.0.0.1:%d/v1/apps/%s/metric_histories/%s"
	if parameters != nil && len(parameters) > 0 {
		url += "?any=any"
		for paramName, paramValue := range parameters {
			url += "&" + paramName + "=" + paramValue
		}
	}
	req, err := http.NewRequest("GET", fmt.Sprintf(url, apiServerPort, pathVariables[0], pathVariables[1]), strings.NewReader(""))
	Expect(err).NotTo(HaveOccurred())
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "bearer fake-token")
	return httpClientTmp.Do(req)
}

func getAppAggregatedMetrics(pathVariables []string, parameters map[string]string) (*http.Response, error) {
	var apiServerPort int
	var httpClientTmp *http.Client
	apiServerPort = components.Ports[APIPublicServer]
	httpClientTmp = httpClientForPublicApi
	url := "https://127.0.0.1:%d/v1/apps/%s/aggregated_metric_histories/%s"
	if parameters != nil && len(parameters) > 0 {
		url += "?any=any"
		for paramName, paramValue := range parameters {
			url += "&" + paramName + "=" + paramValue
		}
	}
	req, err := http.NewRequest("GET", fmt.Sprintf(url, apiServerPort, pathVariables[0], pathVariables[1]), strings.NewReader(""))
	Expect(err).NotTo(HaveOccurred())
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "bearer fake-token")
	return httpClientTmp.Do(req)
}

func readPolicyFromFile(filename string) []byte {
	content, err := ioutil.ReadFile(filename)
	Expect(err).NotTo(HaveOccurred())
	return content
}

func clearDatabase() {
	_, err := dbHelper.Exec("DELETE FROM policy_json")
	Expect(err).NotTo(HaveOccurred())

	_, err = dbHelper.Exec("DELETE FROM binding")
	Expect(err).NotTo(HaveOccurred())

	_, err = dbHelper.Exec("DELETE FROM service_instance")
	Expect(err).NotTo(HaveOccurred())

	_, err = dbHelper.Exec("DELETE FROM app_scaling_recurring_schedule")
	Expect(err).NotTo(HaveOccurred())

	_, err = dbHelper.Exec("DELETE FROM app_scaling_specific_date_schedule")
	Expect(err).NotTo(HaveOccurred())

	_, err = dbHelper.Exec("DELETE FROM app_scaling_active_schedule")
	Expect(err).NotTo(HaveOccurred())

	_, err = dbHelper.Exec("DELETE FROM activeschedule")
	Expect(err).NotTo(HaveOccurred())

	_, err = dbHelper.Exec("DELETE FROM scalinghistory")
	Expect(err).NotTo(HaveOccurred())

	_, err = dbHelper.Exec("DELETE FROM app_metric")
	Expect(err).NotTo(HaveOccurred())

	_, err = dbHelper.Exec("DELETE FROM appinstancemetrics")
	Expect(err).NotTo(HaveOccurred())
}

func insertPolicy(appId string, policyStr string, guid string) {
	query := "INSERT INTO policy_json(app_id, policy_json, guid) VALUES($1, $2, $3)"
	_, err := dbHelper.Exec(query, appId, policyStr, guid)
	Expect(err).NotTo(HaveOccurred())

}

func deletePolicy(appId string) {
	query := "DELETE FROM policy_json WHERE app_id=$1"
	_, err := dbHelper.Exec(query, appId)
	Expect(err).NotTo(HaveOccurred())
}

func insertScalingHistory(history *models.AppScalingHistory) {
	query := "INSERT INTO scalinghistory" +
		"(appid, timestamp, scalingtype, status, oldinstances, newinstances, reason, message, error) " +
		" VALUES($1, $2, $3, $4, $5, $6, $7, $8, $9)"
	_, err := dbHelper.Exec(query, history.AppId, history.Timestamp, history.ScalingType, history.Status,
		history.OldInstances, history.NewInstances, history.Reason, history.Message, history.Error)

	Expect(err).NotTo(HaveOccurred())
}
func getScalingHistoryCount(appId string, oldInstanceCount int, newInstanceCount int) int {
	var count int
	query := "SELECT COUNT(*) FROM scalinghistory WHERE appid=$1 AND oldinstances=$2 AND newinstances=$3"
	err := dbHelper.QueryRow(query, appId, oldInstanceCount, newInstanceCount).Scan(&count)
	Expect(err).NotTo(HaveOccurred())
	return count
}
func getScalingHistoryTotalCount(appId string) int {
	var count int
	query := "SELECT COUNT(*) FROM scalinghistory WHERE appid=$1"
	err := dbHelper.QueryRow(query, appId).Scan(&count)
	Expect(err).NotTo(HaveOccurred())
	return count
}
func insertAppInstanceMetric(appInstanceMetric *models.AppInstanceMetric) {
	query := "INSERT INTO appinstancemetrics" +
		"(appid, instanceindex, collectedat, name, unit, value, timestamp) " +
		"VALUES($1, $2, $3, $4, $5, $6, $7)"
	_, err := dbHelper.Exec(query, appInstanceMetric.AppId, appInstanceMetric.InstanceIndex, appInstanceMetric.CollectedAt, appInstanceMetric.Name, appInstanceMetric.Unit, appInstanceMetric.Value, appInstanceMetric.Timestamp)
	Expect(err).NotTo(HaveOccurred())
}
func insertAppMetric(appMetrics *models.AppMetric) {
	query := "INSERT INTO app_metric" +
		"(app_id, metric_type, unit, value, timestamp) " +
		"VALUES($1, $2, $3, $4, $5)"
	_, err := dbHelper.Exec(query, appMetrics.AppId, appMetrics.MetricType, appMetrics.Unit, appMetrics.Value, appMetrics.Timestamp)
	Expect(err).NotTo(HaveOccurred())
}

func getAppInstanceMetricTotalCount(appId string) int {
	var count int
	query := "SELECT COUNT(*) FROM appinstancemetrics WHERE appid=$1"
	err := dbHelper.QueryRow(query, appId).Scan(&count)
	Expect(err).NotTo(HaveOccurred())
	return count
}

func getAppMetricTotalCount(appId string) int {
	var count int
	query := "SELECT COUNT(*) FROM app_metric WHERE app_id=$1"
	err := dbHelper.QueryRow(query, appId).Scan(&count)
	Expect(err).NotTo(HaveOccurred())
	return count
}

func getCredentialsCount(appId string) int {
	var count int
	query := "SELECT COUNT(*) FROM credentials WHERE id=$1"
	err := dbHelper.QueryRow(query, appId).Scan(&count)
	Expect(err).NotTo(HaveOccurred())
	return count
}

type GetResponse func(id string, apiType APIType) (*http.Response, error)
type GetResponseWithParameters func(pathVariables []string, parameters map[string]string) (*http.Response, error)

func checkResponseContent(getResponse GetResponse, id string, expectHttpStatus int, expectResponseMap map[string]interface{}, apiType APIType) {
	resp, err := getResponse(id, apiType)
	checkResponse(resp, err, expectHttpStatus, expectResponseMap)

}
func checkPublicAPIResponseContentWithParameters(getResponseWithParameters GetResponseWithParameters, pathVariables []string, parameters map[string]string, expectHttpStatus int, expectResponseMap map[string]interface{}) {
	resp, err := getResponseWithParameters(pathVariables, parameters)
	checkResponse(resp, err, expectHttpStatus, expectResponseMap)
}
func checkResponse(resp *http.Response, err error, expectHttpStatus int, expectResponseMap map[string]interface{}) {
	Expect(err).NotTo(HaveOccurred())
	Expect(resp.StatusCode).To(Equal(expectHttpStatus))
	var actual map[string]interface{}
	err = json.NewDecoder(resp.Body).Decode(&actual)
	Expect(err).NotTo(HaveOccurred())
	Expect(actual).To(Equal(expectResponseMap))
	resp.Body.Close()
}

func checkResponseEmptyAndStatusCode(resp *http.Response, err error, expectedStatus int) {
	Expect(err).NotTo(HaveOccurred())
	defer resp.Body.Close()
	body, err := ioutil.ReadAll(resp.Body)
	Expect(err).NotTo(HaveOccurred())
	Expect(body).To(HaveLen(0))
	Expect(resp.StatusCode).To(Equal(expectedStatus))
}

func checkSchedule(appId string, expectHttpStatus int, expectResponseMap map[string]int) bool {
	resp, err := getSchedules(appId)
	Expect(err).NotTo(HaveOccurred())
	Expect(resp.StatusCode).To(Equal(expectHttpStatus))
	defer resp.Body.Close()
	var actual map[string]interface{}
	err = json.NewDecoder(resp.Body).Decode(&actual)
	Expect(err).NotTo(HaveOccurred())
	var schedules map[string]interface{} = actual["schedules"].(map[string]interface{})
	var recurring []interface{} = schedules["recurring_schedule"].([]interface{})
	var specificDate []interface{} = schedules["specific_date"].([]interface{})
	return len(specificDate) == expectResponseMap["specific_date"] && len(recurring) == expectResponseMap["recurring_schedule"]
}

func startFakeCCNOAAUAA(instanceCount int) {
	fakeCCNOAAUAA = ghttp.NewServer()
	fakeCCNOAAUAA.RouteToHandler("GET", "/v2/info", ghttp.RespondWithJSONEncoded(http.StatusOK,
		cf.Endpoints{
			AuthEndpoint:    fakeCCNOAAUAA.URL(),
			TokenEndpoint:   fakeCCNOAAUAA.URL(),
			DopplerEndpoint: strings.Replace(fakeCCNOAAUAA.URL(), "http", "ws", 1),
		}))
	fakeCCNOAAUAA.RouteToHandler("POST", "/oauth/token", ghttp.RespondWithJSONEncoded(http.StatusOK, cf.Tokens{}))
	appState := models.AppStatusStarted
	fakeCCNOAAUAA.RouteToHandler("GET", appSummaryRegPath, ghttp.RespondWithJSONEncoded(http.StatusOK,
		models.AppEntity{Instances: instanceCount, State: &appState}))
	fakeCCNOAAUAA.RouteToHandler("PUT", appInstanceRegPath, ghttp.RespondWith(http.StatusCreated, ""))
	fakeCCNOAAUAA.RouteToHandler("POST", "/check_token", ghttp.RespondWithJSONEncoded(http.StatusOK,
		struct {
			Scope []string `json:"scope"`
		}{
			testUserScope,
		}))
	fakeCCNOAAUAA.RouteToHandler("GET", "/userinfo", ghttp.RespondWithJSONEncoded(http.StatusOK,
		struct {
			UserId string `json:"user_id"`
		}{
			testUserId,
		}))
	fakeCCNOAAUAA.RouteToHandler("GET", checkUserSpaceRegPath, ghttp.RespondWithJSONEncoded(http.StatusOK,
		struct {
			TotalResults int `json:"total_results"`
		}{
			1,
		}))
}
func fakeMetricsPolling(appId string, memoryValue uint64, memQuota uint64) {
	fakeCCNOAAUAA.RouteToHandler("GET", noaaPollingRegPath,
		func(rw http.ResponseWriter, r *http.Request) {
			mp := multipart.NewWriter(rw)
			defer mp.Close()

			rw.Header().Set("Content-Type", `multipart/x-protobuf; boundary=`+mp.Boundary())
			timestamp := time.Now().UnixNano()
			message1 := marshalMessage(createContainerMetric(appId, 0, 3.0, memoryValue, 2048000000, memQuota, 4096000000, timestamp))
			message2 := marshalMessage(createContainerMetric(appId, 1, 4.0, memoryValue, 2048000000, memQuota, 4096000000, timestamp))
			message3 := marshalMessage(createContainerMetric(appId, 2, 5.0, memoryValue, 2048000000, memQuota, 4096000000, timestamp))

			messages := [][]byte{message1, message2, message3}
			for _, msg := range messages {
				partWriter, _ := mp.CreatePart(nil)
				partWriter.Write(msg)
			}
		},
	)

}

func fakeMetricsStreaming(appId string, memoryValue uint64, memQuota uint64) {
	messagesToSend = make(chan []byte, 256)
	wsHandler := testhelpers.NewWebsocketHandler(messagesToSend, 100*time.Millisecond)
	fakeCCNOAAUAA.RouteToHandler("GET", "/apps/"+appId+"/stream", wsHandler.ServeWebsocket)

	streamingDoneChan = make(chan bool)
	ticker := time.NewTicker(500 * time.Millisecond)
	go func() {
		select {
		case <-streamingDoneChan:
			ticker.Stop()
			return
		case <-ticker.C:
			timestamp := time.Now().UnixNano()
			message1 := marshalMessage(createContainerMetric(appId, 0, 3.0, memoryValue, 2048000000, memQuota, 4096000000, timestamp-int64(time.Duration(breachDurationSecs)*time.Second)))
			messagesToSend <- message1
			message2 := marshalMessage(createContainerMetric(appId, 1, 4.0, memoryValue, 2048000000, memQuota, 4096000000, timestamp))
			messagesToSend <- message2
			message3 := marshalMessage(createContainerMetric(appId, 2, 5.0, memoryValue, 2048000000, memQuota, 4096000000, timestamp))
			messagesToSend <- message3
			message4 := marshalMessage(createContainerMetric(appId, 2, 5.0, memoryValue, 2048000000, memQuota, 4096000000, timestamp))
			messagesToSend <- message4
			message5 := marshalMessage(createContainerMetric(appId, 2, 5.0, memoryValue, 2048000000, memQuota, 4096000000, timestamp+int64(time.Duration(breachDurationSecs)*time.Second)))
			messagesToSend <- message5
		}
	}()

	emptyMessageChannel = make(chan []byte, 256)
	emptyWsHandler := testhelpers.NewWebsocketHandler(emptyMessageChannel, 200*time.Millisecond)
	fakeCCNOAAUAA.RouteToHandler("GET", noaaStreamingRegPath, emptyWsHandler.ServeWebsocket)

}

func closeFakeMetricsStreaming() {
	close(streamingDoneChan)
	close(messagesToSend)
	close(emptyMessageChannel)
}

func startFakeRLPServer(appId string, envelopes []*loggregator_v2.Envelope, emitInterval time.Duration) *as_testhelpers.FakeEventProducer {
	fakeRLPServer, err := as_testhelpers.NewFakeEventProducer(filepath.Join(testCertDir, "reverselogproxy.crt"), filepath.Join(testCertDir, "reverselogproxy.key"), filepath.Join(testCertDir, "autoscaler-ca.crt"), emitInterval)
	Expect(err).NotTo(HaveOccurred())
	fakeRLPServer.SetEnvelops(envelopes)
	fakeRLPServer.Start()
	return fakeRLPServer
}
func stopFakeRLPServer(fakeRLPServer *as_testhelpers.FakeEventProducer) {
	stopped := fakeRLPServer.Stop()
	Expect(stopped).To(Equal(true))
}

func createContainerMetric(appId string, instanceIndex int32, cpuPercentage float64, memoryBytes uint64, diskByte uint64, memQuota uint64, diskQuota uint64, timestamp int64) *events.Envelope {
	if timestamp == 0 {
		timestamp = time.Now().UnixNano()
	}
	cm := &events.ContainerMetric{
		ApplicationId:    proto.String(appId),
		InstanceIndex:    proto.Int32(instanceIndex),
		CpuPercentage:    proto.Float64(cpuPercentage),
		MemoryBytes:      proto.Uint64(memoryBytes),
		DiskBytes:        proto.Uint64(diskByte),
		MemoryBytesQuota: proto.Uint64(memQuota),
		DiskBytesQuota:   proto.Uint64(diskQuota),
	}

	return &events.Envelope{
		ContainerMetric: cm,
		EventType:       events.Envelope_ContainerMetric.Enum(),
		Origin:          proto.String("fake-origin-1"),
		Timestamp:       proto.Int64(timestamp),
	}
}
func createContainerEnvelope(appId string, instanceIndex int32, cpuPercentage float64, memoryBytes float64, diskByte float64, memQuota float64, timestamp int64) []*loggregator_v2.Envelope {
	return []*loggregator_v2.Envelope{
		&loggregator_v2.Envelope{
			SourceId:  appId,
			Timestamp: timestamp,
			Message: &loggregator_v2.Envelope_Gauge{
				Gauge: &loggregator_v2.Gauge{
					Metrics: map[string]*loggregator_v2.GaugeValue{
						"cpu": &loggregator_v2.GaugeValue{
							Unit:  "percentage",
							Value: cpuPercentage,
						},
						"disk": &loggregator_v2.GaugeValue{
							Unit:  "bytes",
							Value: diskByte,
						},
						"memory": &loggregator_v2.GaugeValue{
							Unit:  "bytes",
							Value: memoryBytes,
						},
						"memory_quota": &loggregator_v2.GaugeValue{
							Unit:  "bytes",
							Value: memQuota,
						},
					},
				},
			},
		},
	}
}
func createHTTPTimerEnvelope(appId string, start int64, end int64) []*loggregator_v2.Envelope {
	return []*loggregator_v2.Envelope{
		&loggregator_v2.Envelope{
			SourceId: appId,
			Message: &loggregator_v2.Envelope_Timer{
				Timer: &loggregator_v2.Timer{
					Name:  "http",
					Start: start,
					Stop:  end,
				},
			},
		},
	}

}
func createCustomEnvelope(appId string, name string, unit string, value float64, timestamp int64) []*loggregator_v2.Envelope {
	return []*loggregator_v2.Envelope{
		&loggregator_v2.Envelope{
			SourceId:  appId,
			Timestamp: timestamp,
			Tags:      map[string]string{"origin": "autoscaler_metrics_forwarder"},
			Message: &loggregator_v2.Envelope_Gauge{
				Gauge: &loggregator_v2.Gauge{
					Metrics: map[string]*loggregator_v2.GaugeValue{
						name: &loggregator_v2.GaugeValue{
							Unit:  unit,
							Value: value,
						},
					},
				},
			},
		},
	}

}

func marshalMessage(message *events.Envelope) []byte {
	data, err := proto.Marshal(message)
	if err != nil {
		log.Println(err.Error())
	}

	return data
}

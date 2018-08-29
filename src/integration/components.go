package integration

import (
	"autoscaler/cf"
	"autoscaler/db"
	egConfig "autoscaler/eventgenerator/config"
	"autoscaler/helpers"
	mcConfig "autoscaler/metricscollector/config"
	"autoscaler/models"
	opConfig "autoscaler/operator/config"
	seConfig "autoscaler/scalingengine/config"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/url"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"time"

	. "github.com/onsi/gomega"
	"github.com/tedsuo/ifrit/ginkgomon"
	"gopkg.in/yaml.v2"
)

const (
	APIServer             = "apiServer"
	APIPublicServer       = "APIPublicServer"
	ServiceBroker         = "serviceBroker"
	ServiceBrokerInternal = "serviceBrokerInternal"
	Scheduler             = "scheduler"
	MetricsCollector      = "metricsCollector"
	EventGenerator        = "eventGenerator"
	ScalingEngine         = "scalingEngine"
	Operator              = "operator"
	ConsulCluster         = "consulCluster"
)

var serviceCatalogPath string = "../../servicebroker/config/catalog.json"
var schemaValidationPath string = "../../servicebroker/config/catalog.schema.json"
var apiServerInfoFilePath string = "../../api/config/info.json"

type Executables map[string]string
type Ports map[string]int

type Components struct {
	Executables Executables
	Ports       Ports
}

type DBConfig struct {
	URI            string `json:"uri"`
	MinConnections int    `json:"minConnections"`
	MaxConnections int    `json:"maxConnections"`
	IdleTimeout    int    `json:"idleTimeout"`
}
type APIServerClient struct {
	Uri string          `json:"uri"`
	TLS models.TLSCerts `json:"tls"`
}

type ServiceBrokerConfig struct {
	Port       int `json:"port"`
	PublicPort int `json:"publicPort"`
	EnableCustomMetrics bool `json:"enableCustomMetrics"`

	Username string `json:"username"`
	Password string `json:"password"`

	DB DBConfig `json:"db"`

	APIServerClient      APIServerClient `json:"apiserver"`
	HttpRequestTimeout   int             `json:"httpRequestTimeout"`
	TLS                  models.TLSCerts `json:"tls"`
	PublicTLS            models.TLSCerts `json:"publicTls"`
	ServiceCatalogPath   string          `json:"serviceCatalogPath"`
	SchemaValidationPath string          `json:"schemaValidationPath"`
}
type SchedulerClient struct {
	Uri string          `json:"uri"`
	TLS models.TLSCerts `json:"tls"`
}
type ScalingEngineClient struct {
	Uri string          `json:"uri"`
	TLS models.TLSCerts `json:"tls"`
}
type MetricsCollectorClient struct {
	Uri string          `json:"uri"`
	TLS models.TLSCerts `json:"tls"`
}
type EventGeneratorClient struct {
	Uri string          `json:"uri"`
	TLS models.TLSCerts `json:"tls"`
}
type ServiceOffering struct {
	Enabled             bool                `json:"enabled"`
	ServiceBrokerClient ServiceBrokerClient `json:"serviceBroker"`
}
type ServiceBrokerClient struct {
	Uri string          `json:"uri"`
	TLS models.TLSCerts `json:"tls"`
}
type APIServerConfig struct {
	Port                   int                    `json:"port"`
	PublicPort             int                    `json:"publicPort"`
	InfoFilePath           string                 `json:"infoFilePath"`
	CFAPI                  string                 `json:"cfApi"`
	SkipSSLValidation      bool                   `json:"skipSSLValidation"`
	CacheTTL               int                    `json:"cacheTTL"`
	DB                     DBConfig               `json:"db"`
	SchedulerClient        SchedulerClient        `json:"scheduler"`
	ScalingEngineClient    ScalingEngineClient    `json:"scalingEngine"`
	MetricsCollectorClient MetricsCollectorClient `json:"metricsCollector"`
	EventGeneratorClient   EventGeneratorClient   `json:"eventGenerator"`
	ServiceOffering        ServiceOffering        `json:"serviceOffering"`

	TLS       models.TLSCerts `json:"tls"`
	PublicTLS models.TLSCerts `json:"publicTls"`
}

func (components *Components) ServiceBroker(confPath string, argv ...string) *ginkgomon.Runner {
	return ginkgomon.New(ginkgomon.Config{
		Name:              ServiceBroker,
		AnsiColorCode:     "32m",
		StartCheck:        "Service broker app is running",
		StartCheckTimeout: 20 * time.Second,
		Command: exec.Command(
			"node", append([]string{components.Executables[ServiceBroker], "-c", confPath}, argv...)...,
		),
		Cleanup: func() {
		},
	})
}

func (components *Components) ApiServer(confPath string, argv ...string) *ginkgomon.Runner {
	return ginkgomon.New(ginkgomon.Config{
		Name:              APIServer,
		AnsiColorCode:     "33m",
		StartCheck:        "Autoscaler API server started",
		StartCheckTimeout: 20 * time.Second,
		Command: exec.Command(
			"node", append([]string{components.Executables[APIServer], "-c", confPath}, argv...)...,
		),
		Cleanup: func() {
		},
	})
}

func (components *Components) Scheduler(confPath string, argv ...string) *ginkgomon.Runner {
	return ginkgomon.New(ginkgomon.Config{
		Name:              Scheduler,
		AnsiColorCode:     "34m",
		StartCheck:        "Scheduler is ready to start",
		StartCheckTimeout: 120 * time.Second,
		Command: exec.Command(
			"java", append([]string{"-jar", "-Dspring.config.location=" + confPath, "-Dserver.port=" + strconv.FormatInt(int64(components.Ports[Scheduler]), 10), components.Executables[Scheduler]}, argv...)...,
		),
		Cleanup: func() {
		},
	})
}

func (components *Components) MetricsCollector(confPath string, argv ...string) *ginkgomon.Runner {

	return ginkgomon.New(ginkgomon.Config{
		Name:              MetricsCollector,
		AnsiColorCode:     "35m",
		StartCheck:        `"metricscollector.started"`,
		StartCheckTimeout: 20 * time.Second,
		Command: exec.Command(
			components.Executables[MetricsCollector],
			append([]string{
				"-c", confPath,
			}, argv...)...,
		),
	})
}

func (components *Components) EventGenerator(confPath string, argv ...string) *ginkgomon.Runner {

	return ginkgomon.New(ginkgomon.Config{
		Name:              EventGenerator,
		AnsiColorCode:     "36m",
		StartCheck:        `"eventgenerator.started"`,
		StartCheckTimeout: 20 * time.Second,
		Command: exec.Command(
			components.Executables[EventGenerator],
			append([]string{
				"-c", confPath,
			}, argv...)...,
		),
	})
}

func (components *Components) ScalingEngine(confPath string, argv ...string) *ginkgomon.Runner {

	return ginkgomon.New(ginkgomon.Config{
		Name:              ScalingEngine,
		AnsiColorCode:     "37m",
		StartCheck:        `"scalingengine.started"`,
		StartCheckTimeout: 20 * time.Second,
		Command: exec.Command(
			components.Executables[ScalingEngine],
			append([]string{
				"-c", confPath,
			}, argv...)...,
		),
	})
}

func (components *Components) Operator(confPath string, argv ...string) *ginkgomon.Runner {

	return ginkgomon.New(ginkgomon.Config{
		Name:              Operator,
		AnsiColorCode:     "38m",
		StartCheck:        `"operator.started"`,
		StartCheckTimeout: 20 * time.Second,
		Command: exec.Command(
			components.Executables[Operator],
			append([]string{
				"-c", confPath,
			}, argv...)...,
		),
	})
}

func (components *Components) PrepareServiceBrokerConfig(publicPort int, internalPort int, username string, password string, enableCustomMetrics bool, dbUri string, apiServerUri string, brokerApiHttpRequestTimeout time.Duration, tmpDir string) string {
	brokerConfig := ServiceBrokerConfig{
		Port:       internalPort,
		PublicPort: publicPort,
		Username:   username,
		Password:   password,
		EnableCustomMetrics: enableCustomMetrics,
		DB: DBConfig{
			URI:            dbUri,
			MinConnections: 1,
			MaxConnections: 10,
			IdleTimeout:    1000,
		},
		APIServerClient: APIServerClient{
			Uri: apiServerUri,
			TLS: models.TLSCerts{
				KeyFile:    filepath.Join(testCertDir, "api.key"),
				CertFile:   filepath.Join(testCertDir, "api.crt"),
				CACertFile: filepath.Join(testCertDir, "autoscaler-ca.crt"),
			},
		},
		HttpRequestTimeout: int(brokerApiHttpRequestTimeout / time.Millisecond),
		PublicTLS: models.TLSCerts{
			KeyFile:    filepath.Join(testCertDir, "servicebroker.key"),
			CertFile:   filepath.Join(testCertDir, "servicebroker.crt"),
			CACertFile: filepath.Join(testCertDir, "autoscaler-ca.crt"),
		},
		TLS: models.TLSCerts{
			KeyFile:    filepath.Join(testCertDir, "servicebroker_internal.key"),
			CertFile:   filepath.Join(testCertDir, "servicebroker_internal.crt"),
			CACertFile: filepath.Join(testCertDir, "autoscaler-ca.crt"),
		},
		ServiceCatalogPath:   serviceCatalogPath,
		SchemaValidationPath: schemaValidationPath,
	}

	cfgFile, err := ioutil.TempFile(tmpDir, ServiceBroker)
	w := json.NewEncoder(cfgFile)
	err = w.Encode(brokerConfig)
	Expect(err).NotTo(HaveOccurred())
	cfgFile.Close()
	return cfgFile.Name()
}

func (components *Components) PrepareApiServerConfig(port int, publicPort int, skipSSLValidation bool, cacheTTL int, cfApi string, dbUri string, schedulerUri string, scalingEngineUri string, metricsCollectorUri string, eventGeneratorUri string, serviceBrokerUri string, serviceOfferingEnabled bool, tmpDir string) string {

	apiConfig := APIServerConfig{
		Port:              port,
		PublicPort:        publicPort,
		InfoFilePath:      apiServerInfoFilePath,
		CFAPI:             cfApi,
		SkipSSLValidation: skipSSLValidation,
		CacheTTL:          cacheTTL,
		DB: DBConfig{
			URI:            dbUri,
			MinConnections: 1,
			MaxConnections: 10,
			IdleTimeout:    1000,
		},

		SchedulerClient: SchedulerClient{
			Uri: schedulerUri,
			TLS: models.TLSCerts{
				KeyFile:    filepath.Join(testCertDir, "scheduler.key"),
				CertFile:   filepath.Join(testCertDir, "scheduler.crt"),
				CACertFile: filepath.Join(testCertDir, "autoscaler-ca.crt"),
			},
		},
		ScalingEngineClient: ScalingEngineClient{
			Uri: scalingEngineUri,
			TLS: models.TLSCerts{
				KeyFile:    filepath.Join(testCertDir, "scalingengine.key"),
				CertFile:   filepath.Join(testCertDir, "scalingengine.crt"),
				CACertFile: filepath.Join(testCertDir, "autoscaler-ca.crt"),
			},
		},
		MetricsCollectorClient: MetricsCollectorClient{
			Uri: metricsCollectorUri,
			TLS: models.TLSCerts{
				KeyFile:    filepath.Join(testCertDir, "metricscollector.key"),
				CertFile:   filepath.Join(testCertDir, "metricscollector.crt"),
				CACertFile: filepath.Join(testCertDir, "autoscaler-ca.crt"),
			},
		},
		EventGeneratorClient: EventGeneratorClient{
			Uri: eventGeneratorUri,
			TLS: models.TLSCerts{
				KeyFile:    filepath.Join(testCertDir, "eventgenerator.key"),
				CertFile:   filepath.Join(testCertDir, "eventgenerator.crt"),
				CACertFile: filepath.Join(testCertDir, "autoscaler-ca.crt"),
			},
		},
		ServiceOffering: ServiceOffering{
			Enabled: serviceOfferingEnabled,
			ServiceBrokerClient: ServiceBrokerClient{
				Uri: serviceBrokerUri,
				TLS: models.TLSCerts{
					KeyFile:    filepath.Join(testCertDir, "servicebroker_internal.key"),
					CertFile:   filepath.Join(testCertDir, "servicebroker_internal.crt"),
					CACertFile: filepath.Join(testCertDir, "autoscaler-ca.crt"),
				},
			},
		},

		TLS: models.TLSCerts{
			KeyFile:    filepath.Join(testCertDir, "api.key"),
			CertFile:   filepath.Join(testCertDir, "api.crt"),
			CACertFile: filepath.Join(testCertDir, "autoscaler-ca.crt"),
		},

		PublicTLS: models.TLSCerts{
			KeyFile:    filepath.Join(testCertDir, "api_public.key"),
			CertFile:   filepath.Join(testCertDir, "api_public.crt"),
			CACertFile: filepath.Join(testCertDir, "autoscaler-ca.crt"),
		},
	}

	cfgFile, err := ioutil.TempFile(tmpDir, APIServer)
	w := json.NewEncoder(cfgFile)
	err = w.Encode(apiConfig)
	Expect(err).NotTo(HaveOccurred())
	cfgFile.Close()
	return cfgFile.Name()
}

func (components *Components) PrepareSchedulerConfig(dbUri string, scalingEngineUri string, tmpDir string, consulPort string) string {
	dbUrl, _ := url.Parse(dbUri)
	scheme := dbUrl.Scheme
	host := dbUrl.Host
	path := dbUrl.Path
	userInfo := dbUrl.User
	userName := userInfo.Username()
	password, _ := userInfo.Password()
	if scheme == "postgres" {
		scheme = "postgresql"
	}
	jdbcDBUri := fmt.Sprintf("jdbc:%s://%s%s", scheme, host, path)
	settingStrTemplate := `
#datasource for application and quartz
spring.datasource.driverClassName=org.postgresql.Driver
spring.datasource.url=%s
spring.datasource.username=%s
spring.datasource.password=%s
#policy db
spring.policyDbDataSource.driverClassName=org.postgresql.Driver
spring.policyDbDataSource.url=%s
spring.policyDbDataSource.username=%s
spring.policyDbDataSource.password=%s
#quartz job
scalingenginejob.reschedule.interval.millisecond=10000
scalingenginejob.reschedule.maxcount=3
scalingengine.notification.reschedule.maxcount=3
# scaling engine url
autoscaler.scalingengine.url=%s
#ssl
server.ssl.key-store=%s/scheduler.p12
server.ssl.key-alias=scheduler
server.ssl.key-store-password=123456
server.ssl.key-store-type=PKCS12
server.ssl.trust-store=%s/autoscaler.truststore
server.ssl.trust-store-password=123456
client.ssl.key-store=%s/scheduler.p12
client.ssl.key-store-password=123456
client.ssl.key-store-type=PKCS12
client.ssl.trust-store=%s/autoscaler.truststore
client.ssl.trust-store-password=123456
client.ssl.protocol=TLSv1.2
#Quartz
org.quartz.scheduler.instanceName=app-autoscaler-%d
org.quartz.scheduler.instanceId=app-autoscaler-%d

spring.application.name=scheduler
spring.mvc.servlet.load-on-startup=1
spring.aop.auto=false
endpoints.enabled=false
spring.data.jpa.repositories.enabled=false
`
	settingJsonStr := fmt.Sprintf(settingStrTemplate, jdbcDBUri, userName, password, jdbcDBUri, userName, password, scalingEngineUri, testCertDir, testCertDir, testCertDir, testCertDir, components.Ports[Scheduler], components.Ports[Scheduler], consulPort)
	cfgFile, err := os.Create(filepath.Join(tmpDir, "application.properties"))
	Expect(err).NotTo(HaveOccurred())
	ioutil.WriteFile(cfgFile.Name(), []byte(settingJsonStr), 0777)
	cfgFile.Close()
	return cfgFile.Name()
}

func (components *Components) PrepareMetricsCollectorConfig(dbUri string, port int, ccNOAAUAAUrl string, cfGrantTypePassword string, collectInterval time.Duration,
	refreshInterval time.Duration, saveInterval time.Duration, collectMethod string, tmpDir string) string {
	cfg := mcConfig.Config{
		Cf: cf.CfConfig{
			Api:       ccNOAAUAAUrl,
			GrantType: cfGrantTypePassword,
			Username:  "admin",
			Password:  "admin",
		},
		Server: mcConfig.ServerConfig{
			Port: port,
			TLS: models.TLSCerts{
				KeyFile:    filepath.Join(testCertDir, "metricscollector.key"),
				CertFile:   filepath.Join(testCertDir, "metricscollector.crt"),
				CACertFile: filepath.Join(testCertDir, "autoscaler-ca.crt"),
			},
			NodeAddrs: []string{"localhost"},
			NodeIndex: 0,
		},
		Logging: helpers.LoggingConfig{
			Level: LOGLEVEL,
		},
		Db: mcConfig.DbConfig{
			InstanceMetricsDb: db.DatabaseConfig{
				Url: dbUri,
			},
			PolicyDb: db.DatabaseConfig{
				Url: dbUri,
			},
		},
		Collector: mcConfig.CollectorConfig{
			CollectInterval: collectInterval,
			RefreshInterval: refreshInterval,
			CollectMethod:   collectMethod,
			SaveInterval:    saveInterval,
		},
	}
	return writeYmlConfig(tmpDir, MetricsCollector, &cfg)
}

func (components *Components) PrepareEventGeneratorConfig(dbUri string, port int, metricsCollectorUrl string, scalingEngineUrl string, aggregatorExecuteInterval time.Duration,
	policyPollerInterval time.Duration, saveInterval time.Duration, evaluationManagerInterval time.Duration, tmpDir string) string {
	conf := &egConfig.Config{
		Logging: helpers.LoggingConfig{
			Level: LOGLEVEL,
		},
		Server: egConfig.ServerConfig{
			Port: port,
			TLS: models.TLSCerts{
				KeyFile:    filepath.Join(testCertDir, "eventgenerator.key"),
				CertFile:   filepath.Join(testCertDir, "eventgenerator.crt"),
				CACertFile: filepath.Join(testCertDir, "autoscaler-ca.crt"),
			},
			NodeAddrs: []string{"localhost"},
			NodeIndex: 0,
		},
		Aggregator: egConfig.AggregatorConfig{
			AggregatorExecuteInterval: aggregatorExecuteInterval,
			PolicyPollerInterval:      policyPollerInterval,
			SaveInterval:              saveInterval,
			MetricPollerCount:         1,
			AppMonitorChannelSize:     1,
			AppMetricChannelSize:      1,
		},
		Evaluator: egConfig.EvaluatorConfig{
			EvaluationManagerInterval: evaluationManagerInterval,
			EvaluatorCount:            1,
			TriggerArrayChannelSize:   1,
		},
		DB: egConfig.DBConfig{
			PolicyDB: db.DatabaseConfig{
				Url: dbUri,
			},
			AppMetricDB: db.DatabaseConfig{
				Url: dbUri,
			},
		},
		ScalingEngine: egConfig.ScalingEngineConfig{
			ScalingEngineUrl: scalingEngineUrl,
			TLSClientCerts: models.TLSCerts{
				KeyFile:    filepath.Join(testCertDir, "eventgenerator.key"),
				CertFile:   filepath.Join(testCertDir, "eventgenerator.crt"),
				CACertFile: filepath.Join(testCertDir, "autoscaler-ca.crt"),
			},
		},
		MetricCollector: egConfig.MetricCollectorConfig{
			MetricCollectorUrl: metricsCollectorUrl,
			TLSClientCerts: models.TLSCerts{
				KeyFile:    filepath.Join(testCertDir, "eventgenerator.key"),
				CertFile:   filepath.Join(testCertDir, "eventgenerator.crt"),
				CACertFile: filepath.Join(testCertDir, "autoscaler-ca.crt"),
			},
		},
		DefaultBreachDurationSecs: 600,
		DefaultStatWindowSecs:     60,
	}
	return writeYmlConfig(tmpDir, EventGenerator, &conf)
}

func (components *Components) PrepareScalingEngineConfig(dbUri string, port int, ccUAAUrl string, cfGrantTypePassword string, tmpDir string) string {
	conf := seConfig.Config{
		Cf: cf.CfConfig{
			Api:       ccUAAUrl,
			GrantType: cfGrantTypePassword,
			Username:  "admin",
			Password:  "admin",
		},
		Server: seConfig.ServerConfig{
			Port: port,
			TLS: models.TLSCerts{
				KeyFile:    filepath.Join(testCertDir, "scalingengine.key"),
				CertFile:   filepath.Join(testCertDir, "scalingengine.crt"),
				CACertFile: filepath.Join(testCertDir, "autoscaler-ca.crt"),
			},
		},
		Logging: helpers.LoggingConfig{
			Level: LOGLEVEL,
		},
		Db: seConfig.DbConfig{
			PolicyDb: db.DatabaseConfig{
				Url: dbUri,
			},
			ScalingEngineDb: db.DatabaseConfig{
				Url: dbUri,
			},
			SchedulerDb: db.DatabaseConfig{
				Url: dbUri,
			},
		},
		DefaultCoolDownSecs: 300,
		LockSize:            32,
	}

	return writeYmlConfig(tmpDir, ScalingEngine, &conf)
}

func (components *Components) PrepareOperatorConfig(dbUri string, scalingEngineUrl string, schedulerUrl string, syncInterval time.Duration, cutOffDays int, tmpDir string) string {
	conf := &opConfig.Config{
		Logging: helpers.LoggingConfig{
			Level: LOGLEVEL,
		},
		InstanceMetricsDb: opConfig.InstanceMetricsDbPrunerConfig{
			RefreshInterval: 2 * time.Minute,
			CutoffDays:      cutOffDays,
			Db: db.DatabaseConfig{
				Url: dbUri,
			},
		},
		AppMetricsDb: opConfig.AppMetricsDbPrunerConfig{
			RefreshInterval: 2 * time.Minute,
			CutoffDays:      cutOffDays,
			Db: db.DatabaseConfig{
				Url: dbUri,
			},
		},
		ScalingEngineDb: opConfig.ScalingEngineDbPrunerConfig{
			RefreshInterval: 2 * time.Minute,
			CutoffDays:      cutOffDays,
			Db: db.DatabaseConfig{
				Url: dbUri,
			},
		},
		ScalingEngine: opConfig.ScalingEngineConfig{
			Url:          scalingEngineUrl,
			SyncInterval: syncInterval,
			TLSClientCerts: models.TLSCerts{
				KeyFile:    filepath.Join(testCertDir, "scalingengine.key"),
				CertFile:   filepath.Join(testCertDir, "scalingengine.crt"),
				CACertFile: filepath.Join(testCertDir, "autoscaler-ca.crt"),
			},
		},
		Scheduler: opConfig.SchedulerConfig{
			Url:          schedulerUrl,
			SyncInterval: syncInterval,
			TLSClientCerts: models.TLSCerts{
				KeyFile:    filepath.Join(testCertDir, "scheduler.key"),
				CertFile:   filepath.Join(testCertDir, "scheduler.crt"),
				CACertFile: filepath.Join(testCertDir, "autoscaler-ca.crt"),
			},
		},
		Lock: opConfig.LockConfig{
			LockTTL:             15 * time.Second,
			LockRetryInterval:   5 * time.Second,
			ConsulClusterConfig: consulRunner.ConsulCluster(),
		},
		EnableDBLock: true,
		DBLock: opConfig.DBLockConfig{
			LockTTL: 30 * time.Second,
			LockDB: db.DatabaseConfig{
				Url: dbUri,
			},
			LockRetryInterval: 15 * time.Second,
		},
	}
	return writeYmlConfig(tmpDir, Operator, &conf)
}

func writeYmlConfig(dir string, componentName string, c interface{}) string {
	cfgFile, err := ioutil.TempFile(dir, componentName)
	Expect(err).NotTo(HaveOccurred())
	defer cfgFile.Close()
	configBytes, err := yaml.Marshal(c)
	ioutil.WriteFile(cfgFile.Name(), configBytes, 0777)
	return cfgFile.Name()

}

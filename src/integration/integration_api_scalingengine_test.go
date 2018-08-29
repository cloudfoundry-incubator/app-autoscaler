package integration

import (
	"autoscaler/cf"
	"autoscaler/models"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"strings"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/onsi/gomega/ghttp"
)

type ScalingHistoryResult struct {
	TotalResults int                        `json:"total_results"`
	TotalPages   int                        `json:"total_pages"`
	Page         int                        `json:"page"`
	PrevUrl      string                     `json:"prev_url"`
	NextUrl      string                     `json:"next_url"`
	Resources    []models.AppScalingHistory `json:"resources"`
}

var _ = Describe("Integration_Api_ScalingEngine", func() {
	var (
		initInstanceCount int = 2
		appId             string
		pathVariables     []string
		parameters        map[string]string
		history           *models.AppScalingHistory
	)

	BeforeEach(func() {
		initializeHttpClient("api.crt", "api.key", "autoscaler-ca.crt", apiScalingEngineHttpRequestTimeout)
		initializeHttpClientForPublicApi("api_public.crt", "api_public.key", "autoscaler-ca.crt", apiMetricsCollectorHttpRequestTimeout)
		startFakeCCNOAAUAA(initInstanceCount)
		scalingEngineConfPath = components.PrepareScalingEngineConfig(dbUrl, components.Ports[ScalingEngine], fakeCCNOAAUAA.URL(), cf.GrantTypePassword, tmpDir)
		startScalingEngine()

		apiServerConfPath = components.PrepareApiServerConfig(components.Ports[APIServer], components.Ports[APIPublicServer], false, 200, fakeCCNOAAUAA.URL(), dbUrl, fmt.Sprintf("https://127.0.0.1:%d", components.Ports[Scheduler]), fmt.Sprintf("https://127.0.0.1:%d", components.Ports[ScalingEngine]), fmt.Sprintf("https://127.0.0.1:%d", components.Ports[MetricsCollector]), fmt.Sprintf("https://127.0.0.1:%d", components.Ports[EventGenerator]), fmt.Sprintf("https://127.0.0.1:%d", components.Ports[ServiceBrokerInternal]), true, tmpDir)
		startApiServer()
		appId = getRandomId()
		pathVariables = []string{appId}

	})

	AfterEach(func() {
		stopApiServer()
		stopScalingEngine()
	})
	Describe("Get scaling histories", func() {

		Context("Cloud Controller api is not available", func() {
			BeforeEach(func() {
				fakeCCNOAAUAA.Reset()
				fakeCCNOAAUAA.AllowUnhandledRequests = true
				parameters = map[string]string{"start-time": "1111", "end-time": "9999", "order-direction": "desc", "page": "1", "results-per-page": "5"}
			})
			It("should error with status code 500", func() {
				By("check public api")
				checkPublicAPIResponseContentWithParameters(getScalingHistories, pathVariables, parameters, http.StatusInternalServerError, map[string]interface{}{})
			})
		})

		Context("UAA api is not available", func() {
			BeforeEach(func() {
				fakeCCNOAAUAA.Reset()
				fakeCCNOAAUAA.AllowUnhandledRequests = true
				fakeCCNOAAUAA.RouteToHandler("GET", "/v2/info", ghttp.RespondWithJSONEncoded(http.StatusOK,
					cf.Endpoints{
						AuthEndpoint:    fakeCCNOAAUAA.URL(),
						DopplerEndpoint: strings.Replace(fakeCCNOAAUAA.URL(), "http", "ws", 1),
					}))
				parameters = map[string]string{"start-time": "1111", "end-time": "9999", "order-direction": "desc", "page": "1", "results-per-page": "5"}
			})
			It("should error with status code 500", func() {
				By("check public api")
				checkPublicAPIResponseContentWithParameters(getScalingHistories, pathVariables, parameters, http.StatusInternalServerError, map[string]interface{}{})
			})
		})

		Context("UAA api returns 401", func() {
			BeforeEach(func() {
				fakeCCNOAAUAA.Reset()
				fakeCCNOAAUAA.AllowUnhandledRequests = true
				fakeCCNOAAUAA.RouteToHandler("GET", "/v2/info", ghttp.RespondWithJSONEncoded(http.StatusOK,
					cf.Endpoints{
						AuthEndpoint:    fakeCCNOAAUAA.URL(),
						DopplerEndpoint: strings.Replace(fakeCCNOAAUAA.URL(), "http", "ws", 1),
					}))
				fakeCCNOAAUAA.RouteToHandler("GET", "/userinfo", ghttp.RespondWithJSONEncoded(http.StatusUnauthorized, struct{}{}))
				parameters = map[string]string{"start-time": "1111", "end-time": "9999", "order-direction": "desc", "page": "1", "results-per-page": "5"}
			})
			It("should error with status code 401", func() {
				By("check public api")
				checkPublicAPIResponseContentWithParameters(getScalingHistories, pathVariables, parameters, http.StatusUnauthorized, map[string]interface{}{})
			})
		})

		Context("Check permission not passed", func() {
			BeforeEach(func() {
				fakeCCNOAAUAA.RouteToHandler("GET", checkUserSpaceRegPath, ghttp.RespondWithJSONEncoded(http.StatusOK,
					struct {
						TotalResults int `json:"total_results"`
					}{
						0,
					}))
				parameters = map[string]string{"start-time": "1111", "end-time": "9999", "order-direction": "desc", "page": "1", "results-per-page": "5"}
			})
			It("should error with status code 401", func() {
				By("check public api")
				checkPublicAPIResponseContentWithParameters(getScalingHistories, pathVariables, parameters, http.StatusUnauthorized, map[string]interface{}{})
			})
		})

		Context("ScalingEngine is down", func() {
			JustBeforeEach(func() {
				stopScalingEngine()
				parameters = map[string]string{"start-time": "1111", "end-time": "9999", "order-direction": "desc", "page": "1", "results-per-page": "5"}
			})

			It("should error with status code 500", func() {
				By("check public api")
				checkPublicAPIResponseContentWithParameters(getScalingHistories, pathVariables, parameters, http.StatusInternalServerError, map[string]interface{}{"error": fmt.Sprintf("connect ECONNREFUSED 127.0.0.1:%d", components.Ports[ScalingEngine])})

			})

		})

		Context("Get scaling histories", func() {
			BeforeEach(func() {
				history = &models.AppScalingHistory{
					AppId:        appId,
					OldInstances: 2,
					NewInstances: 4,
					Reason:       "a reason",
					Message:      "a message",
					ScalingType:  models.ScalingTypeDynamic,
					Status:       models.ScalingStatusSucceeded,
					Error:        "",
				}

				history.Timestamp = 666666
				insertScalingHistory(history)

				history.Timestamp = 222222
				insertScalingHistory(history)

				history.Timestamp = 555555
				insertScalingHistory(history)

				history.Timestamp = 333333
				insertScalingHistory(history)

				history.Timestamp = 444444
				insertScalingHistory(history)

				//add some other app id
				history.AppId = "some-other-app-id"
				history.Timestamp = 444444
				insertScalingHistory(history)

			})
			It("should get the scaling histories ", func() {
				By("get the 1st page")
				parameters = map[string]string{"start-time": "111111", "end-time": "999999", "order-direction": "desc", "page": "1", "results-per-page": "2"}
				result := ScalingHistoryResult{
					TotalResults: 5,
					TotalPages:   3,
					Page:         1,
					NextUrl:      getScalingHistoriesUrl(appId, parameters, 2),
					Resources: []models.AppScalingHistory{
						models.AppScalingHistory{
							AppId:        appId,
							Timestamp:    666666,
							ScalingType:  models.ScalingTypeDynamic,
							Status:       models.ScalingStatusSucceeded,
							OldInstances: 2,
							NewInstances: 4,
							Reason:       "a reason",
							Message:      "a message",
							Error:        "",
						},
						models.AppScalingHistory{
							AppId:        appId,
							Timestamp:    555555,
							ScalingType:  models.ScalingTypeDynamic,
							Status:       models.ScalingStatusSucceeded,
							OldInstances: 2,
							NewInstances: 4,
							Reason:       "a reason",
							Message:      "a message",
							Error:        "",
						},
					},
				}
				By("check public api")
				checkScalingHistoryResult(pathVariables, parameters, result)

				By("get the 2nd page")
				parameters = map[string]string{"start-time": "111111", "end-time": "999999", "order-direction": "desc", "page": "2", "results-per-page": "2"}
				result = ScalingHistoryResult{
					TotalResults: 5,
					TotalPages:   3,
					Page:         2,
					PrevUrl:      getScalingHistoriesUrl(appId, parameters, 1),
					NextUrl:      getScalingHistoriesUrl(appId, parameters, 3),
					Resources: []models.AppScalingHistory{
						models.AppScalingHistory{
							AppId:        appId,
							Timestamp:    444444,
							ScalingType:  models.ScalingTypeDynamic,
							Status:       models.ScalingStatusSucceeded,
							OldInstances: 2,
							NewInstances: 4,
							Reason:       "a reason",
							Message:      "a message",
							Error:        "",
						},
						models.AppScalingHistory{
							AppId:        appId,
							Timestamp:    333333,
							ScalingType:  models.ScalingTypeDynamic,
							Status:       models.ScalingStatusSucceeded,
							OldInstances: 2,
							NewInstances: 4,
							Reason:       "a reason",
							Message:      "a message",
							Error:        "",
						},
					},
				}
				By("check public api")
				checkScalingHistoryResult(pathVariables, parameters, result)

				By("get the 3rd page")
				parameters = map[string]string{"start-time": "111111", "end-time": "999999", "order-direction": "desc", "page": "3", "results-per-page": "2"}
				result = ScalingHistoryResult{
					TotalResults: 5,
					TotalPages:   3,
					Page:         3,
					PrevUrl:      getScalingHistoriesUrl(appId, parameters, 2),
					Resources: []models.AppScalingHistory{
						models.AppScalingHistory{
							AppId:        appId,
							Timestamp:    222222,
							ScalingType:  models.ScalingTypeDynamic,
							Status:       models.ScalingStatusSucceeded,
							OldInstances: 2,
							NewInstances: 4,
							Reason:       "a reason",
							Message:      "a message",
							Error:        "",
						},
					},
				}
				By("check public api")
				checkScalingHistoryResult(pathVariables, parameters, result)

				By("the 4th page should be empty")
				parameters = map[string]string{"start-time": "111111", "end-time": "999999", "order-direction": "desc", "page": "4", "results-per-page": "2"}
				result = ScalingHistoryResult{
					TotalResults: 5,
					TotalPages:   3,
					Page:         4,
					PrevUrl:      getScalingHistoriesUrl(appId, parameters, 3),
					Resources:    []models.AppScalingHistory{},
				}
				By("check public api")
				checkScalingHistoryResult(pathVariables, parameters, result)
			})

		})
	})
})

func getScalingHistoriesUrl(appId string, parameteters map[string]string, pageNo int) string {
	return fmt.Sprintf("/v1/apps/%s/scaling_histories?any=any&start-time=%s&end-time=%s&order-direction=%s&page=%d&results-per-page=%s", appId, parameteters["start-time"], parameteters["end-time"], parameteters["order-direction"], pageNo, parameteters["results-per-page"])
}

func compareScalingHistoryResult(o1, o2 ScalingHistoryResult) {
	Expect(o1.Page).To(Equal(o2.Page))
	Expect(o1.TotalPages).To(Equal(o2.TotalPages))
	Expect(o1.TotalResults).To(Equal(o2.TotalResults))
	Expect(o1.Resources).To(Equal(o2.Resources))

	prevUrl1, err1 := url.Parse(o1.PrevUrl)
	Expect(err1).NotTo(HaveOccurred())
	prevUrl2, err2 := url.Parse(o2.PrevUrl)
	Expect(err2).NotTo(HaveOccurred())
	queries1 := prevUrl1.Query()
	queries2 := prevUrl2.Query()
	Expect(queries1).To(Equal(queries2))

	nextUrl1, err1 := url.Parse(o1.NextUrl)
	Expect(err1).NotTo(HaveOccurred())
	nextUrl2, err2 := url.Parse(o2.NextUrl)
	Expect(err2).NotTo(HaveOccurred())
	queries1 = nextUrl1.Query()
	queries2 = nextUrl2.Query()
	Expect(queries1).To(Equal(queries2))

}
func checkScalingHistoryResult(pathVariables []string, parameters map[string]string, result ScalingHistoryResult) {
	var actual ScalingHistoryResult
	resp, err := getScalingHistories(pathVariables, parameters)
	defer resp.Body.Close()
	Expect(err).NotTo(HaveOccurred())
	Expect(resp.StatusCode).To(Equal(http.StatusOK))
	err = json.NewDecoder(resp.Body).Decode(&actual)
	Expect(err).NotTo(HaveOccurred())
	compareScalingHistoryResult(actual, result)

}

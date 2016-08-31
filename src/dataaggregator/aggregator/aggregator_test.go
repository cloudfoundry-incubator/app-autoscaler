package aggregator_test

import (
	"code.cloudfoundry.org/clock/fakeclock"
	"code.cloudfoundry.org/lager"
	. "dataaggregator/aggregator"
	"dataaggregator/aggregator/fakes"
	. "dataaggregator/appmetric"
	. "dataaggregator/policy"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/onsi/gomega/ghttp"
	. "metricscollector/metrics"
	"net/http"
	"regexp"
	"time"
)

var _ = Describe("Aggregator", func() {
	var (
		aggregator            *Aggregator
		database              *fakes.FakeDB
		clock                 *fakeclock.FakeClock
		logger                lager.Logger
		metricServer          *ghttp.Server
		TestMetricPollerCount int    = 3
		testAppId             string = "testAppId"
		testAppId2            string = "testAppId2"
		testAppId3            string = "testAppId3"
		testAppId4            string = "testAppId4"
		timestamp             int64  = time.Now().UnixNano()
		metricType            string = "MemoryUsage"
		unit                  string = "bytes"
		policyStr                    = `
		{
		   "instance_min_count":1,
		   "instance_max_count":5,
		   "scaling_rules":[
		      {
		         "metric_type":"MemoryUsage",
		         "stat_window":300,
		         "breach_duration":300,
		         "threshold":30,
		         "operator":"<",
		         "cool_down_duration":300,
		         "adjustment":"-1"
		      }
		   ]
		}`
		metrics []*Metric = []*Metric{
			&Metric{
				Name:      metricType,
				Unit:      unit,
				AppId:     testAppId,
				TimeStamp: timestamp,
				Instances: []InstanceMetric{InstanceMetric{
					Timestamp: timestamp,
					Index:     0,
					Value:     "100",
				}, InstanceMetric{
					Timestamp: timestamp,
					Index:     1,
					Value:     "200",
				}},
			},
			&Metric{
				Name:      metricType,
				Unit:      unit,
				AppId:     testAppId,
				TimeStamp: timestamp,
				Instances: []InstanceMetric{InstanceMetric{
					Timestamp: timestamp,
					Index:     0,
					Value:     "300",
				}, InstanceMetric{
					Timestamp: timestamp,
					Index:     1,
					Value:     "400",
				}},
			},
		}
	)
	BeforeEach(func() {
		database = &fakes.FakeDB{}
		database.RetrievePoliciesStub = func() ([]*PolicyJson, error) {
			return []*PolicyJson{&PolicyJson{AppId: testAppId, PolicyStr: policyStr}}, nil
		}
		database.SaveAppMetricStub = func(appMetric *AppMetric) error {
			Expect(appMetric.AppId).To(Equal("testAppId"))
			Expect(appMetric.MetricType).To(Equal(metricType))
			Expect(appMetric.Unit).To(Equal(unit))
			Expect(appMetric.Value).To(Equal(int64(250)))
			return nil
		}
		clock = fakeclock.NewFakeClock(time.Now())
		logger = lager.NewLogger("Aggregator-test")
		metricServer = ghttp.NewServer()
		regPath := regexp.MustCompile(`^/v1/apps/.*/metrics_history/memory$`)
		metricServer.RouteToHandler("GET", regPath, ghttp.RespondWithJSONEncoded(http.StatusOK,
			&metrics))
	})
	Context("ConsumeTrigger", func() {
		var triggerMap map[string]*Trigger
		var appChan chan *AppMonitor
		var appMonitor *AppMonitor
		BeforeEach(func() {
			appChan = make(chan *AppMonitor, 1)
			aggregator = NewAggregator(logger, clock, TestPolicyPollerInterval, database, metricServer.URL(), TestMetricPollerCount)
			triggerMap = map[string]*Trigger{testAppId: &Trigger{
				AppId: testAppId,
				TriggerRecord: &TriggerRecord{
					InstanceMaxCount: 5,
					InstanceMinCount: 1,
					ScalingRules: []*ScalingRule{&ScalingRule{
						MetricType:       "MemoryUsage",
						StatWindow:       300,
						BreachDuration:   300,
						CoolDownDuration: 300,
						Threshold:        30,
						Operator:         "<",
						Adjustment:       "-1",
					}}},
			}}
		})
		Context("when there are data in triggerMap", func() {
			JustBeforeEach(func() {
				aggregator.ConsumeTrigger(triggerMap, appChan)
			})
			It("should parse the triggers to appmonitor and put them in appChan", func() {
				Eventually(appChan).Should(Receive(&appMonitor))
				Expect(appMonitor).To(Equal(&AppMonitor{
					AppId:      testAppId,
					MetricType: "MemoryUsage",
					StatWindow: 300,
				}))
			})
		})
		Context("when there is not data in triggerMap", func() {
			BeforeEach(func() {
				triggerMap = map[string]*Trigger{}
			})
			JustBeforeEach(func() {
				aggregator.ConsumeTrigger(triggerMap, appChan)
			})
			It("should not receive any data from the appChan", func() {
				Consistently(appChan).ShouldNot(Receive())
			})
		})
		Context("when the triggerMap is nil", func() {
			BeforeEach(func() {
				triggerMap = nil
			})
			JustBeforeEach(func() {
				aggregator.ConsumeTrigger(triggerMap, appChan)
			})
			It("should not receive any data from the appChan", func() {
				Consistently(appChan).ShouldNot(Receive())
			})
		})

	})
	Context("ConsumeAppMetric", func() {
		var appmetric *AppMetric
		BeforeEach(func() {
			aggregator = NewAggregator(logger, clock, TestPolicyPollerInterval, database, metricServer.URL(), TestMetricPollerCount)
			appmetric = &AppMetric{
				AppId:      testAppId,
				MetricType: metricType,
				Value:      250,
				Unit:       "bytes",
				Timestamp:  timestamp}
		})
		Context("when there is data in appmetric", func() {
			JustBeforeEach(func() {
				aggregator.ConsumeAppMetric(appmetric)
			})
			It("should call database.SaveAppmetric to save the appmetric to database", func() {
				Eventually(database.SaveAppMetricCallCount).Should(Equal(1))
			})
		})
		Context("when the appmetric is nil", func() {
			BeforeEach(func() {
				appmetric = nil
			})
			JustBeforeEach(func() {
				aggregator.ConsumeAppMetric(appmetric)
			})
			It("should call database.SaveAppmetric to save the appmetric to database", func() {
				Consistently(database.SaveAppMetricCallCount).Should(Equal(0))
			})
		})

	})
	Context("Start", func() {
		JustBeforeEach(func() {
			aggregator = NewAggregator(logger, clock, TestPolicyPollerInterval, database, metricServer.URL(), TestMetricPollerCount)
			aggregator.Start()
		})
		AfterEach(func() {
			aggregator.Stop()
		})
		It("should save the appmetric to database", func() {
			clock.Increment(2 * TestPolicyPollerInterval * time.Second)
			Eventually(database.RetrievePoliciesCallCount).Should(BeNumerically(">=", 2))
			Eventually(database.SaveAppMetricCallCount).Should(BeNumerically(">=", 2))

		})
		Context("MetricPoller", func() {
			var unBlockChan chan bool
			var calledChan chan string
			BeforeEach(func() {
				TestMetricPollerCount = 3
				unBlockChan = make(chan bool)
				calledChan = make(chan string)
				database.RetrievePoliciesStub = func() ([]*PolicyJson, error) {
					return []*PolicyJson{&PolicyJson{AppId: testAppId, PolicyStr: policyStr}, &PolicyJson{AppId: testAppId2, PolicyStr: policyStr}, &PolicyJson{AppId: testAppId3, PolicyStr: policyStr}, &PolicyJson{AppId: testAppId4, PolicyStr: policyStr}}, nil
				}
				database.SaveAppMetricStub = func(appMetric *AppMetric) error {
					calledChan <- appMetric.AppId
					<-unBlockChan
					return nil
				}
			})
			It("should create MetricPollerCount metric-pollers", func() {
				for i := 0; i < TestMetricPollerCount; i++ {
					Eventually(calledChan).Should(Receive())
				}
				Consistently(calledChan).ShouldNot(Receive())
				Eventually(database.SaveAppMetricCallCount).Should(Equal(int(TestMetricPollerCount)))
				for i := 0; i < TestMetricPollerCount; i++ {
					unBlockChan <- true
				}
				<-calledChan
				unBlockChan <- true
			})
		})
	})
	Context("Stop", func() {
		var retrievePoliciesCallCount, saveAppMetricCallCount int
		JustBeforeEach(func() {
			aggregator = NewAggregator(logger, clock, TestPolicyPollerInterval, database, metricServer.URL(), TestMetricPollerCount)
			aggregator.Start()
			aggregator.Stop()
			retrievePoliciesCallCount = database.RetrievePoliciesCallCount()
			saveAppMetricCallCount = database.SaveAppMetricCallCount()

		})
		It("should return 1", func() {
			clock.Increment(10 * TestPolicyPollerInterval * time.Second)
			Eventually(database.RetrievePoliciesCallCount).Should(Equal(retrievePoliciesCallCount))
			Eventually(database.SaveAppMetricCallCount).Should(Equal(saveAppMetricCallCount))

		})
	})
})

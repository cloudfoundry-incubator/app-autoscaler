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
	"time"
)

var _ = Describe("Aggregator", func() {
	var (
		aggregator   *Aggregator
		database     *fakes.FakeDB
		clock        *fakeclock.FakeClock
		logger       lager.Logger
		metricServer *ghttp.Server
		testAppId    string = "testAppId"
		timestamp    int64  = time.Now().UnixNano()
		metricType   string = "MemoryUsage"
		unit         string = "bytes"
		policyStr           = `
		{
		   "instance_min_count":1,
		   "instance_max_count":5,
		   "scaling_rules":[
		      {
		         "metric_type":"MemoryUsage",
		         "stat_window_secs":300,
		         "breach_duration_secs":300,
		         "threshold":30,
		         "operator":"<",
		         "cool_down_secs":300,
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
			Expect(appMetric.Value).To(BeNumerically("==", 250))
			return nil
		}
		clock = fakeclock.NewFakeClock(time.Now())
		logger = lager.NewLogger("Aggregator-test")
		metricServer = ghttp.NewServer()
		metricServer.RouteToHandler("GET", "/v1/apps/"+testAppId+"/metrics_history/memory", ghttp.RespondWithJSONEncoded(http.StatusOK,
			&metrics))
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
			Eventually(database.RetrievePoliciesCallCount).Should(BeNumerically("==", retrievePoliciesCallCount))
			Eventually(database.SaveAppMetricCallCount).Should(BeNumerically("==", saveAppMetricCallCount))

		})
	})
})

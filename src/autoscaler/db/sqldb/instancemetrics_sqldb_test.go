package sqldb_test

import (
	"autoscaler/db"
	. "autoscaler/db/sqldb"
	"autoscaler/models"

	"code.cloudfoundry.org/lager"
	"github.com/lib/pq"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/onsi/gomega/gstruct"

	"os"
	"time"
)

var _ = Describe("InstancemetricsSqldb", func() {
	var (
		dbConfig  db.DatabaseConfig
		idb       *InstanceMetricsSQLDB
		logger    lager.Logger
		err       error
		metric    *models.AppInstanceMetric
		mtrcs     []*models.AppInstanceMetric
		start     int64
		end       int64
		before    int64
		orderType db.OrderType

		appId          string
		instanceIndex  int64
		metricName     string
		testAppId      string = "Test-App-ID"
		testMetricName string = "TestMetricType"
		testMetricUnit string = "TestMetricUnit"
	)

	BeforeEach(func() {
		logger = lager.NewLogger("instance-metrics-sqldb-test")
		dbConfig = db.DatabaseConfig{
			Url:                   os.Getenv("DBURL"),
			MaxOpenConnections:    10,
			MaxIdleConnections:    5,
			ConnectionMaxLifetime: 10 * time.Second,
		}
		instanceIndex = -1
	})

	Describe("NewInstanceMetricsSQLDB", func() {
		JustBeforeEach(func() {
			idb, err = NewInstanceMetricsSQLDB(dbConfig, logger)
		})

		AfterEach(func() {
			if idb != nil {
				err = idb.Close()
				Expect(err).NotTo(HaveOccurred())
			}
		})

		Context("when db url is not correct", func() {
			BeforeEach(func() {
				dbConfig.Url = "postgres://not-exist-user:not-exist-password@localhost/autoscaler?sslmode=disable"
			})
			It("should error", func() {
				Expect(err).To(BeAssignableToTypeOf(&pq.Error{}))
			})

		})

		Context("when url is correct", func() {
			It("should not error", func() {
				Expect(err).NotTo(HaveOccurred())
				Expect(idb).NotTo(BeNil())
			})
		})
	})

	Describe("SaveMetric", func() {
		BeforeEach(func() {
			idb, err = NewInstanceMetricsSQLDB(dbConfig, logger)
			Expect(err).NotTo(HaveOccurred())
			cleanInstanceMetricsTable()
		})

		AfterEach(func() {
			err = idb.Close()
			Expect(err).NotTo(HaveOccurred())
		})

		Context("When inserting a metric", func() {
			BeforeEach(func() {
				metric = &models.AppInstanceMetric{
					AppId:         testAppId,
					InstanceIndex: 0,
					CollectedAt:   111111,
					Name:          testMetricName,
					Unit:          testMetricUnit,
					Value:         "123",
					Timestamp:     110000,
				}
				err = idb.SaveMetric(metric)
			})

			It("has the metric in database", func() {
				Expect(err).NotTo(HaveOccurred())
				Expect(hasInstanceMetric(testAppId, 0, testMetricName, 110000)).To(BeTrue())
			})
		})

		Context("When inserting multiple metrics", func() {
			BeforeEach(func() {
				metric = &models.AppInstanceMetric{
					AppId: testAppId,
					Name:  testMetricName,
					Unit:  testMetricUnit,
				}
				metric.InstanceIndex = 0
				metric.CollectedAt = 111111
				metric.Value = "123"
				metric.Timestamp = 111100
				err = idb.SaveMetric(metric)
				Expect(err).NotTo(HaveOccurred())

				metric.InstanceIndex = 1
				metric.CollectedAt = 111111
				metric.Value = "214365"
				metric.Timestamp = 110000
				err = idb.SaveMetric(metric)
				Expect(err).NotTo(HaveOccurred())

				metric.InstanceIndex = 0
				metric.CollectedAt = 222222
				metric.Value = "654321"
				metric.Timestamp = 220000
				err = idb.SaveMetric(metric)
				Expect(err).NotTo(HaveOccurred())
			})

			It("has all the metrics in database", func() {
				Expect(hasInstanceMetric(testAppId, 0, testMetricName, 111100)).To(BeTrue())
				Expect(hasInstanceMetric(testAppId, 1, testMetricName, 110000)).To(BeTrue())
				Expect(hasInstanceMetric(testAppId, 0, testMetricName, 220000)).To(BeTrue())
			})
		})
	})

	Describe("SaveMetricsInBulk", func() {
		BeforeEach(func() {
			idb, err = NewInstanceMetricsSQLDB(dbConfig, logger)
			Expect(err).NotTo(HaveOccurred())
			cleanInstanceMetricsTable()
		})

		AfterEach(func() {
			err = idb.Close()
			Expect(err).NotTo(HaveOccurred())
		})

		Context("When inserting an array of metrics", func() {
			BeforeEach(func() {
				metric1 := models.AppInstanceMetric{
					AppId:         testAppId,
					InstanceIndex: 0,
					CollectedAt:   111111,
					Name:          testMetricName,
					Unit:          testMetricUnit,
					Value:         "123",
					Timestamp:     110000,
				}
				metric2 := models.AppInstanceMetric{
					AppId:         testAppId,
					InstanceIndex: 1,
					CollectedAt:   222222,
					Name:          testMetricName,
					Unit:          testMetricUnit,
					Value:         "234",
					Timestamp:     220000,
				}
				err = idb.SaveMetricsInBulk([]*models.AppInstanceMetric{&metric1, &metric2})
			})

			It("has the metrics in database", func() {
				Expect(err).NotTo(HaveOccurred())
				Expect(hasInstanceMetric(testAppId, 0, testMetricName, 110000)).To(BeTrue())
				Expect(hasInstanceMetric(testAppId, 1, testMetricName, 220000)).To(BeTrue())
			})
		})
	})

	Describe("RetrieveInstanceMetrics", func() {
		BeforeEach(func() {
			idb, err = NewInstanceMetricsSQLDB(dbConfig, logger)
			Expect(err).NotTo(HaveOccurred())
			cleanInstanceMetricsTable()

			metric = &models.AppInstanceMetric{
				AppId: testAppId,
				Name:  testMetricName,
				Unit:  testMetricUnit,
			}

			metric.InstanceIndex = 0
			metric.CollectedAt = 111111
			metric.Value = "654321"
			metric.Timestamp = 111100
			err = idb.SaveMetric(metric)
			Expect(err).NotTo(HaveOccurred())

			metric.InstanceIndex = 1
			metric.CollectedAt = 111111
			metric.Value = "214365"
			metric.Timestamp = 110000
			err = idb.SaveMetric(metric)
			Expect(err).NotTo(HaveOccurred())

			metric.InstanceIndex = 1
			metric.CollectedAt = 222222
			metric.Value = "321765"
			metric.Timestamp = 222200
			err = idb.SaveMetric(metric)
			Expect(err).NotTo(HaveOccurred())

			metric.InstanceIndex = 0
			metric.CollectedAt = 222222
			metric.Value = "654321"
			metric.Timestamp = 111100
			err = idb.SaveMetric(metric)
			Expect(err).NotTo(HaveOccurred())

			start = 0
			end = -1
			appId = testAppId
			metricName = testMetricName
			orderType = db.DESC

		})

		AfterEach(func() {
			err = idb.Close()
			Expect(err).NotTo(HaveOccurred())
		})

		JustBeforeEach(func() {
			mtrcs, err = idb.RetrieveInstanceMetrics(appId, instanceIndex, metricName, start, end, orderType)
		})

		Context("The app has no instance metrics", func() {
			BeforeEach(func() {
				appId = "app-id-no-metric"
			})

			It("returns empty metrics", func() {
				Expect(err).NotTo(HaveOccurred())
				Expect(mtrcs).To(BeEmpty())
			})

		})

		Context("The app has no instance metrics for a given instance index", func() {
			BeforeEach(func() {
				instanceIndex = 2
			})

			It("returns empty metrics", func() {
				Expect(err).NotTo(HaveOccurred())
				Expect(mtrcs).To(BeEmpty())
			})

		})

		Context("when the app has no instance metrics with the given metric name", func() {
			BeforeEach(func() {
				metricName = "metric-name-no-metrics"
			})

			It("returns empty metrics", func() {
				Expect(err).NotTo(HaveOccurred())
				Expect(mtrcs).To(BeEmpty())
			})
		})

		Context("when retrieving all the metrics of all the instances( start = 0, end = -1, orderType = ASC) ", func() {
			BeforeEach(func() {
				orderType = db.ASC
				instanceIndex = -1
			})
			It("removes duplicates and returns all the instance metrics of the app ordered by timestamp asc", func() {
				Expect(err).NotTo(HaveOccurred())
				Expect(mtrcs).To(HaveLen(3))
				Expect(*mtrcs[0]).To(gstruct.MatchAllFields(gstruct.Fields{
					"AppId":         Equal(testAppId),
					"InstanceIndex": BeEquivalentTo(1),
					"CollectedAt":   BeEquivalentTo(111111),
					"Name":          Equal(testMetricName),
					"Unit":          Equal(testMetricUnit),
					"Value":         Equal("214365"),
					"Timestamp":     BeEquivalentTo(110000),
				}))

				Expect(*mtrcs[1]).To(gstruct.MatchAllFields(gstruct.Fields{
					"AppId":         Equal(testAppId),
					"InstanceIndex": BeEquivalentTo(0),
					"CollectedAt":   BeNumerically(">=", 111111),
					"Name":          Equal(testMetricName),
					"Unit":          Equal(testMetricUnit),
					"Value":         Equal("654321"),
					"Timestamp":     BeEquivalentTo(111100),
				}))

				Expect(*mtrcs[2]).To(gstruct.MatchAllFields(gstruct.Fields{
					"AppId":         Equal(testAppId),
					"InstanceIndex": BeEquivalentTo(1),
					"CollectedAt":   BeEquivalentTo(222222),
					"Name":          Equal(testMetricName),
					"Unit":          Equal(testMetricUnit),
					"Value":         Equal("321765"),
					"Timestamp":     BeEquivalentTo(222200),
				}))
			})
		})

		Context("when retrieving all the metrics of all the instances( start = 0, end = -1, orderType = DESC) ", func() {
			BeforeEach(func() {
				orderType = db.DESC
				instanceIndex = -1
			})
			It("removes duplicates and returns all the instance metrics of the app ordered by timestamp desc", func() {
				Expect(err).NotTo(HaveOccurred())
				Expect(mtrcs).To(HaveLen(3))
				Expect(*mtrcs[2]).To(gstruct.MatchAllFields(gstruct.Fields{
					"AppId":         Equal(testAppId),
					"InstanceIndex": BeEquivalentTo(1),
					"CollectedAt":   BeEquivalentTo(111111),
					"Name":          Equal(testMetricName),
					"Unit":          Equal(testMetricUnit),
					"Value":         Equal("214365"),
					"Timestamp":     BeEquivalentTo(110000),
				}))

				Expect(*mtrcs[1]).To(gstruct.MatchAllFields(gstruct.Fields{
					"AppId":         Equal(testAppId),
					"InstanceIndex": BeEquivalentTo(0),
					"CollectedAt":   BeNumerically(">=", 111111),
					"Name":          Equal(testMetricName),
					"Unit":          Equal(testMetricUnit),
					"Value":         Equal("654321"),
					"Timestamp":     BeEquivalentTo(111100),
				}))

				Expect(*mtrcs[0]).To(gstruct.MatchAllFields(gstruct.Fields{
					"AppId":         Equal(testAppId),
					"InstanceIndex": BeEquivalentTo(1),
					"CollectedAt":   BeEquivalentTo(222222),
					"Name":          Equal(testMetricName),
					"Unit":          Equal(testMetricUnit),
					"Value":         Equal("321765"),
					"Timestamp":     BeEquivalentTo(222200),
				}))
			})
		})

		Context("when retrieving all the metrics of a given instance index( start = 0, end = -1, orderType = ASC) ", func() {
			BeforeEach(func() {
				orderType = db.ASC
				instanceIndex = 1
			})
			It("removes duplicates and returns all the instance metrics of the app ordered by timestamp asc", func() {
				Expect(err).NotTo(HaveOccurred())
				Expect(mtrcs).To(HaveLen(2))
				Expect(*mtrcs[0]).To(gstruct.MatchAllFields(gstruct.Fields{
					"AppId":         Equal(testAppId),
					"InstanceIndex": BeEquivalentTo(1),
					"CollectedAt":   BeEquivalentTo(111111),
					"Name":          Equal(testMetricName),
					"Unit":          Equal(testMetricUnit),
					"Value":         Equal("214365"),
					"Timestamp":     BeEquivalentTo(110000),
				}))

				Expect(*mtrcs[1]).To(gstruct.MatchAllFields(gstruct.Fields{
					"AppId":         Equal(testAppId),
					"InstanceIndex": BeEquivalentTo(1),
					"CollectedAt":   BeEquivalentTo(222222),
					"Name":          Equal(testMetricName),
					"Unit":          Equal(testMetricUnit),
					"Value":         Equal("321765"),
					"Timestamp":     BeEquivalentTo(222200),
				}))
			})
		})

		Context("when retrieving all the metrics of a given instance index( start = 0, end = -1, orderType = DESC) ", func() {
			BeforeEach(func() {
				instanceIndex = 1
			})
			It("removes duplicates and returns all the instance metrics of the app ordered by timestamp desc", func() {
				Expect(err).NotTo(HaveOccurred())
				Expect(mtrcs).To(HaveLen(2))
				Expect(*mtrcs[1]).To(gstruct.MatchAllFields(gstruct.Fields{
					"AppId":         Equal(testAppId),
					"InstanceIndex": BeEquivalentTo(1),
					"CollectedAt":   BeEquivalentTo(111111),
					"Name":          Equal(testMetricName),
					"Unit":          Equal(testMetricUnit),
					"Value":         Equal("214365"),
					"Timestamp":     BeEquivalentTo(110000),
				}))

				Expect(*mtrcs[0]).To(gstruct.MatchAllFields(gstruct.Fields{
					"AppId":         Equal(testAppId),
					"InstanceIndex": BeEquivalentTo(1),
					"CollectedAt":   BeEquivalentTo(222222),
					"Name":          Equal(testMetricName),
					"Unit":          Equal(testMetricUnit),
					"Value":         Equal("321765"),
					"Timestamp":     BeEquivalentTo(222200),
				}))
			})
		})

		Context("when end time = -1)", func() {
			BeforeEach(func() {
				start = 111111
				end = -1
			})

			It("returns metrics from start time to now", func() {
				Expect(err).NotTo(HaveOccurred())
				Expect(mtrcs).To(HaveLen(1))
			})

		})

		Context("when end time is before all the metrics timestamps", func() {
			BeforeEach(func() {
				start = 22222
				end = 100000
			})

			It("returns empty metrics", func() {
				Expect(err).NotTo(HaveOccurred())
				Expect(mtrcs).To(BeEmpty())
			})

		})

		Context("when start time is after all the metrics timestamps", func() {
			BeforeEach(func() {
				start = 222222
				end = 888888
			})

			It("returns empty metrics", func() {
				Expect(err).NotTo(HaveOccurred())
				Expect(mtrcs).To(BeEmpty())
			})

		})

		Context("when start time > end time", func() {
			BeforeEach(func() {
				start = 200000
				end = 111111
			})

			It("returns empty metrics", func() {
				Expect(err).NotTo(HaveOccurred())
				Expect(mtrcs).To(BeEmpty())
			})
		})

		Context("when retrieving part of all the instances's metrics", func() {
			BeforeEach(func() {
				start = 111100
				end = 222222
			})

			It("returns correct metrics", func() {
				Expect(err).NotTo(HaveOccurred())
				Expect(mtrcs).To(HaveLen(2))
				Expect(*mtrcs[1]).To(gstruct.MatchAllFields(gstruct.Fields{
					"AppId":         Equal(testAppId),
					"InstanceIndex": BeEquivalentTo(0),
					"CollectedAt":   BeNumerically(">=", 111111),
					"Name":          Equal(testMetricName),
					"Unit":          Equal(testMetricUnit),
					"Value":         Equal("654321"),
					"Timestamp":     BeEquivalentTo(111100),
				}))
				Expect(*mtrcs[0]).To(gstruct.MatchAllFields(gstruct.Fields{
					"AppId":         Equal(testAppId),
					"InstanceIndex": BeEquivalentTo(1),
					"CollectedAt":   BeEquivalentTo(222222),
					"Name":          Equal(testMetricName),
					"Unit":          Equal(testMetricUnit),
					"Value":         Equal("321765"),
					"Timestamp":     BeEquivalentTo(222200),
				}))
			})
		})

		Context("when retrieving part of a given instance's metrics", func() {
			BeforeEach(func() {
				start = 111100
				end = 222222
				instanceIndex = 1
			})

			It("returns correct metrics", func() {
				Expect(err).NotTo(HaveOccurred())
				Expect(mtrcs).To(HaveLen(1))
				Expect(*mtrcs[0]).To(gstruct.MatchAllFields(gstruct.Fields{
					"AppId":         Equal(testAppId),
					"InstanceIndex": BeEquivalentTo(1),
					"CollectedAt":   BeEquivalentTo(222222),
					"Name":          Equal(testMetricName),
					"Unit":          Equal(testMetricUnit),
					"Value":         Equal("321765"),
					"Timestamp":     BeEquivalentTo(222200),
				}))
			})
		})

		Context("when db fails", func() {
			BeforeEach(func() {
				idb.Close()
			})

			It("should error", func() {
				Expect(err).To(MatchError(MatchRegexp("sql: .*")))
			})

		})
	})

	Describe("PruneMetrics", func() {
		BeforeEach(func() {
			idb, err = NewInstanceMetricsSQLDB(dbConfig, logger)
			Expect(err).NotTo(HaveOccurred())

			cleanInstanceMetricsTable()

			metric = &models.AppInstanceMetric{
				AppId: testAppId,
				Name:  testMetricName,
				Unit:  testMetricUnit,
			}

			metric.InstanceIndex = 0
			metric.CollectedAt = 111111
			metric.Value = "654321"
			metric.Timestamp = 111100
			err = idb.SaveMetric(metric)
			Expect(err).NotTo(HaveOccurred())

			metric.InstanceIndex = 1
			metric.CollectedAt = 111111
			metric.Value = "214365"
			metric.Timestamp = 110000
			err = idb.SaveMetric(metric)
			Expect(err).NotTo(HaveOccurred())

			metric.InstanceIndex = 1
			metric.CollectedAt = 222222
			metric.Value = "321765"
			metric.Timestamp = 222200
			err = idb.SaveMetric(metric)
			Expect(err).NotTo(HaveOccurred())

			metric.InstanceIndex = 0
			metric.CollectedAt = 222222
			metric.Value = "654321"
			metric.Timestamp = 111100
			err = idb.SaveMetric(metric)
			Expect(err).NotTo(HaveOccurred())
		})

		AfterEach(func() {
			err = idb.Close()
			Expect(err).NotTo(HaveOccurred())
		})

		JustBeforeEach(func() {
			err = idb.PruneInstanceMetrics(before)
		})

		Context("when pruning metrics before all the timestamps of metrics", func() {
			BeforeEach(func() {
				before = 1000
			})

			It("does not remove any metrics", func() {
				Expect(err).NotTo(HaveOccurred())
				Expect(getNumberOfInstanceMetrics()).To(Equal(4))
			})
		})

		Context("when pruning all the metrics", func() {
			BeforeEach(func() {
				before = time.Now().UnixNano()
			})

			It("empties the metrics table", func() {
				Expect(err).NotTo(HaveOccurred())
				Expect(getNumberOfInstanceMetrics()).To(Equal(0))
			})
		})

		Context("when pruning part of the metrics", func() {
			BeforeEach(func() {
				before = 111000
			})

			It("removes metrics before the time specified", func() {
				Expect(err).NotTo(HaveOccurred())
				Expect(getNumberOfInstanceMetrics()).To(Equal(3))
			})
		})

		Context("when db fails", func() {
			BeforeEach(func() {
				idb.Close()
			})

			It("should error", func() {
				Expect(err).To(MatchError(MatchRegexp("sql: .*")))
			})

		})

	})

})

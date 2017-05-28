package noaa_test

import (
	. "autoscaler/metricscollector/noaa"
	"autoscaler/models"

	"github.com/cloudfoundry/sonde-go/events"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("AppEvents", func() {
	Describe("GetInstanceMemoryMetricFromContainerMetricsEvent", func() {
		var (
			event  *events.Envelope
			metric *models.AppInstanceMetric
		)
		JustBeforeEach(func() {
			metric = GetInstanceMemoryMetricFromContainerMetricEvent(123456, "an-app-id", event)
		})

		Context("when it is a containermetric event", func() {
			BeforeEach(func() {
				event = NewContainerEnvelope(111111, "an-app-id", 0, 12.11, 1000, 1000, 3000, 2000)
			})
			It("returns the memory metric", func() {
				Expect(metric).To(Equal(&models.AppInstanceMetric{
					AppId:         "an-app-id",
					InstanceIndex: 0,
					CollectedAt:   123456,
					Name:          models.MetricNameMemory,
					Unit:          models.UnitPercentage,
					Value:         "33",
					Timestamp:     111111,
				}))
			})
		})

		Context("when there is no containermetric in the event", func() {
			BeforeEach(func() {
				event = &events.Envelope{
					ContainerMetric: nil,
				}
			})
			It("returns nil", func() {
				Expect(metric).To(BeNil())
			})
		})

		Context("when it is a containermetric event of other app", func() {
			BeforeEach(func() {
				event = NewContainerEnvelope(111111, "different-app-id", 0, 12.11, 1000, 1000, 3000, 2000)
			})
			It("returns nil", func() {
				Expect(metric).To(BeNil())
			})
		})

	})
	Describe("GetInstanceMemoryMetricFromContainerEnvelopes", func() {
		var (
			containerEnvelops []*events.Envelope
			metrics           []*models.AppInstanceMetric
		)

		JustBeforeEach(func() {
			metrics = GetInstanceMemoryMetricFromContainerEnvelopes(123456, "an-app-id", containerEnvelops)
		})

		Context("when metrics are empty", func() {
			BeforeEach(func() {
				containerEnvelops = []*events.Envelope{}
			})

			It("should return empty instance memory metrics", func() {
				Expect(metrics).To(BeEmpty())
			})
		})

		Context("when no metric is available for the given app", func() {
			BeforeEach(func() {
				containerEnvelops = []*events.Envelope{
					NewContainerEnvelope(111111, "different-app-id", 0, 12.11, 1000, 1000, 3000, 2000),
					NewContainerEnvelope(222222, "different-app-id", 1, 0.211, 2000, 1000, 3000, 2000),
					NewContainerEnvelope(333333, "another-different-app-id", 0, 0.211, 1000, 1000, 3000, 2000),
				}
			})

			It("should return empty instance memory metrics", func() {
				Expect(metrics).To(BeEmpty())
			})
		})

		Context("when metrics from both given app and other apps", func() {
			BeforeEach(func() {
				containerEnvelops = []*events.Envelope{
					NewContainerEnvelope(111111, "an-app-id", 0, 12.11, 1000, 1000, 3000, 2000),
					NewContainerEnvelope(222222, "different-app-id", 2, 0.211, 1000, 1000, 3000, 2000),
					NewContainerEnvelope(333333, "an-app-id", 1, 0.211, 2000, 1000, 3000, 2000),
				}
			})

			It("should return instance memory metrics from given app", func() {
				Expect(metrics).To(ConsistOf(
					&models.AppInstanceMetric{
						AppId:         "an-app-id",
						InstanceIndex: 0,
						CollectedAt:   123456,
						Name:          models.MetricNameMemory,
						Unit:          models.UnitPercentage,
						Value:         "33",
						Timestamp:     111111,
					},
					&models.AppInstanceMetric{
						AppId:         "an-app-id",
						InstanceIndex: 1,
						CollectedAt:   123456,
						Name:          models.MetricNameMemory,
						Unit:          models.UnitPercentage,
						Value:         "67",
						Timestamp:     333333,
					},
				))
			})
		})
	})

})

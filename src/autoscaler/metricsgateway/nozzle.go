package metricsgateway

import (
	"context"
	"crypto/tls"

	loggregator "code.cloudfoundry.org/go-loggregator"
	"code.cloudfoundry.org/go-loggregator/rpc/loggregator_v2"
	"code.cloudfoundry.org/lager"
)

var selectors = []*loggregator_v2.Selector{
	{
		Message: &loggregator_v2.Selector_Gauge{
			Gauge: &loggregator_v2.GaugeSelector{},
		},
	},
	{
		Message: &loggregator_v2.Selector_Timer{
			Timer: &loggregator_v2.TimerSelector{},
		},
	},
}

type Nozzle struct {
	logger        lager.Logger
	rlpAddr       string
	tls           *tls.Config
	envelopChan   chan *loggregator_v2.Envelope
	doneChan      chan bool
	index         int
	shardID       string
	appIDs        map[string]string
	getAppIDsFunc GetAppIDsFunc
}

func NewNozzle(logger lager.Logger, index int, shardID string, rlpAddr string, tls *tls.Config, envelopChan chan *loggregator_v2.Envelope, getAppIDsFunc GetAppIDsFunc) *Nozzle {
	return &Nozzle{
		logger:        logger.Session("Nozzle"),
		index:         index,
		shardID:       shardID,
		rlpAddr:       rlpAddr,
		tls:           tls,
		envelopChan:   envelopChan,
		getAppIDsFunc: getAppIDsFunc,
		doneChan:      make(chan bool),
	}
}

func (n *Nozzle) Start() {
	go n.streamMetrics()
	n.logger.Info("nozzle-started", lager.Data{"index": n.index})
}

func (n *Nozzle) Stop() {
	n.doneChan <- true
}

func (n *Nozzle) streamMetrics() {
	streamConnector := loggregator.NewEnvelopeStreamConnector(n.rlpAddr, n.tls)
	ctx := context.Background()
	rx := streamConnector.Stream(ctx, &loggregator_v2.EgressBatchRequest{
		ShardId:   n.shardID,
		Selectors: selectors,
	})

	for {
		select {
		case <-n.doneChan:
			ctx.Done()
			n.logger.Info("nozzle-stopped", lager.Data{"index": n.index})
			return
		default:
		}
		envelops := rx()
		n.logger.Info("envs", lager.Data{"envs": envelops})
		n.processEnvelope(envelops)

	}
}

func (n *Nozzle) processEnvelope(envelops []*loggregator_v2.Envelope) {
	appIDs := n.getAppIDsFunc()
	if appIDs == nil || len(appIDs) == 0 {
		n.logger.Info("app-id-map-is-empty")
		return
	}
	for _, e := range envelops {
		_, ok := appIDs[e.SourceId]
		if ok {
			switch e.Message.(type) {
			case *loggregator_v2.Envelope_Gauge, *loggregator_v2.Envelope_Timer:
				n.logger.Info("accept-envelope", lager.Data{"index": n.index, "appID": e.SourceId, "message": e.Message})
				n.envelopChan <- e
			default:
				n.logger.Info("discard-envelope", lager.Data{"index": n.index, "appID": e.SourceId, "message": e.Message})
			}
		} else {
			n.logger.Info("discard-envelope", lager.Data{"index": n.index, "appID": e.SourceId, "message": e.Message})
		}
	}
}

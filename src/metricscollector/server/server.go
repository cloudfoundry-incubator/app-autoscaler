package server

import (
	"fmt"
	"metricscollector/cf"
	"metricscollector/config"
	"net/http"

	"code.cloudfoundry.org/lager"

	"github.com/gorilla/mux"
	"github.com/tedsuo/ifrit"
	"github.com/tedsuo/ifrit/http_server"
)

const PathMemoryMetric = "/v1/apps/{appid}/metrics/memory"
const RouteNameMemoryMetric = "memory-metric"

type VarsFunc func(w http.ResponseWriter, r *http.Request, vars map[string]string)

func (vh VarsFunc) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	vh(w, r, vars)
}

func NewServer(logger lager.Logger, conf config.ServerConfig, cfc cf.CfClient, consumer NoaaConsumer) ifrit.Runner {
	mmh := NewMemoryMetricHandler(logger, cfc, consumer)

	r := mux.NewRouter()
	r.Methods("GET").Path(PathMemoryMetric).Handler(VarsFunc(mmh.GetMemoryMetric)).Name(RouteNameMemoryMetric)

	addr := fmt.Sprintf("0.0.0.0:%d", conf.Port)
	return http_server.New(addr, r)
}

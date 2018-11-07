// This file was generated by counterfeiter
package fakes

import (
	"autoscaler/db"
	"autoscaler/models"
	"database/sql"
	"sync"
)

type FakeAppMetricDB struct {
	GetDBStatusStub        func() sql.DBStats
	getDBStatusMutex       sync.RWMutex
	getDBStatusArgsForCall []struct{}
	getDBStatusReturns     struct {
		result1 sql.DBStats
	}
	SaveAppMetricStub        func(appMetric *models.AppMetric) error
	saveAppMetricMutex       sync.RWMutex
	saveAppMetricArgsForCall []struct {
		appMetric *models.AppMetric
	}
	saveAppMetricReturns struct {
		result1 error
	}
	SaveAppMetricsInBulkStub        func(metrics []*models.AppMetric) error
	saveAppMetricsInBulkMutex       sync.RWMutex
	saveAppMetricsInBulkArgsForCall []struct {
		metrics []*models.AppMetric
	}
	saveAppMetricsInBulkReturns struct {
		result1 error
	}
	RetrieveAppMetricsStub        func(appId string, metricType string, start int64, end int64, orderType db.OrderType) ([]*models.AppMetric, error)
	retrieveAppMetricsMutex       sync.RWMutex
	retrieveAppMetricsArgsForCall []struct {
		appId      string
		metricType string
		start      int64
		end        int64
		orderType  db.OrderType
	}
	retrieveAppMetricsReturns struct {
		result1 []*models.AppMetric
		result2 error
	}
	PruneAppMetricsStub        func(before int64) error
	pruneAppMetricsMutex       sync.RWMutex
	pruneAppMetricsArgsForCall []struct {
		before int64
	}
	pruneAppMetricsReturns struct {
		result1 error
	}
	CloseStub        func() error
	closeMutex       sync.RWMutex
	closeArgsForCall []struct{}
	closeReturns     struct {
		result1 error
	}
	invocations      map[string][][]interface{}
	invocationsMutex sync.RWMutex
}

func (fake *FakeAppMetricDB) GetDBStatus() sql.DBStats {
	fake.getDBStatusMutex.Lock()
	fake.getDBStatusArgsForCall = append(fake.getDBStatusArgsForCall, struct{}{})
	fake.recordInvocation("GetDBStatus", []interface{}{})
	fake.getDBStatusMutex.Unlock()
	if fake.GetDBStatusStub != nil {
		return fake.GetDBStatusStub()
	}
	return fake.getDBStatusReturns.result1
}

func (fake *FakeAppMetricDB) GetDBStatusCallCount() int {
	fake.getDBStatusMutex.RLock()
	defer fake.getDBStatusMutex.RUnlock()
	return len(fake.getDBStatusArgsForCall)
}

func (fake *FakeAppMetricDB) GetDBStatusReturns(result1 sql.DBStats) {
	fake.GetDBStatusStub = nil
	fake.getDBStatusReturns = struct {
		result1 sql.DBStats
	}{result1}
}

func (fake *FakeAppMetricDB) SaveAppMetric(appMetric *models.AppMetric) error {
	fake.saveAppMetricMutex.Lock()
	fake.saveAppMetricArgsForCall = append(fake.saveAppMetricArgsForCall, struct {
		appMetric *models.AppMetric
	}{appMetric})
	fake.recordInvocation("SaveAppMetric", []interface{}{appMetric})
	fake.saveAppMetricMutex.Unlock()
	if fake.SaveAppMetricStub != nil {
		return fake.SaveAppMetricStub(appMetric)
	}
	return fake.saveAppMetricReturns.result1
}

func (fake *FakeAppMetricDB) SaveAppMetricCallCount() int {
	fake.saveAppMetricMutex.RLock()
	defer fake.saveAppMetricMutex.RUnlock()
	return len(fake.saveAppMetricArgsForCall)
}

func (fake *FakeAppMetricDB) SaveAppMetricArgsForCall(i int) *models.AppMetric {
	fake.saveAppMetricMutex.RLock()
	defer fake.saveAppMetricMutex.RUnlock()
	return fake.saveAppMetricArgsForCall[i].appMetric
}

func (fake *FakeAppMetricDB) SaveAppMetricReturns(result1 error) {
	fake.SaveAppMetricStub = nil
	fake.saveAppMetricReturns = struct {
		result1 error
	}{result1}
}

func (fake *FakeAppMetricDB) SaveAppMetricsInBulk(metrics []*models.AppMetric) error {
	var metricsCopy []*models.AppMetric
	if metrics != nil {
		metricsCopy = make([]*models.AppMetric, len(metrics))
		copy(metricsCopy, metrics)
	}
	fake.saveAppMetricsInBulkMutex.Lock()
	fake.saveAppMetricsInBulkArgsForCall = append(fake.saveAppMetricsInBulkArgsForCall, struct {
		metrics []*models.AppMetric
	}{metricsCopy})
	fake.recordInvocation("SaveAppMetricsInBulk", []interface{}{metricsCopy})
	fake.saveAppMetricsInBulkMutex.Unlock()
	if fake.SaveAppMetricsInBulkStub != nil {
		return fake.SaveAppMetricsInBulkStub(metrics)
	}
	return fake.saveAppMetricsInBulkReturns.result1
}

func (fake *FakeAppMetricDB) SaveAppMetricsInBulkCallCount() int {
	fake.saveAppMetricsInBulkMutex.RLock()
	defer fake.saveAppMetricsInBulkMutex.RUnlock()
	return len(fake.saveAppMetricsInBulkArgsForCall)
}

func (fake *FakeAppMetricDB) SaveAppMetricsInBulkArgsForCall(i int) []*models.AppMetric {
	fake.saveAppMetricsInBulkMutex.RLock()
	defer fake.saveAppMetricsInBulkMutex.RUnlock()
	return fake.saveAppMetricsInBulkArgsForCall[i].metrics
}

func (fake *FakeAppMetricDB) SaveAppMetricsInBulkReturns(result1 error) {
	fake.SaveAppMetricsInBulkStub = nil
	fake.saveAppMetricsInBulkReturns = struct {
		result1 error
	}{result1}
}

func (fake *FakeAppMetricDB) RetrieveAppMetrics(appId string, metricType string, start int64, end int64, orderType db.OrderType) ([]*models.AppMetric, error) {
	fake.retrieveAppMetricsMutex.Lock()
	fake.retrieveAppMetricsArgsForCall = append(fake.retrieveAppMetricsArgsForCall, struct {
		appId      string
		metricType string
		start      int64
		end        int64
		orderType  db.OrderType
	}{appId, metricType, start, end, orderType})
	fake.recordInvocation("RetrieveAppMetrics", []interface{}{appId, metricType, start, end, orderType})
	fake.retrieveAppMetricsMutex.Unlock()
	if fake.RetrieveAppMetricsStub != nil {
		return fake.RetrieveAppMetricsStub(appId, metricType, start, end, orderType)
	}
	return fake.retrieveAppMetricsReturns.result1, fake.retrieveAppMetricsReturns.result2
}

func (fake *FakeAppMetricDB) RetrieveAppMetricsCallCount() int {
	fake.retrieveAppMetricsMutex.RLock()
	defer fake.retrieveAppMetricsMutex.RUnlock()
	return len(fake.retrieveAppMetricsArgsForCall)
}

func (fake *FakeAppMetricDB) RetrieveAppMetricsArgsForCall(i int) (string, string, int64, int64, db.OrderType) {
	fake.retrieveAppMetricsMutex.RLock()
	defer fake.retrieveAppMetricsMutex.RUnlock()
	return fake.retrieveAppMetricsArgsForCall[i].appId, fake.retrieveAppMetricsArgsForCall[i].metricType, fake.retrieveAppMetricsArgsForCall[i].start, fake.retrieveAppMetricsArgsForCall[i].end, fake.retrieveAppMetricsArgsForCall[i].orderType
}

func (fake *FakeAppMetricDB) RetrieveAppMetricsReturns(result1 []*models.AppMetric, result2 error) {
	fake.RetrieveAppMetricsStub = nil
	fake.retrieveAppMetricsReturns = struct {
		result1 []*models.AppMetric
		result2 error
	}{result1, result2}
}

func (fake *FakeAppMetricDB) PruneAppMetrics(before int64) error {
	fake.pruneAppMetricsMutex.Lock()
	fake.pruneAppMetricsArgsForCall = append(fake.pruneAppMetricsArgsForCall, struct {
		before int64
	}{before})
	fake.recordInvocation("PruneAppMetrics", []interface{}{before})
	fake.pruneAppMetricsMutex.Unlock()
	if fake.PruneAppMetricsStub != nil {
		return fake.PruneAppMetricsStub(before)
	}
	return fake.pruneAppMetricsReturns.result1
}

func (fake *FakeAppMetricDB) PruneAppMetricsCallCount() int {
	fake.pruneAppMetricsMutex.RLock()
	defer fake.pruneAppMetricsMutex.RUnlock()
	return len(fake.pruneAppMetricsArgsForCall)
}

func (fake *FakeAppMetricDB) PruneAppMetricsArgsForCall(i int) int64 {
	fake.pruneAppMetricsMutex.RLock()
	defer fake.pruneAppMetricsMutex.RUnlock()
	return fake.pruneAppMetricsArgsForCall[i].before
}

func (fake *FakeAppMetricDB) PruneAppMetricsReturns(result1 error) {
	fake.PruneAppMetricsStub = nil
	fake.pruneAppMetricsReturns = struct {
		result1 error
	}{result1}
}

func (fake *FakeAppMetricDB) Close() error {
	fake.closeMutex.Lock()
	fake.closeArgsForCall = append(fake.closeArgsForCall, struct{}{})
	fake.recordInvocation("Close", []interface{}{})
	fake.closeMutex.Unlock()
	if fake.CloseStub != nil {
		return fake.CloseStub()
	}
	return fake.closeReturns.result1
}

func (fake *FakeAppMetricDB) CloseCallCount() int {
	fake.closeMutex.RLock()
	defer fake.closeMutex.RUnlock()
	return len(fake.closeArgsForCall)
}

func (fake *FakeAppMetricDB) CloseReturns(result1 error) {
	fake.CloseStub = nil
	fake.closeReturns = struct {
		result1 error
	}{result1}
}

func (fake *FakeAppMetricDB) Invocations() map[string][][]interface{} {
	fake.invocationsMutex.RLock()
	defer fake.invocationsMutex.RUnlock()
	fake.getDBStatusMutex.RLock()
	defer fake.getDBStatusMutex.RUnlock()
	fake.saveAppMetricMutex.RLock()
	defer fake.saveAppMetricMutex.RUnlock()
	fake.saveAppMetricsInBulkMutex.RLock()
	defer fake.saveAppMetricsInBulkMutex.RUnlock()
	fake.retrieveAppMetricsMutex.RLock()
	defer fake.retrieveAppMetricsMutex.RUnlock()
	fake.pruneAppMetricsMutex.RLock()
	defer fake.pruneAppMetricsMutex.RUnlock()
	fake.closeMutex.RLock()
	defer fake.closeMutex.RUnlock()
	return fake.invocations
}

func (fake *FakeAppMetricDB) recordInvocation(key string, args []interface{}) {
	fake.invocationsMutex.Lock()
	defer fake.invocationsMutex.Unlock()
	if fake.invocations == nil {
		fake.invocations = map[string][][]interface{}{}
	}
	if fake.invocations[key] == nil {
		fake.invocations[key] = [][]interface{}{}
	}
	fake.invocations[key] = append(fake.invocations[key], args)
}

var _ db.AppMetricDB = new(FakeAppMetricDB)

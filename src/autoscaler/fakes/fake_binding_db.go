// This file was generated by counterfeiter
package fakes

import (
	"autoscaler/db"
	"database/sql"
	"sync"
)

type FakeBindingDB struct {
	GetDBStatusStub        func() sql.DBStats
	getDBStatusMutex       sync.RWMutex
	getDBStatusArgsForCall []struct{}
	getDBStatusReturns     struct {
		result1 sql.DBStats
	}
	CreateServiceInstanceStub        func(serviceInstanceId string, orgId string, spaceId string) error
	createServiceInstanceMutex       sync.RWMutex
	createServiceInstanceArgsForCall []struct {
		serviceInstanceId string
		orgId             string
		spaceId           string
	}
	createServiceInstanceReturns struct {
		result1 error
	}
	DeleteServiceInstanceStub        func(serviceInstanceId string) error
	deleteServiceInstanceMutex       sync.RWMutex
	deleteServiceInstanceArgsForCall []struct {
		serviceInstanceId string
	}
	deleteServiceInstanceReturns struct {
		result1 error
	}
	CreateServiceBindingStub        func(bindingId string, serviceInstanceId string, appId string) error
	createServiceBindingMutex       sync.RWMutex
	createServiceBindingArgsForCall []struct {
		bindingId         string
		serviceInstanceId string
		appId             string
	}
	createServiceBindingReturns struct {
		result1 error
	}
	DeleteServiceBindingStub        func(bindingId string) error
	deleteServiceBindingMutex       sync.RWMutex
	deleteServiceBindingArgsForCall []struct {
		bindingId string
	}
	deleteServiceBindingReturns struct {
		result1 error
	}
	DeleteServiceBindingByAppIdStub        func(appId string) error
	deleteServiceBindingByAppIdMutex       sync.RWMutex
	deleteServiceBindingByAppIdArgsForCall []struct {
		appId string
	}
	deleteServiceBindingByAppIdReturns struct {
		result1 error
	}
	CheckServiceBindingStub        func(appId string) bool
	checkServiceBindingMutex       sync.RWMutex
	checkServiceBindingArgsForCall []struct {
		appId string
	}
	checkServiceBindingReturns struct {
		result1 bool
	}
	GetAppIdByBindingIdStub        func(bindingId string) (string, error)
	getAppIdByBindingIdMutex       sync.RWMutex
	getAppIdByBindingIdArgsForCall []struct {
		bindingId string
	}
	getAppIdByBindingIdReturns struct {
		result1 string
		result2 error
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

func (fake *FakeBindingDB) GetDBStatus() sql.DBStats {
	fake.getDBStatusMutex.Lock()
	fake.getDBStatusArgsForCall = append(fake.getDBStatusArgsForCall, struct{}{})
	fake.recordInvocation("GetDBStatus", []interface{}{})
	fake.getDBStatusMutex.Unlock()
	if fake.GetDBStatusStub != nil {
		return fake.GetDBStatusStub()
	}
	return fake.getDBStatusReturns.result1
}

func (fake *FakeBindingDB) GetDBStatusCallCount() int {
	fake.getDBStatusMutex.RLock()
	defer fake.getDBStatusMutex.RUnlock()
	return len(fake.getDBStatusArgsForCall)
}

func (fake *FakeBindingDB) GetDBStatusReturns(result1 sql.DBStats) {
	fake.GetDBStatusStub = nil
	fake.getDBStatusReturns = struct {
		result1 sql.DBStats
	}{result1}
}

func (fake *FakeBindingDB) CreateServiceInstance(serviceInstanceId string, orgId string, spaceId string) error {
	fake.createServiceInstanceMutex.Lock()
	fake.createServiceInstanceArgsForCall = append(fake.createServiceInstanceArgsForCall, struct {
		serviceInstanceId string
		orgId             string
		spaceId           string
	}{serviceInstanceId, orgId, spaceId})
	fake.recordInvocation("CreateServiceInstance", []interface{}{serviceInstanceId, orgId, spaceId})
	fake.createServiceInstanceMutex.Unlock()
	if fake.CreateServiceInstanceStub != nil {
		return fake.CreateServiceInstanceStub(serviceInstanceId, orgId, spaceId)
	}
	return fake.createServiceInstanceReturns.result1
}

func (fake *FakeBindingDB) CreateServiceInstanceCallCount() int {
	fake.createServiceInstanceMutex.RLock()
	defer fake.createServiceInstanceMutex.RUnlock()
	return len(fake.createServiceInstanceArgsForCall)
}

func (fake *FakeBindingDB) CreateServiceInstanceArgsForCall(i int) (string, string, string) {
	fake.createServiceInstanceMutex.RLock()
	defer fake.createServiceInstanceMutex.RUnlock()
	return fake.createServiceInstanceArgsForCall[i].serviceInstanceId, fake.createServiceInstanceArgsForCall[i].orgId, fake.createServiceInstanceArgsForCall[i].spaceId
}

func (fake *FakeBindingDB) CreateServiceInstanceReturns(result1 error) {
	fake.CreateServiceInstanceStub = nil
	fake.createServiceInstanceReturns = struct {
		result1 error
	}{result1}
}

func (fake *FakeBindingDB) DeleteServiceInstance(serviceInstanceId string) error {
	fake.deleteServiceInstanceMutex.Lock()
	fake.deleteServiceInstanceArgsForCall = append(fake.deleteServiceInstanceArgsForCall, struct {
		serviceInstanceId string
	}{serviceInstanceId})
	fake.recordInvocation("DeleteServiceInstance", []interface{}{serviceInstanceId})
	fake.deleteServiceInstanceMutex.Unlock()
	if fake.DeleteServiceInstanceStub != nil {
		return fake.DeleteServiceInstanceStub(serviceInstanceId)
	}
	return fake.deleteServiceInstanceReturns.result1
}

func (fake *FakeBindingDB) DeleteServiceInstanceCallCount() int {
	fake.deleteServiceInstanceMutex.RLock()
	defer fake.deleteServiceInstanceMutex.RUnlock()
	return len(fake.deleteServiceInstanceArgsForCall)
}

func (fake *FakeBindingDB) DeleteServiceInstanceArgsForCall(i int) string {
	fake.deleteServiceInstanceMutex.RLock()
	defer fake.deleteServiceInstanceMutex.RUnlock()
	return fake.deleteServiceInstanceArgsForCall[i].serviceInstanceId
}

func (fake *FakeBindingDB) DeleteServiceInstanceReturns(result1 error) {
	fake.DeleteServiceInstanceStub = nil
	fake.deleteServiceInstanceReturns = struct {
		result1 error
	}{result1}
}

func (fake *FakeBindingDB) CreateServiceBinding(bindingId string, serviceInstanceId string, appId string) error {
	fake.createServiceBindingMutex.Lock()
	fake.createServiceBindingArgsForCall = append(fake.createServiceBindingArgsForCall, struct {
		bindingId         string
		serviceInstanceId string
		appId             string
	}{bindingId, serviceInstanceId, appId})
	fake.recordInvocation("CreateServiceBinding", []interface{}{bindingId, serviceInstanceId, appId})
	fake.createServiceBindingMutex.Unlock()
	if fake.CreateServiceBindingStub != nil {
		return fake.CreateServiceBindingStub(bindingId, serviceInstanceId, appId)
	}
	return fake.createServiceBindingReturns.result1
}

func (fake *FakeBindingDB) CreateServiceBindingCallCount() int {
	fake.createServiceBindingMutex.RLock()
	defer fake.createServiceBindingMutex.RUnlock()
	return len(fake.createServiceBindingArgsForCall)
}

func (fake *FakeBindingDB) CreateServiceBindingArgsForCall(i int) (string, string, string) {
	fake.createServiceBindingMutex.RLock()
	defer fake.createServiceBindingMutex.RUnlock()
	return fake.createServiceBindingArgsForCall[i].bindingId, fake.createServiceBindingArgsForCall[i].serviceInstanceId, fake.createServiceBindingArgsForCall[i].appId
}

func (fake *FakeBindingDB) CreateServiceBindingReturns(result1 error) {
	fake.CreateServiceBindingStub = nil
	fake.createServiceBindingReturns = struct {
		result1 error
	}{result1}
}

func (fake *FakeBindingDB) DeleteServiceBinding(bindingId string) error {
	fake.deleteServiceBindingMutex.Lock()
	fake.deleteServiceBindingArgsForCall = append(fake.deleteServiceBindingArgsForCall, struct {
		bindingId string
	}{bindingId})
	fake.recordInvocation("DeleteServiceBinding", []interface{}{bindingId})
	fake.deleteServiceBindingMutex.Unlock()
	if fake.DeleteServiceBindingStub != nil {
		return fake.DeleteServiceBindingStub(bindingId)
	}
	return fake.deleteServiceBindingReturns.result1
}

func (fake *FakeBindingDB) DeleteServiceBindingCallCount() int {
	fake.deleteServiceBindingMutex.RLock()
	defer fake.deleteServiceBindingMutex.RUnlock()
	return len(fake.deleteServiceBindingArgsForCall)
}

func (fake *FakeBindingDB) DeleteServiceBindingArgsForCall(i int) string {
	fake.deleteServiceBindingMutex.RLock()
	defer fake.deleteServiceBindingMutex.RUnlock()
	return fake.deleteServiceBindingArgsForCall[i].bindingId
}

func (fake *FakeBindingDB) DeleteServiceBindingReturns(result1 error) {
	fake.DeleteServiceBindingStub = nil
	fake.deleteServiceBindingReturns = struct {
		result1 error
	}{result1}
}

func (fake *FakeBindingDB) DeleteServiceBindingByAppId(appId string) error {
	fake.deleteServiceBindingByAppIdMutex.Lock()
	fake.deleteServiceBindingByAppIdArgsForCall = append(fake.deleteServiceBindingByAppIdArgsForCall, struct {
		appId string
	}{appId})
	fake.recordInvocation("DeleteServiceBindingByAppId", []interface{}{appId})
	fake.deleteServiceBindingByAppIdMutex.Unlock()
	if fake.DeleteServiceBindingByAppIdStub != nil {
		return fake.DeleteServiceBindingByAppIdStub(appId)
	}
	return fake.deleteServiceBindingByAppIdReturns.result1
}

func (fake *FakeBindingDB) DeleteServiceBindingByAppIdCallCount() int {
	fake.deleteServiceBindingByAppIdMutex.RLock()
	defer fake.deleteServiceBindingByAppIdMutex.RUnlock()
	return len(fake.deleteServiceBindingByAppIdArgsForCall)
}

func (fake *FakeBindingDB) DeleteServiceBindingByAppIdArgsForCall(i int) string {
	fake.deleteServiceBindingByAppIdMutex.RLock()
	defer fake.deleteServiceBindingByAppIdMutex.RUnlock()
	return fake.deleteServiceBindingByAppIdArgsForCall[i].appId
}

func (fake *FakeBindingDB) DeleteServiceBindingByAppIdReturns(result1 error) {
	fake.DeleteServiceBindingByAppIdStub = nil
	fake.deleteServiceBindingByAppIdReturns = struct {
		result1 error
	}{result1}
}

func (fake *FakeBindingDB) CheckServiceBinding(appId string) bool {
	fake.checkServiceBindingMutex.Lock()
	fake.checkServiceBindingArgsForCall = append(fake.checkServiceBindingArgsForCall, struct {
		appId string
	}{appId})
	fake.recordInvocation("CheckServiceBinding", []interface{}{appId})
	fake.checkServiceBindingMutex.Unlock()
	if fake.CheckServiceBindingStub != nil {
		return fake.CheckServiceBindingStub(appId)
	}
	return fake.checkServiceBindingReturns.result1
}

func (fake *FakeBindingDB) CheckServiceBindingCallCount() int {
	fake.checkServiceBindingMutex.RLock()
	defer fake.checkServiceBindingMutex.RUnlock()
	return len(fake.checkServiceBindingArgsForCall)
}

func (fake *FakeBindingDB) CheckServiceBindingArgsForCall(i int) string {
	fake.checkServiceBindingMutex.RLock()
	defer fake.checkServiceBindingMutex.RUnlock()
	return fake.checkServiceBindingArgsForCall[i].appId
}

func (fake *FakeBindingDB) CheckServiceBindingReturns(result1 bool) {
	fake.CheckServiceBindingStub = nil
	fake.checkServiceBindingReturns = struct {
		result1 bool
	}{result1}
}

func (fake *FakeBindingDB) GetAppIdByBindingId(bindingId string) (string, error) {
	fake.getAppIdByBindingIdMutex.Lock()
	fake.getAppIdByBindingIdArgsForCall = append(fake.getAppIdByBindingIdArgsForCall, struct {
		bindingId string
	}{bindingId})
	fake.recordInvocation("GetAppIdByBindingId", []interface{}{bindingId})
	fake.getAppIdByBindingIdMutex.Unlock()
	if fake.GetAppIdByBindingIdStub != nil {
		return fake.GetAppIdByBindingIdStub(bindingId)
	}
	return fake.getAppIdByBindingIdReturns.result1, fake.getAppIdByBindingIdReturns.result2
}

func (fake *FakeBindingDB) GetAppIdByBindingIdCallCount() int {
	fake.getAppIdByBindingIdMutex.RLock()
	defer fake.getAppIdByBindingIdMutex.RUnlock()
	return len(fake.getAppIdByBindingIdArgsForCall)
}

func (fake *FakeBindingDB) GetAppIdByBindingIdArgsForCall(i int) string {
	fake.getAppIdByBindingIdMutex.RLock()
	defer fake.getAppIdByBindingIdMutex.RUnlock()
	return fake.getAppIdByBindingIdArgsForCall[i].bindingId
}

func (fake *FakeBindingDB) GetAppIdByBindingIdReturns(result1 string, result2 error) {
	fake.GetAppIdByBindingIdStub = nil
	fake.getAppIdByBindingIdReturns = struct {
		result1 string
		result2 error
	}{result1, result2}
}

func (fake *FakeBindingDB) Close() error {
	fake.closeMutex.Lock()
	fake.closeArgsForCall = append(fake.closeArgsForCall, struct{}{})
	fake.recordInvocation("Close", []interface{}{})
	fake.closeMutex.Unlock()
	if fake.CloseStub != nil {
		return fake.CloseStub()
	}
	return fake.closeReturns.result1
}

func (fake *FakeBindingDB) CloseCallCount() int {
	fake.closeMutex.RLock()
	defer fake.closeMutex.RUnlock()
	return len(fake.closeArgsForCall)
}

func (fake *FakeBindingDB) CloseReturns(result1 error) {
	fake.CloseStub = nil
	fake.closeReturns = struct {
		result1 error
	}{result1}
}

func (fake *FakeBindingDB) Invocations() map[string][][]interface{} {
	fake.invocationsMutex.RLock()
	defer fake.invocationsMutex.RUnlock()
	fake.getDBStatusMutex.RLock()
	defer fake.getDBStatusMutex.RUnlock()
	fake.createServiceInstanceMutex.RLock()
	defer fake.createServiceInstanceMutex.RUnlock()
	fake.deleteServiceInstanceMutex.RLock()
	defer fake.deleteServiceInstanceMutex.RUnlock()
	fake.createServiceBindingMutex.RLock()
	defer fake.createServiceBindingMutex.RUnlock()
	fake.deleteServiceBindingMutex.RLock()
	defer fake.deleteServiceBindingMutex.RUnlock()
	fake.deleteServiceBindingByAppIdMutex.RLock()
	defer fake.deleteServiceBindingByAppIdMutex.RUnlock()
	fake.checkServiceBindingMutex.RLock()
	defer fake.checkServiceBindingMutex.RUnlock()
	fake.getAppIdByBindingIdMutex.RLock()
	defer fake.getAppIdByBindingIdMutex.RUnlock()
	fake.closeMutex.RLock()
	defer fake.closeMutex.RUnlock()
	return fake.invocations
}

func (fake *FakeBindingDB) recordInvocation(key string, args []interface{}) {
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

var _ db.BindingDB = new(FakeBindingDB)

// Code generated by counterfeiter. DO NOT EDIT.
package fakes

import (
	"sync"

	"github.com/alphagov/paas-rds-broker/rdsbroker"
)

type FakeParameterGroupSelector struct {
	SelectParameterGroupStub        func(rdsbroker.ServicePlan, rdsbroker.ProvisionParameters) (string, error)
	selectParameterGroupMutex       sync.RWMutex
	selectParameterGroupArgsForCall []struct {
		arg1 rdsbroker.ServicePlan
		arg2 rdsbroker.ProvisionParameters
	}
	selectParameterGroupReturns struct {
		result1 string
		result2 error
	}
	selectParameterGroupReturnsOnCall map[int]struct {
		result1 string
		result2 error
	}
	invocations      map[string][][]interface{}
	invocationsMutex sync.RWMutex
}

func (fake *FakeParameterGroupSelector) SelectParameterGroup(arg1 rdsbroker.ServicePlan, arg2 rdsbroker.ProvisionParameters) (string, error) {
	fake.selectParameterGroupMutex.Lock()
	ret, specificReturn := fake.selectParameterGroupReturnsOnCall[len(fake.selectParameterGroupArgsForCall)]
	fake.selectParameterGroupArgsForCall = append(fake.selectParameterGroupArgsForCall, struct {
		arg1 rdsbroker.ServicePlan
		arg2 rdsbroker.ProvisionParameters
	}{arg1, arg2})
	fake.recordInvocation("SelectParameterGroup", []interface{}{arg1, arg2})
	fake.selectParameterGroupMutex.Unlock()
	if fake.SelectParameterGroupStub != nil {
		return fake.SelectParameterGroupStub(arg1, arg2)
	}
	if specificReturn {
		return ret.result1, ret.result2
	}
	fakeReturns := fake.selectParameterGroupReturns
	return fakeReturns.result1, fakeReturns.result2
}

func (fake *FakeParameterGroupSelector) SelectParameterGroupCallCount() int {
	fake.selectParameterGroupMutex.RLock()
	defer fake.selectParameterGroupMutex.RUnlock()
	return len(fake.selectParameterGroupArgsForCall)
}

func (fake *FakeParameterGroupSelector) SelectParameterGroupCalls(stub func(rdsbroker.ServicePlan, rdsbroker.ProvisionParameters) (string, error)) {
	fake.selectParameterGroupMutex.Lock()
	defer fake.selectParameterGroupMutex.Unlock()
	fake.SelectParameterGroupStub = stub
}

func (fake *FakeParameterGroupSelector) SelectParameterGroupArgsForCall(i int) (rdsbroker.ServicePlan, rdsbroker.ProvisionParameters) {
	fake.selectParameterGroupMutex.RLock()
	defer fake.selectParameterGroupMutex.RUnlock()
	argsForCall := fake.selectParameterGroupArgsForCall[i]
	return argsForCall.arg1, argsForCall.arg2
}

func (fake *FakeParameterGroupSelector) SelectParameterGroupReturns(result1 string, result2 error) {
	fake.selectParameterGroupMutex.Lock()
	defer fake.selectParameterGroupMutex.Unlock()
	fake.SelectParameterGroupStub = nil
	fake.selectParameterGroupReturns = struct {
		result1 string
		result2 error
	}{result1, result2}
}

func (fake *FakeParameterGroupSelector) SelectParameterGroupReturnsOnCall(i int, result1 string, result2 error) {
	fake.selectParameterGroupMutex.Lock()
	defer fake.selectParameterGroupMutex.Unlock()
	fake.SelectParameterGroupStub = nil
	if fake.selectParameterGroupReturnsOnCall == nil {
		fake.selectParameterGroupReturnsOnCall = make(map[int]struct {
			result1 string
			result2 error
		})
	}
	fake.selectParameterGroupReturnsOnCall[i] = struct {
		result1 string
		result2 error
	}{result1, result2}
}

func (fake *FakeParameterGroupSelector) Invocations() map[string][][]interface{} {
	fake.invocationsMutex.RLock()
	defer fake.invocationsMutex.RUnlock()
	fake.selectParameterGroupMutex.RLock()
	defer fake.selectParameterGroupMutex.RUnlock()
	copiedInvocations := map[string][][]interface{}{}
	for key, value := range fake.invocations {
		copiedInvocations[key] = value
	}
	return copiedInvocations
}

func (fake *FakeParameterGroupSelector) recordInvocation(key string, args []interface{}) {
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

var _ rdsbroker.ParameterGroupSelector = new(FakeParameterGroupSelector)

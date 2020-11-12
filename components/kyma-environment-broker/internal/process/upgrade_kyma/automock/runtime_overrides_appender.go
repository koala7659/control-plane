// Code generated by mockery v2.3.0. DO NOT EDIT.

package automock

import (
	runtimeoverrides "github.com/kyma-project/control-plane/components/kyma-environment-broker/internal/runtimeoverrides"
	mock "github.com/stretchr/testify/mock"
)

// RuntimeOverridesAppender is an autogenerated mock type for the RuntimeOverridesAppender type
type RuntimeOverridesAppender struct {
	mock.Mock
}

// Append provides a mock function with given fields: input, planID, kymaVersion
func (_m *RuntimeOverridesAppender) Append(input runtimeoverrides.InputAppender, planID string, kymaVersion string) error {
	ret := _m.Called(input, planID, kymaVersion)

	var r0 error
	if rf, ok := ret.Get(0).(func(runtimeoverrides.InputAppender, string, string) error); ok {
		r0 = rf(input, planID, kymaVersion)
	} else {
		r0 = ret.Error(0)
	}

	return r0
}
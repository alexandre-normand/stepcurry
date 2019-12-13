// Code generated by mockery v1.0.0. DO NOT EDIT.

package mocks

import context "context"
import gax "github.com/googleapis/gax-go/v2"
import mock "github.com/stretchr/testify/mock"

import tasks "google.golang.org/genproto/googleapis/cloud/tasks/v2beta3"

// TaskScheduler is an autogenerated mock type for the TaskScheduler type
type TaskScheduler struct {
	mock.Mock
}

// Connect provides a mock function with given fields:
func (_m *TaskScheduler) Connect() error {
	ret := _m.Called()

	var r0 error
	if rf, ok := ret.Get(0).(func() error); ok {
		r0 = rf()
	} else {
		r0 = ret.Error(0)
	}

	return r0
}

// CreateTask provides a mock function with given fields: ctx, req, opts
func (_m *TaskScheduler) CreateTask(ctx context.Context, req *tasks.CreateTaskRequest, opts ...gax.CallOption) (*tasks.Task, error) {
	_va := make([]interface{}, len(opts))
	for _i := range opts {
		_va[_i] = opts[_i]
	}
	var _ca []interface{}
	_ca = append(_ca, ctx, req)
	_ca = append(_ca, _va...)
	ret := _m.Called(_ca...)

	var r0 *tasks.Task
	if rf, ok := ret.Get(0).(func(context.Context, *tasks.CreateTaskRequest, ...gax.CallOption) *tasks.Task); ok {
		r0 = rf(ctx, req, opts...)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(*tasks.Task)
		}
	}

	var r1 error
	if rf, ok := ret.Get(1).(func(context.Context, *tasks.CreateTaskRequest, ...gax.CallOption) error); ok {
		r1 = rf(ctx, req, opts...)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// GenerateQueueID provides a mock function with given fields:
func (_m *TaskScheduler) GenerateQueueID() string {
	ret := _m.Called()

	var r0 string
	if rf, ok := ret.Get(0).(func() string); ok {
		r0 = rf()
	} else {
		r0 = ret.Get(0).(string)
	}

	return r0
}

// Code generated by mockery v1.0.0. DO NOT EDIT.

package mocks

import mock "github.com/stretchr/testify/mock"

import slack "github.com/nlopes/slack"

// UserInfoFinder is an autogenerated mock type for the UserInfoFinder type
type UserInfoFinder struct {
	mock.Mock
}

// GetBotInfo provides a mock function with given fields: botID
func (_m *UserInfoFinder) GetBotInfo(botID string) (*slack.Bot, error) {
	ret := _m.Called(botID)

	var r0 *slack.Bot
	if rf, ok := ret.Get(0).(func(string) *slack.Bot); ok {
		r0 = rf(botID)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(*slack.Bot)
		}
	}

	var r1 error
	if rf, ok := ret.Get(1).(func(string) error); ok {
		r1 = rf(botID)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// GetUserInfo provides a mock function with given fields: userID
func (_m *UserInfoFinder) GetUserInfo(userID string) (*slack.User, error) {
	ret := _m.Called(userID)

	var r0 *slack.User
	if rf, ok := ret.Get(0).(func(string) *slack.User); ok {
		r0 = rf(userID)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(*slack.User)
		}
	}

	var r1 error
	if rf, ok := ret.Get(1).(func(string) error); ok {
		r1 = rf(userID)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

package rogerchallenger

import (
	"fmt"
	"github.com/alexandre-normand/rogerchallenger/mocks"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"testing"
)

func TestNewRogerChallenger(t *testing.T) {
	verifier := &mocks.Verifier{}
	defer verifier.AssertExpectations(t)

	storer := &mocks.Datastorer{}
	defer storer.AssertExpectations(t)

	messenger := &mocks.Messenger{}
	defer messenger.AssertExpectations(t)

	taskScheduler := &mocks.TaskScheduler{}
	defer taskScheduler.AssertExpectations(t)

	userInfoFinder := &mocks.UserInfoFinder{}
	defer userInfoFinder.AssertExpectations(t)

	channelInfoFinder := &mocks.ChannelInfoFinder{}
	defer channelInfoFinder.AssertExpectations(t)

	tests := map[string]struct {
		baseURL            string
		fitbitClientID     string
		fitbitClientSecret string
		opts               []Option
		expectedInstance   *RogerChallenger
		expectedErr        error
	}{
		"WithAllDependencies": {
			baseURL:            "https://rogerchallenger.com",
			fitbitClientID:     "clientID1",
			fitbitClientSecret: "clientSecret1",
			opts:               []Option{OptionMessenger(messenger), OptionStorer(storer), OptionVerifier(verifier), OptionTaskScheduler(taskScheduler), OptionChannelInfoFinder(channelInfoFinder), OptionUserInfoFinder(userInfoFinder)},
			expectedInstance:   &RogerChallenger{baseURL: "https://rogerchallenger.com", fitbitAPIBaseURL: defaultFitbitAPIBaseURL, fitbitAuthBaseURL: defaultFitbitAuthBaseURL, fitbitClientID: "clientID1", fitbitClientSecret: "clientSecret1"},
			expectedErr:        nil},
		"WithFitbitURLsOverride": {
			baseURL:            "https://rogerchallenger.com",
			fitbitClientID:     "clientID1",
			fitbitClientSecret: "clientSecret1",
			opts:               []Option{OptionFitbitURLs("https://beta.fitbit.com/auth", "https://beta.api.fitbit.com"), OptionMessenger(messenger), OptionStorer(storer), OptionVerifier(verifier), OptionTaskScheduler(taskScheduler), OptionChannelInfoFinder(channelInfoFinder), OptionUserInfoFinder(userInfoFinder)},
			expectedInstance:   &RogerChallenger{baseURL: "https://rogerchallenger.com", fitbitAPIBaseURL: "https://beta.api.fitbit.com", fitbitAuthBaseURL: "https://beta.fitbit.com/auth", fitbitClientID: "clientID1", fitbitClientSecret: "clientSecret1"},
			expectedErr:        nil},
		"WithoutDatastorer": {
			baseURL:            "",
			fitbitClientID:     "",
			fitbitClientSecret: "",
			opts:               []Option{OptionMessenger(messenger), OptionVerifier(verifier), OptionTaskScheduler(taskScheduler), OptionChannelInfoFinder(channelInfoFinder), OptionUserInfoFinder(userInfoFinder)},
			expectedInstance:   nil,
			expectedErr:        fmt.Errorf("storer is nil after applying all Options. Did you forget to set one?")},
		"WithoutVerifier": {
			baseURL:            "",
			fitbitClientID:     "",
			fitbitClientSecret: "",
			opts:               []Option{OptionMessenger(messenger), OptionStorer(storer), OptionTaskScheduler(taskScheduler), OptionChannelInfoFinder(channelInfoFinder), OptionUserInfoFinder(userInfoFinder)},
			expectedInstance:   nil,
			expectedErr:        fmt.Errorf("verifier is nil after applying all Options. Did you forget to set one?")},
		"WithoutTaskScheduler": {
			baseURL:            "",
			fitbitClientID:     "",
			fitbitClientSecret: "",
			opts:               []Option{OptionMessenger(messenger), OptionStorer(storer), OptionVerifier(verifier), OptionChannelInfoFinder(channelInfoFinder), OptionUserInfoFinder(userInfoFinder)},
			expectedInstance:   nil,
			expectedErr:        fmt.Errorf("taskScheduler is nil after applying all Options. Did you forget to set one?")},
		"WithoutChannelInfoFinder": {
			baseURL:            "",
			fitbitClientID:     "",
			fitbitClientSecret: "",
			opts:               []Option{OptionMessenger(messenger), OptionStorer(storer), OptionVerifier(verifier), OptionTaskScheduler(taskScheduler), OptionUserInfoFinder(userInfoFinder)},
			expectedInstance:   nil,
			expectedErr:        fmt.Errorf("channelInfoFinder is nil after applying all Options. Did you forget to set one?")},
		"WithoutUserInfoFinder": {
			baseURL:            "",
			fitbitClientID:     "",
			fitbitClientSecret: "",
			opts:               []Option{OptionMessenger(messenger), OptionStorer(storer), OptionVerifier(verifier), OptionTaskScheduler(taskScheduler), OptionChannelInfoFinder(channelInfoFinder)},
			expectedInstance:   nil,
			expectedErr:        fmt.Errorf("userInfoFinder is nil after applying all Options. Did you forget to set one?")},
		"WithoutMessenger": {
			baseURL:            "",
			fitbitClientID:     "",
			fitbitClientSecret: "",
			opts:               []Option{OptionStorer(storer), OptionVerifier(verifier), OptionTaskScheduler(taskScheduler), OptionChannelInfoFinder(channelInfoFinder), OptionUserInfoFinder(userInfoFinder)},
			expectedInstance:   nil,
			expectedErr:        fmt.Errorf("messenger is nil after applying all Options. Did you forget to set one?")},
	}

	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
			testRc, err := New(tc.baseURL, tc.fitbitClientID, tc.fitbitClientSecret, tc.opts...)

			if tc.expectedErr != nil {
				require.EqualError(t, err, tc.expectedErr.Error())
				assert.Nil(t, testRc)
			} else {
				assert.Equal(t, tc.expectedInstance.baseURL, testRc.baseURL)
				assert.Equal(t, tc.expectedInstance.fitbitClientID, testRc.fitbitClientID)
				assert.Equal(t, tc.expectedInstance.fitbitClientSecret, testRc.fitbitClientSecret)
				assert.Equal(t, tc.expectedInstance.fitbitAuthBaseURL, testRc.fitbitAuthBaseURL)
				assert.Equal(t, tc.expectedInstance.fitbitAPIBaseURL, testRc.fitbitAPIBaseURL)
			}
		})
	}
}

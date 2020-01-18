package stepcurry

import (
	"fmt"
	"github.com/alexandre-normand/stepcurry/mocks"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"testing"
)

func TestNewStepCurry(t *testing.T) {
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

	conversationMemberFinder := &mocks.ConversationMemberFinder{}
	defer conversationMemberFinder.AssertExpectations(t)

	teamRouter := &SingleTenantRouter{}

	tests := map[string]struct {
		baseURL            string
		appID              string
		fitbitClientID     string
		fitbitClientSecret string
		slackClientID      string
		slackClientSecret  string
		opts               []Option
		expectedInstance   *StepCurry
		expectedErr        error
	}{
		"WithAllDependencies": {
			baseURL:            "https://stepcurry.com",
			appID:              "roger",
			fitbitClientID:     "clientID1",
			fitbitClientSecret: "clientSecret1",
			slackClientID:      "slackID1",
			slackClientSecret:  "slackSecret1",
			opts:               []Option{OptionTeamRouter(teamRouter), OptionStorer(storer), OptionVerifier(verifier), OptionTaskScheduler(taskScheduler)},
			expectedInstance:   &StepCurry{baseURL: "https://stepcurry.com", slackAppID: "roger", fitbitAPIBaseURL: defaultFitbitAPIBaseURL, fitbitAuthBaseURL: defaultFitbitAuthBaseURL, slackBaseURL: defaultSlackBaseURL, fitbitClientID: "clientID1", fitbitClientSecret: "clientSecret1", slackClientID: "slackID1", slackClientSecret: "slackSecret1", slashCommands: SlashCommands{Link: commandLinkFitbit, Challenge: commandChallenge, Standings: commandStandings}, paths: Paths{UpdateChallenge: updateChallengePath, FitbitAuthCallback: oauthCallbackPath, LinkAccount: linkAccountPath, StartChallenge: startChallengePath, Standings: standingsPath}},
			expectedErr:        nil},
		"WithFitbitURLsOverride": {
			baseURL:            "https://stepcurry.com",
			appID:              "roger",
			fitbitClientID:     "clientID1",
			fitbitClientSecret: "clientSecret1",
			slackClientID:      "slackID1",
			slackClientSecret:  "slackSecret1",
			opts:               []Option{OptionFitbitURLs("https://beta.fitbit.com/auth", "https://beta.api.fitbit.com"), OptionTeamRouter(teamRouter), OptionStorer(storer), OptionVerifier(verifier), OptionTaskScheduler(taskScheduler)},
			expectedInstance:   &StepCurry{baseURL: "https://stepcurry.com", slackAppID: "roger", fitbitAPIBaseURL: "https://beta.api.fitbit.com", fitbitAuthBaseURL: "https://beta.fitbit.com/auth", slackBaseURL: defaultSlackBaseURL, fitbitClientID: "clientID1", fitbitClientSecret: "clientSecret1", slackClientID: "slackID1", slackClientSecret: "slackSecret1", slashCommands: SlashCommands{Link: commandLinkFitbit, Challenge: commandChallenge, Standings: commandStandings}, paths: Paths{UpdateChallenge: updateChallengePath, FitbitAuthCallback: oauthCallbackPath, LinkAccount: linkAccountPath, StartChallenge: startChallengePath, Standings: standingsPath}},
			expectedErr:        nil},
		"WithSlackURLOverride": {
			baseURL:            "https://stepcurry.com",
			appID:              "slackRoger",
			fitbitClientID:     "clientID1",
			fitbitClientSecret: "clientSecret1",
			slackClientID:      "slackID1",
			slackClientSecret:  "slackSecret1",
			opts:               []Option{OptionSlackBaseURL("https://slack.io"), OptionTeamRouter(teamRouter), OptionStorer(storer), OptionVerifier(verifier), OptionTaskScheduler(taskScheduler)},
			expectedInstance:   &StepCurry{baseURL: "https://stepcurry.com", slackAppID: "slackRoger", fitbitAPIBaseURL: defaultFitbitAPIBaseURL, fitbitAuthBaseURL: defaultFitbitAuthBaseURL, slackBaseURL: "https://slack.io", fitbitClientID: "clientID1", fitbitClientSecret: "clientSecret1", slackClientID: "slackID1", slackClientSecret: "slackSecret1", slashCommands: SlashCommands{Link: commandLinkFitbit, Challenge: commandChallenge, Standings: commandStandings}, paths: Paths{UpdateChallenge: updateChallengePath, FitbitAuthCallback: oauthCallbackPath, LinkAccount: linkAccountPath, StartChallenge: startChallengePath, Standings: standingsPath}},
			expectedErr:        nil},
		"WithSlashCommandsOverride": {
			baseURL:            "https://stepcurry.com",
			appID:              "slackRoger",
			fitbitClientID:     "clientID1",
			fitbitClientSecret: "clientSecret1",
			slackClientID:      "slackID1",
			slackClientSecret:  "slackSecret1",
			opts:               []Option{OptionSlackBaseURL("https://slack.io"), OptionTeamRouter(teamRouter), OptionStorer(storer), OptionVerifier(verifier), OptionTaskScheduler(taskScheduler), OptionSlashCommands(SlashCommands{Link: "/roger-link", Challenge: "/roger-challenge", Standings: "/roger-standings"})},
			expectedInstance:   &StepCurry{baseURL: "https://stepcurry.com", slackAppID: "slackRoger", fitbitAPIBaseURL: defaultFitbitAPIBaseURL, fitbitAuthBaseURL: defaultFitbitAuthBaseURL, slackBaseURL: "https://slack.io", fitbitClientID: "clientID1", fitbitClientSecret: "clientSecret1", slackClientID: "slackID1", slackClientSecret: "slackSecret1", slashCommands: SlashCommands{Link: "/roger-link", Challenge: "/roger-challenge", Standings: "/roger-standings"}, paths: Paths{UpdateChallenge: updateChallengePath, FitbitAuthCallback: oauthCallbackPath, LinkAccount: linkAccountPath, StartChallenge: startChallengePath, Standings: standingsPath}},
			expectedErr:        nil},
		"WithPathsOverride": {
			baseURL:            "https://stepcurry.com",
			appID:              "slackRoger",
			fitbitClientID:     "clientID1",
			fitbitClientSecret: "clientSecret1",
			slackClientID:      "slackID1",
			slackClientSecret:  "slackSecret1",
			opts:               []Option{OptionSlackBaseURL("https://slack.io"), OptionTeamRouter(teamRouter), OptionStorer(storer), OptionVerifier(verifier), OptionTaskScheduler(taskScheduler), OptionPaths(Paths{UpdateChallenge: "upt", FitbitAuthCallback: "callback", LinkAccount: "link", StartChallenge: "start", Standings: "stand"})},
			expectedInstance:   &StepCurry{baseURL: "https://stepcurry.com", slackAppID: "slackRoger", fitbitAPIBaseURL: defaultFitbitAPIBaseURL, fitbitAuthBaseURL: defaultFitbitAuthBaseURL, slackBaseURL: "https://slack.io", fitbitClientID: "clientID1", fitbitClientSecret: "clientSecret1", slackClientID: "slackID1", slackClientSecret: "slackSecret1", slashCommands: SlashCommands{Link: commandLinkFitbit, Challenge: commandChallenge, Standings: commandStandings}, paths: Paths{UpdateChallenge: "upt", FitbitAuthCallback: "callback", LinkAccount: "link", StartChallenge: "start", Standings: "stand"}},
			expectedErr:        nil},
		"WithoutDatastorer": {
			baseURL:            "",
			fitbitClientID:     "",
			fitbitClientSecret: "",
			slackClientID:      "",
			slackClientSecret:  "",
			opts:               []Option{OptionTeamRouter(teamRouter), OptionVerifier(verifier), OptionTaskScheduler(taskScheduler)},
			expectedInstance:   nil,
			expectedErr:        fmt.Errorf("storer is nil after applying all Options. Did you forget to set one?")},
		"WithoutVerifier": {
			baseURL:            "",
			fitbitClientID:     "",
			fitbitClientSecret: "",
			slackClientID:      "",
			slackClientSecret:  "",
			opts:               []Option{OptionTeamRouter(teamRouter), OptionStorer(storer), OptionTaskScheduler(taskScheduler)},
			expectedInstance:   nil,
			expectedErr:        fmt.Errorf("verifier is nil after applying all Options. Did you forget to set one?")},
		"WithoutTaskScheduler": {
			baseURL:            "",
			fitbitClientID:     "",
			fitbitClientSecret: "",
			slackClientID:      "",
			slackClientSecret:  "",
			opts:               []Option{OptionTeamRouter(teamRouter), OptionStorer(storer), OptionVerifier(verifier)},
			expectedInstance:   nil,
			expectedErr:        fmt.Errorf("taskScheduler is nil after applying all Options. Did you forget to set one?")},
		"WithoutTeamRouter": {
			baseURL:            "",
			fitbitClientID:     "",
			fitbitClientSecret: "",
			slackClientID:      "",
			slackClientSecret:  "",
			opts:               []Option{OptionStorer(storer), OptionVerifier(verifier), OptionTaskScheduler(taskScheduler)},
			expectedInstance:   nil,
			expectedErr:        fmt.Errorf("teamRouter is nil after applying all Options. Did you forget to set one?")},
	}

	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
			testRc, err := New(tc.baseURL, tc.appID, tc.fitbitClientID, tc.fitbitClientSecret, tc.slackClientID, tc.slackClientSecret, tc.opts...)

			if tc.expectedErr != nil {
				require.EqualError(t, err, tc.expectedErr.Error())
				assert.Nil(t, testRc)
			} else {
				assert.Equal(t, tc.expectedInstance.baseURL, testRc.baseURL)
				assert.Equal(t, tc.expectedInstance.slashCommands, testRc.slashCommands)
				assert.Equal(t, tc.expectedInstance.slackAppID, testRc.slackAppID)
				assert.Equal(t, tc.expectedInstance.fitbitClientID, testRc.fitbitClientID)
				assert.Equal(t, tc.expectedInstance.fitbitClientSecret, testRc.fitbitClientSecret)
				assert.Equal(t, tc.expectedInstance.slackClientID, testRc.slackClientID)
				assert.Equal(t, tc.expectedInstance.slackClientSecret, testRc.slackClientSecret)
				assert.Equal(t, tc.expectedInstance.slackBaseURL, testRc.slackBaseURL)
				assert.Equal(t, tc.expectedInstance.fitbitAuthBaseURL, testRc.fitbitAuthBaseURL)
				assert.Equal(t, tc.expectedInstance.fitbitAPIBaseURL, testRc.fitbitAPIBaseURL)
			}
		})
	}
}

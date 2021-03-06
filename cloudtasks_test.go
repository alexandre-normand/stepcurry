package stepcurry

import (
	"fmt"
	"github.com/alexandre-normand/stepcurry/mocks"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
	taskspb "google.golang.org/genproto/googleapis/cloud/tasks/v2beta3"
	"testing"
	"time"
)

func TestScheduleChallengeUpdate(t *testing.T) {
	verifier := &mocks.Verifier{}
	defer verifier.AssertExpectations(t)

	storer := &mocks.Datastorer{}
	defer storer.AssertExpectations(t)

	messenger := &mocks.Messenger{}
	defer messenger.AssertExpectations(t)

	taskScheduler := &mocks.TaskScheduler{}
	taskScheduler.On("GenerateQueueID").Return("queue/path")
	taskScheduler.On("CreateTask", mock.Anything, mock.MatchedBy(func(req *taskspb.CreateTaskRequest) bool {
		return req.GetParent() == "queue/path" && req.GetTask().GetHttpRequest().GetHttpMethod() == taskspb.HttpMethod_POST &&
			req.GetTask().GetHttpRequest().GetUrl() == "https://stepcurry.com/"+updateChallengePath && string(req.GetTask().GetHttpRequest().GetBody()) == "{\"ChannelID\":\"CID\",\"TeamID\":\"TEAMID\",\"Date\":\"2019-10-11\"}"
	})).Return(nil, nil)
	defer taskScheduler.AssertExpectations(t)

	userInfoFinder := &mocks.UserInfoFinder{}
	defer userInfoFinder.AssertExpectations(t)

	conversationMemberFinder := &mocks.ConversationMemberFinder{}
	defer conversationMemberFinder.AssertExpectations(t)

	teamRouter, err := NewSingleTenantRouter(userInfoFinder, nil, messenger, conversationMemberFinder)
	require.NoError(t, err)

	sc, err := New("https://stepcurry.com", "roger", "fitbitClientID", "fitbitClientSecret", "slackClientID", "slackClientSecret", OptionTeamRouter(teamRouter), OptionStorer(storer), OptionVerifier(verifier), OptionTaskScheduler(taskScheduler))
	require.NoError(t, err)

	err = sc.scheduleChallengeUpdate(ChallengeID{TeamID: "TEAMID", ChannelID: "CID", Date: "2019-10-11"}, time.Date(2019, 10, 12, 8, 0, 0, 0, time.UTC))
	require.NoError(t, err)
}

func TestScheduleChallengeUpdateErrorOnCreateTask(t *testing.T) {
	verifier := &mocks.Verifier{}
	defer verifier.AssertExpectations(t)

	storer := &mocks.Datastorer{}
	defer storer.AssertExpectations(t)

	messenger := &mocks.Messenger{}
	defer messenger.AssertExpectations(t)

	taskScheduler := &mocks.TaskScheduler{}
	taskScheduler.On("GenerateQueueID").Return("queue/path")
	taskScheduler.On("CreateTask", mock.Anything, mock.MatchedBy(func(req *taskspb.CreateTaskRequest) bool {
		return req.GetParent() == "queue/path" && req.GetTask().GetHttpRequest().GetHttpMethod() == taskspb.HttpMethod_POST &&
			req.GetTask().GetHttpRequest().GetUrl() == "https://stepcurry.com/"+updateChallengePath && string(req.GetTask().GetHttpRequest().GetBody()) == "{\"ChannelID\":\"CID\",\"TeamID\":\"TEAMID\",\"Date\":\"2019-10-11\"}"
	})).Return(nil, fmt.Errorf("cloud tasks unavailable"))
	defer taskScheduler.AssertExpectations(t)

	userInfoFinder := &mocks.UserInfoFinder{}
	defer userInfoFinder.AssertExpectations(t)

	conversationMemberFinder := &mocks.ConversationMemberFinder{}
	defer conversationMemberFinder.AssertExpectations(t)
	teamRouter, err := NewSingleTenantRouter(userInfoFinder, nil, messenger, conversationMemberFinder)
	require.NoError(t, err)

	sc, err := New("https://stepcurry.com", "roger", "fitbitClientID", "fitbitClientSecret", "slackClientID", "slackClientSecret", OptionTeamRouter(teamRouter), OptionStorer(storer), OptionVerifier(verifier), OptionTaskScheduler(taskScheduler))
	require.NoError(t, err)

	err = sc.scheduleChallengeUpdate(ChallengeID{TeamID: "TEAMID", ChannelID: "CID", Date: "2019-10-11"}, time.Date(2019, 10, 12, 8, 0, 0, 0, time.UTC))
	require.Error(t, err, "cloud tasks unavailable")
}

func TestGenerateQueueID(t *testing.T) {
	gtc := &cloudTaskClient{gcpProject: "roger", gcpLocation: "us-east1", taskQueueName: "challenge-updates"}

	assert.Equal(t, "projects/roger/locations/us-east1/queues/challenge-updates", gtc.GenerateQueueID())
}

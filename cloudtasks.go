package stepcurry

import (
	cloudtasks "cloud.google.com/go/cloudtasks/apiv2beta3"
	"context"
	"encoding/json"
	"fmt"
	"github.com/golang/protobuf/ptypes/timestamp"
	gax "github.com/googleapis/gax-go/v2"
	"google.golang.org/api/option"
	taskspb "google.golang.org/genproto/googleapis/cloud/tasks/v2beta3"
	"time"
)

// TaskScheduler is implemented by any value that implements all of its methods. It is
// meant to allow easier testing decoupled from an actual cloudtasks backend and the
// methods defined are methods implemented by the cloudtasks.Client that this package
// uses
type TaskScheduler interface {
	Connecter
	// GenerateQueueID generates a fully qualified task queue id
	GenerateQueueID() (queueID string)
	// CreateTask creates a task. See https://godoc.org/cloud.google.com/go/cloudtasks/apiv2beta3#Client.CreateTask
	CreateTask(ctx context.Context, req *taskspb.CreateTaskRequest, opts ...gax.CallOption) (*taskspb.Task, error)
}

// cloudTaskClient holds the data for what is an actual cloudtasks backend powered implementation of TaskScheduler
type cloudTaskClient struct {
	gcpProject       string
	gcpLocation      string
	taskQueueName    string
	client           *cloudtasks.Client
	gcloudClientOpts []option.ClientOption
}

type retryableOperationWithTask func() (task *taskspb.Task, err error)

func (ctc *cloudTaskClient) tryTaskOperationWithRecovery(operation retryableOperationWithTask) (task *taskspb.Task, err error) {
	task, err = operation()
	for attempt := 1; attempt < maxAttemptCount && err != nil && shouldRetryTaskOperation(err); attempt = attempt + 1 {
		err = ctc.Connect()
		if err == nil {
			task, err = operation()
		}
	}

	return task, err
}

// shouldRetryTaskOperation returns true if the given task API error should be retried or false if not.
// What's done here is to be a little conservative and retry on everything.
// This means we could still retry when it's pointless to do so at the expense of added latency.
func shouldRetryTaskOperation(err error) bool {
	return true
}

// GenerateQueueID generates a queue id from the gcp project, location and queue name
func (ctc *cloudTaskClient) GenerateQueueID() (queueID string) {
	return fmt.Sprintf("projects/%s/locations/%s/queues/%s", ctc.gcpProject, ctc.gcpLocation, ctc.taskQueueName)
}

// Connect creates a new cloudtasks client
func (ctc *cloudTaskClient) Connect() (err error) {
	ctx := context.Background()
	ctc.client, err = cloudtasks.NewClient(ctx, ctc.gcloudClientOpts...)
	if err != nil {
		return err
	}

	return nil
}

// CreateTask creates a task via a real cloudtasks Client
func (ctc *cloudTaskClient) CreateTask(ctx context.Context, req *taskspb.CreateTaskRequest, opts ...gax.CallOption) (task *taskspb.Task, err error) {
	return ctc.tryTaskOperationWithRecovery(func() (task *taskspb.Task, err error) {
		return ctc.client.CreateTask(ctx, req, opts...)
	})
}

// NewTaskScheduler creates a new instance of a TaskScheduler backed by the real cloudtasks Client
func NewTaskScheduler(gcpProject string, gcpLocation string, taskQueueName string, gcloudOpts ...option.ClientOption) (ctc *cloudTaskClient, err error) {
	ctc = new(cloudTaskClient)
	ctc.gcpProject = gcpProject
	ctc.gcpLocation = gcpLocation
	ctc.taskQueueName = taskQueueName
	ctc.gcloudClientOpts = gcloudOpts

	err = ctc.Connect()
	if err != nil {
		return nil, err
	}

	return ctc, nil
}

// scheduleChallengeUpdate creates a new task to update a challenge at the given scheduled time
func (sc *StepCurry) scheduleChallengeUpdate(challengeID ChallengeID, scheduledTime time.Time) (err error) {
	queueID := sc.taskScheduler.GenerateQueueID()

	scheduledTimestamp := timestamp.Timestamp{Seconds: scheduledTime.Unix()}

	req := &taskspb.CreateTaskRequest{
		Parent: queueID,
		Task: &taskspb.Task{
			PayloadType: &taskspb.Task_HttpRequest{
				HttpRequest: &taskspb.HttpRequest{
					HttpMethod: taskspb.HttpMethod_POST,
					Url:        fmt.Sprintf("%s/%s", sc.baseURL, sc.paths.UpdateChallenge),
				},
			},
			ScheduleTime: &scheduledTimestamp,
		},
	}

	message, err := json.Marshal(challengeID)
	if err != nil {
		return err
	}

	req.Task.GetHttpRequest().Body = message

	ctx := context.Background()
	_, err = sc.taskScheduler.CreateTask(ctx, req)

	return err
}

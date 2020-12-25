package stepcurry

// DO NOT EDIT!
// This code is generated with http://github.com/hexdigest/gowrap tool
// using https://raw.githubusercontent.com/alexandre-normand/slackscot/master/opentelemetry.template template

//go:generate gowrap gen -p github.com/alexandre-normand/stepcurry -i TaskScheduler -t https://raw.githubusercontent.com/alexandre-normand/slackscot/master/opentelemetry.template -o taskschedulermetrics.go

import (
	"context"
	"time"
	"unicode"

	gax "github.com/googleapis/gax-go/v2"
	"go.opentelemetry.io/otel/label"
	"go.opentelemetry.io/otel/metric"
	taskspb "google.golang.org/genproto/googleapis/cloud/tasks/v2beta3"
)

// TaskSchedulerWithTelemetry implements TaskScheduler interface with all methods wrapped
// with open telemetry metrics
type TaskSchedulerWithTelemetry struct {
	base                     TaskScheduler
	methodCounters           map[string]metric.BoundInt64Counter
	errCounters              map[string]metric.BoundInt64Counter
	methodTimeValueRecorders map[string]metric.BoundInt64ValueRecorder
}

// NewTaskSchedulerWithTelemetry returns an instance of the TaskScheduler decorated with open telemetry timing and count metrics
func NewTaskSchedulerWithTelemetry(base TaskScheduler, name string, meter metric.Meter) TaskSchedulerWithTelemetry {
	return TaskSchedulerWithTelemetry{
		base:                     base,
		methodCounters:           newTaskSchedulerMethodCounters("Calls", name, meter),
		errCounters:              newTaskSchedulerMethodCounters("Errors", name, meter),
		methodTimeValueRecorders: newTaskSchedulerMethodTimeValueRecorders(name, meter),
	}
}

func newTaskSchedulerMethodTimeValueRecorders(appName string, meter metric.Meter) (boundTimeValueRecorders map[string]metric.BoundInt64ValueRecorder) {
	boundTimeValueRecorders = make(map[string]metric.BoundInt64ValueRecorder)
	mt := metric.Must(meter)

	nConnectValRecorder := []rune("TaskScheduler_Connect_ProcessingTimeMillis")
	nConnectValRecorder[0] = unicode.ToLower(nConnectValRecorder[0])
	mConnect := mt.NewInt64ValueRecorder(string(nConnectValRecorder))
	boundTimeValueRecorders["Connect"] = mConnect.Bind(label.String("name", appName))

	nCreateTaskValRecorder := []rune("TaskScheduler_CreateTask_ProcessingTimeMillis")
	nCreateTaskValRecorder[0] = unicode.ToLower(nCreateTaskValRecorder[0])
	mCreateTask := mt.NewInt64ValueRecorder(string(nCreateTaskValRecorder))
	boundTimeValueRecorders["CreateTask"] = mCreateTask.Bind(label.String("name", appName))

	nGenerateQueueIDValRecorder := []rune("TaskScheduler_GenerateQueueID_ProcessingTimeMillis")
	nGenerateQueueIDValRecorder[0] = unicode.ToLower(nGenerateQueueIDValRecorder[0])
	mGenerateQueueID := mt.NewInt64ValueRecorder(string(nGenerateQueueIDValRecorder))
	boundTimeValueRecorders["GenerateQueueID"] = mGenerateQueueID.Bind(label.String("name", appName))

	return boundTimeValueRecorders
}

func newTaskSchedulerMethodCounters(suffix string, appName string, meter metric.Meter) (boundCounters map[string]metric.BoundInt64Counter) {
	boundCounters = make(map[string]metric.BoundInt64Counter)
	mt := metric.Must(meter)

	nConnectCounter := []rune("TaskScheduler_Connect_" + suffix)
	nConnectCounter[0] = unicode.ToLower(nConnectCounter[0])
	cConnect := mt.NewInt64Counter(string(nConnectCounter))
	boundCounters["Connect"] = cConnect.Bind(label.String("name", appName))

	nCreateTaskCounter := []rune("TaskScheduler_CreateTask_" + suffix)
	nCreateTaskCounter[0] = unicode.ToLower(nCreateTaskCounter[0])
	cCreateTask := mt.NewInt64Counter(string(nCreateTaskCounter))
	boundCounters["CreateTask"] = cCreateTask.Bind(label.String("name", appName))

	nGenerateQueueIDCounter := []rune("TaskScheduler_GenerateQueueID_" + suffix)
	nGenerateQueueIDCounter[0] = unicode.ToLower(nGenerateQueueIDCounter[0])
	cGenerateQueueID := mt.NewInt64Counter(string(nGenerateQueueIDCounter))
	boundCounters["GenerateQueueID"] = cGenerateQueueID.Bind(label.String("name", appName))

	return boundCounters
}

// Connect implements TaskScheduler
func (_d TaskSchedulerWithTelemetry) Connect() (err error) {
	_since := time.Now()
	defer func() {
		if err != nil {
			errCounter := _d.errCounters["Connect"]
			errCounter.Add(context.Background(), 1)
		}

		methodCounter := _d.methodCounters["Connect"]
		methodCounter.Add(context.Background(), 1)

		methodTimeMeasure := _d.methodTimeValueRecorders["Connect"]
		methodTimeMeasure.Record(context.Background(), time.Since(_since).Milliseconds())
	}()
	return _d.base.Connect()
}

// CreateTask implements TaskScheduler
func (_d TaskSchedulerWithTelemetry) CreateTask(ctx context.Context, req *taskspb.CreateTaskRequest, opts ...gax.CallOption) (tp1 *taskspb.Task, err error) {
	_since := time.Now()
	defer func() {
		if err != nil {
			errCounter := _d.errCounters["CreateTask"]
			errCounter.Add(context.Background(), 1)
		}

		methodCounter := _d.methodCounters["CreateTask"]
		methodCounter.Add(context.Background(), 1)

		methodTimeMeasure := _d.methodTimeValueRecorders["CreateTask"]
		methodTimeMeasure.Record(context.Background(), time.Since(_since).Milliseconds())
	}()
	return _d.base.CreateTask(ctx, req, opts...)
}

// GenerateQueueID implements TaskScheduler
func (_d TaskSchedulerWithTelemetry) GenerateQueueID() (queueID string) {
	_since := time.Now()
	defer func() {

		methodCounter := _d.methodCounters["GenerateQueueID"]
		methodCounter.Add(context.Background(), 1)

		methodTimeMeasure := _d.methodTimeValueRecorders["GenerateQueueID"]
		methodTimeMeasure.Record(context.Background(), time.Since(_since).Milliseconds())
	}()
	return _d.base.GenerateQueueID()
}

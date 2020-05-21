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
	"go.opentelemetry.io/otel/api/key"
	"go.opentelemetry.io/otel/api/metric"
	taskspb "google.golang.org/genproto/googleapis/cloud/tasks/v2beta3"
)

// TaskSchedulerWithTelemetry implements TaskScheduler interface with all methods wrapped
// with open telemetry metrics
type TaskSchedulerWithTelemetry struct {
	base               TaskScheduler
	methodCounters     map[string]metric.BoundInt64Counter
	errCounters        map[string]metric.BoundInt64Counter
	methodTimeMeasures map[string]metric.BoundInt64Measure
}

// NewTaskSchedulerWithTelemetry returns an instance of the TaskScheduler decorated with open telemetry timing and count metrics
func NewTaskSchedulerWithTelemetry(base TaskScheduler, name string, meter metric.Meter) TaskSchedulerWithTelemetry {
	return TaskSchedulerWithTelemetry{
		base:               base,
		methodCounters:     newTaskSchedulerMethodCounters("Calls", name, meter),
		errCounters:        newTaskSchedulerMethodCounters("Errors", name, meter),
		methodTimeMeasures: newTaskSchedulerMethodTimeMeasures(name, meter),
	}
}

func newTaskSchedulerMethodTimeMeasures(appName string, meter metric.Meter) (boundTimeMeasures map[string]metric.BoundInt64Measure) {
	boundTimeMeasures = make(map[string]metric.BoundInt64Measure)

	nConnectMeasure := []rune("TaskScheduler_Connect_ProcessingTimeMillis")
	nConnectMeasure[0] = unicode.ToLower(nConnectMeasure[0])
	mConnect := meter.NewInt64Measure(string(nConnectMeasure), metric.WithKeys(key.New("name")))
	boundTimeMeasures["Connect"] = mConnect.Bind(meter.Labels(key.New("name").String(appName)))

	nCreateTaskMeasure := []rune("TaskScheduler_CreateTask_ProcessingTimeMillis")
	nCreateTaskMeasure[0] = unicode.ToLower(nCreateTaskMeasure[0])
	mCreateTask := meter.NewInt64Measure(string(nCreateTaskMeasure), metric.WithKeys(key.New("name")))
	boundTimeMeasures["CreateTask"] = mCreateTask.Bind(meter.Labels(key.New("name").String(appName)))

	nGenerateQueueIDMeasure := []rune("TaskScheduler_GenerateQueueID_ProcessingTimeMillis")
	nGenerateQueueIDMeasure[0] = unicode.ToLower(nGenerateQueueIDMeasure[0])
	mGenerateQueueID := meter.NewInt64Measure(string(nGenerateQueueIDMeasure), metric.WithKeys(key.New("name")))
	boundTimeMeasures["GenerateQueueID"] = mGenerateQueueID.Bind(meter.Labels(key.New("name").String(appName)))

	return boundTimeMeasures
}

func newTaskSchedulerMethodCounters(suffix string, appName string, meter metric.Meter) (boundCounters map[string]metric.BoundInt64Counter) {
	boundCounters = make(map[string]metric.BoundInt64Counter)

	nConnectCounter := []rune("TaskScheduler_Connect_" + suffix)
	nConnectCounter[0] = unicode.ToLower(nConnectCounter[0])
	cConnect := meter.NewInt64Counter(string(nConnectCounter), metric.WithKeys(key.New("name")))
	boundCounters["Connect"] = cConnect.Bind(meter.Labels(key.New("name").String(appName)))

	nCreateTaskCounter := []rune("TaskScheduler_CreateTask_" + suffix)
	nCreateTaskCounter[0] = unicode.ToLower(nCreateTaskCounter[0])
	cCreateTask := meter.NewInt64Counter(string(nCreateTaskCounter), metric.WithKeys(key.New("name")))
	boundCounters["CreateTask"] = cCreateTask.Bind(meter.Labels(key.New("name").String(appName)))

	nGenerateQueueIDCounter := []rune("TaskScheduler_GenerateQueueID_" + suffix)
	nGenerateQueueIDCounter[0] = unicode.ToLower(nGenerateQueueIDCounter[0])
	cGenerateQueueID := meter.NewInt64Counter(string(nGenerateQueueIDCounter), metric.WithKeys(key.New("name")))
	boundCounters["GenerateQueueID"] = cGenerateQueueID.Bind(meter.Labels(key.New("name").String(appName)))

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

		methodTimeMeasure := _d.methodTimeMeasures["Connect"]
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

		methodTimeMeasure := _d.methodTimeMeasures["CreateTask"]
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

		methodTimeMeasure := _d.methodTimeMeasures["GenerateQueueID"]
		methodTimeMeasure.Record(context.Background(), time.Since(_since).Milliseconds())
	}()
	return _d.base.GenerateQueueID()
}

package stepcurry

// DO NOT EDIT!
// This code is generated with http://github.com/hexdigest/gowrap tool
// using https://raw.githubusercontent.com/alexandre-normand/slackscot/master/opentelemetry.template template

//go:generate gowrap gen -p github.com/alexandre-normand/stepcurry -i Datastorer -t https://raw.githubusercontent.com/alexandre-normand/slackscot/master/opentelemetry.template -o datastorermetrics.go

import (
	"context"
	"time"
	"unicode"

	"cloud.google.com/go/datastore"
	"go.opentelemetry.io/otel/label"
	"go.opentelemetry.io/otel/metric"
)

// DatastorerWithTelemetry implements Datastorer interface with all methods wrapped
// with open telemetry metrics
type DatastorerWithTelemetry struct {
	base                     Datastorer
	methodCounters           map[string]metric.BoundInt64Counter
	errCounters              map[string]metric.BoundInt64Counter
	methodTimeValueRecorders map[string]metric.BoundInt64ValueRecorder
}

// NewDatastorerWithTelemetry returns an instance of the Datastorer decorated with open telemetry timing and count metrics
func NewDatastorerWithTelemetry(base Datastorer, name string, meter metric.Meter) DatastorerWithTelemetry {
	return DatastorerWithTelemetry{
		base:                     base,
		methodCounters:           newDatastorerMethodCounters("Calls", name, meter),
		errCounters:              newDatastorerMethodCounters("Errors", name, meter),
		methodTimeValueRecorders: newDatastorerMethodTimeValueRecorders(name, meter),
	}
}

func newDatastorerMethodTimeValueRecorders(appName string, meter metric.Meter) (boundTimeValueRecorders map[string]metric.BoundInt64ValueRecorder) {
	boundTimeValueRecorders = make(map[string]metric.BoundInt64ValueRecorder)
	mt := metric.Must(meter)

	nCloseValRecorder := []rune("Datastorer_Close_ProcessingTimeMillis")
	nCloseValRecorder[0] = unicode.ToLower(nCloseValRecorder[0])
	mClose := mt.NewInt64ValueRecorder(string(nCloseValRecorder))
	boundTimeValueRecorders["Close"] = mClose.Bind(label.String("name", appName))

	nConnectValRecorder := []rune("Datastorer_Connect_ProcessingTimeMillis")
	nConnectValRecorder[0] = unicode.ToLower(nConnectValRecorder[0])
	mConnect := mt.NewInt64ValueRecorder(string(nConnectValRecorder))
	boundTimeValueRecorders["Connect"] = mConnect.Bind(label.String("name", appName))

	nDeleteValRecorder := []rune("Datastorer_Delete_ProcessingTimeMillis")
	nDeleteValRecorder[0] = unicode.ToLower(nDeleteValRecorder[0])
	mDelete := mt.NewInt64ValueRecorder(string(nDeleteValRecorder))
	boundTimeValueRecorders["Delete"] = mDelete.Bind(label.String("name", appName))

	nGetValRecorder := []rune("Datastorer_Get_ProcessingTimeMillis")
	nGetValRecorder[0] = unicode.ToLower(nGetValRecorder[0])
	mGet := mt.NewInt64ValueRecorder(string(nGetValRecorder))
	boundTimeValueRecorders["Get"] = mGet.Bind(label.String("name", appName))

	nPutValRecorder := []rune("Datastorer_Put_ProcessingTimeMillis")
	nPutValRecorder[0] = unicode.ToLower(nPutValRecorder[0])
	mPut := mt.NewInt64ValueRecorder(string(nPutValRecorder))
	boundTimeValueRecorders["Put"] = mPut.Bind(label.String("name", appName))

	nRunValRecorder := []rune("Datastorer_Run_ProcessingTimeMillis")
	nRunValRecorder[0] = unicode.ToLower(nRunValRecorder[0])
	mRun := mt.NewInt64ValueRecorder(string(nRunValRecorder))
	boundTimeValueRecorders["Run"] = mRun.Bind(label.String("name", appName))

	return boundTimeValueRecorders
}

func newDatastorerMethodCounters(suffix string, appName string, meter metric.Meter) (boundCounters map[string]metric.BoundInt64Counter) {
	boundCounters = make(map[string]metric.BoundInt64Counter)
	mt := metric.Must(meter)

	nCloseCounter := []rune("Datastorer_Close_" + suffix)
	nCloseCounter[0] = unicode.ToLower(nCloseCounter[0])
	cClose := mt.NewInt64Counter(string(nCloseCounter))
	boundCounters["Close"] = cClose.Bind(label.String("name", appName))

	nConnectCounter := []rune("Datastorer_Connect_" + suffix)
	nConnectCounter[0] = unicode.ToLower(nConnectCounter[0])
	cConnect := mt.NewInt64Counter(string(nConnectCounter))
	boundCounters["Connect"] = cConnect.Bind(label.String("name", appName))

	nDeleteCounter := []rune("Datastorer_Delete_" + suffix)
	nDeleteCounter[0] = unicode.ToLower(nDeleteCounter[0])
	cDelete := mt.NewInt64Counter(string(nDeleteCounter))
	boundCounters["Delete"] = cDelete.Bind(label.String("name", appName))

	nGetCounter := []rune("Datastorer_Get_" + suffix)
	nGetCounter[0] = unicode.ToLower(nGetCounter[0])
	cGet := mt.NewInt64Counter(string(nGetCounter))
	boundCounters["Get"] = cGet.Bind(label.String("name", appName))

	nPutCounter := []rune("Datastorer_Put_" + suffix)
	nPutCounter[0] = unicode.ToLower(nPutCounter[0])
	cPut := mt.NewInt64Counter(string(nPutCounter))
	boundCounters["Put"] = cPut.Bind(label.String("name", appName))

	nRunCounter := []rune("Datastorer_Run_" + suffix)
	nRunCounter[0] = unicode.ToLower(nRunCounter[0])
	cRun := mt.NewInt64Counter(string(nRunCounter))
	boundCounters["Run"] = cRun.Bind(label.String("name", appName))

	return boundCounters
}

// Close implements Datastorer
func (_d DatastorerWithTelemetry) Close() (err error) {
	_since := time.Now()
	defer func() {
		if err != nil {
			errCounter := _d.errCounters["Close"]
			errCounter.Add(context.Background(), 1)
		}

		methodCounter := _d.methodCounters["Close"]
		methodCounter.Add(context.Background(), 1)

		methodTimeMeasure := _d.methodTimeValueRecorders["Close"]
		methodTimeMeasure.Record(context.Background(), time.Since(_since).Milliseconds())
	}()
	return _d.base.Close()
}

// Connect implements Datastorer
func (_d DatastorerWithTelemetry) Connect() (err error) {
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

// Delete implements Datastorer
func (_d DatastorerWithTelemetry) Delete(ctx context.Context, k *datastore.Key) (err error) {
	_since := time.Now()
	defer func() {
		if err != nil {
			errCounter := _d.errCounters["Delete"]
			errCounter.Add(context.Background(), 1)
		}

		methodCounter := _d.methodCounters["Delete"]
		methodCounter.Add(context.Background(), 1)

		methodTimeMeasure := _d.methodTimeValueRecorders["Delete"]
		methodTimeMeasure.Record(context.Background(), time.Since(_since).Milliseconds())
	}()
	return _d.base.Delete(ctx, k)
}

// Get implements Datastorer
func (_d DatastorerWithTelemetry) Get(ctx context.Context, k *datastore.Key, dest interface{}) (err error) {
	_since := time.Now()
	defer func() {
		if err != nil {
			errCounter := _d.errCounters["Get"]
			errCounter.Add(context.Background(), 1)
		}

		methodCounter := _d.methodCounters["Get"]
		methodCounter.Add(context.Background(), 1)

		methodTimeMeasure := _d.methodTimeValueRecorders["Get"]
		methodTimeMeasure.Record(context.Background(), time.Since(_since).Milliseconds())
	}()
	return _d.base.Get(ctx, k, dest)
}

// Put implements Datastorer
func (_d DatastorerWithTelemetry) Put(ctx context.Context, k *datastore.Key, v interface{}) (key *datastore.Key, err error) {
	_since := time.Now()
	defer func() {
		if err != nil {
			errCounter := _d.errCounters["Put"]
			errCounter.Add(context.Background(), 1)
		}

		methodCounter := _d.methodCounters["Put"]
		methodCounter.Add(context.Background(), 1)

		methodTimeMeasure := _d.methodTimeValueRecorders["Put"]
		methodTimeMeasure.Record(context.Background(), time.Since(_since).Milliseconds())
	}()
	return _d.base.Put(ctx, k, v)
}

// Run implements Datastorer
func (_d DatastorerWithTelemetry) Run(ctx context.Context, q *datastore.Query) (ip1 *datastore.Iterator) {
	_since := time.Now()
	defer func() {

		methodCounter := _d.methodCounters["Run"]
		methodCounter.Add(context.Background(), 1)

		methodTimeMeasure := _d.methodTimeValueRecorders["Run"]
		methodTimeMeasure.Record(context.Background(), time.Since(_since).Milliseconds())
	}()
	return _d.base.Run(ctx, q)
}

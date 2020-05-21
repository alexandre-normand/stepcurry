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
	"go.opentelemetry.io/otel/api/key"
	"go.opentelemetry.io/otel/api/metric"
)

// DatastorerWithTelemetry implements Datastorer interface with all methods wrapped
// with open telemetry metrics
type DatastorerWithTelemetry struct {
	base               Datastorer
	methodCounters     map[string]metric.BoundInt64Counter
	errCounters        map[string]metric.BoundInt64Counter
	methodTimeMeasures map[string]metric.BoundInt64Measure
}

// NewDatastorerWithTelemetry returns an instance of the Datastorer decorated with open telemetry timing and count metrics
func NewDatastorerWithTelemetry(base Datastorer, name string, meter metric.Meter) DatastorerWithTelemetry {
	return DatastorerWithTelemetry{
		base:               base,
		methodCounters:     newDatastorerMethodCounters("Calls", name, meter),
		errCounters:        newDatastorerMethodCounters("Errors", name, meter),
		methodTimeMeasures: newDatastorerMethodTimeMeasures(name, meter),
	}
}

func newDatastorerMethodTimeMeasures(appName string, meter metric.Meter) (boundTimeMeasures map[string]metric.BoundInt64Measure) {
	boundTimeMeasures = make(map[string]metric.BoundInt64Measure)

	nCloseMeasure := []rune("Datastorer_Close_ProcessingTimeMillis")
	nCloseMeasure[0] = unicode.ToLower(nCloseMeasure[0])
	mClose := meter.NewInt64Measure(string(nCloseMeasure), metric.WithKeys(key.New("name")))
	boundTimeMeasures["Close"] = mClose.Bind(meter.Labels(key.New("name").String(appName)))

	nConnectMeasure := []rune("Datastorer_Connect_ProcessingTimeMillis")
	nConnectMeasure[0] = unicode.ToLower(nConnectMeasure[0])
	mConnect := meter.NewInt64Measure(string(nConnectMeasure), metric.WithKeys(key.New("name")))
	boundTimeMeasures["Connect"] = mConnect.Bind(meter.Labels(key.New("name").String(appName)))

	nDeleteMeasure := []rune("Datastorer_Delete_ProcessingTimeMillis")
	nDeleteMeasure[0] = unicode.ToLower(nDeleteMeasure[0])
	mDelete := meter.NewInt64Measure(string(nDeleteMeasure), metric.WithKeys(key.New("name")))
	boundTimeMeasures["Delete"] = mDelete.Bind(meter.Labels(key.New("name").String(appName)))

	nGetMeasure := []rune("Datastorer_Get_ProcessingTimeMillis")
	nGetMeasure[0] = unicode.ToLower(nGetMeasure[0])
	mGet := meter.NewInt64Measure(string(nGetMeasure), metric.WithKeys(key.New("name")))
	boundTimeMeasures["Get"] = mGet.Bind(meter.Labels(key.New("name").String(appName)))

	nPutMeasure := []rune("Datastorer_Put_ProcessingTimeMillis")
	nPutMeasure[0] = unicode.ToLower(nPutMeasure[0])
	mPut := meter.NewInt64Measure(string(nPutMeasure), metric.WithKeys(key.New("name")))
	boundTimeMeasures["Put"] = mPut.Bind(meter.Labels(key.New("name").String(appName)))

	nRunMeasure := []rune("Datastorer_Run_ProcessingTimeMillis")
	nRunMeasure[0] = unicode.ToLower(nRunMeasure[0])
	mRun := meter.NewInt64Measure(string(nRunMeasure), metric.WithKeys(key.New("name")))
	boundTimeMeasures["Run"] = mRun.Bind(meter.Labels(key.New("name").String(appName)))

	return boundTimeMeasures
}

func newDatastorerMethodCounters(suffix string, appName string, meter metric.Meter) (boundCounters map[string]metric.BoundInt64Counter) {
	boundCounters = make(map[string]metric.BoundInt64Counter)

	nCloseCounter := []rune("Datastorer_Close_" + suffix)
	nCloseCounter[0] = unicode.ToLower(nCloseCounter[0])
	cClose := meter.NewInt64Counter(string(nCloseCounter), metric.WithKeys(key.New("name")))
	boundCounters["Close"] = cClose.Bind(meter.Labels(key.New("name").String(appName)))

	nConnectCounter := []rune("Datastorer_Connect_" + suffix)
	nConnectCounter[0] = unicode.ToLower(nConnectCounter[0])
	cConnect := meter.NewInt64Counter(string(nConnectCounter), metric.WithKeys(key.New("name")))
	boundCounters["Connect"] = cConnect.Bind(meter.Labels(key.New("name").String(appName)))

	nDeleteCounter := []rune("Datastorer_Delete_" + suffix)
	nDeleteCounter[0] = unicode.ToLower(nDeleteCounter[0])
	cDelete := meter.NewInt64Counter(string(nDeleteCounter), metric.WithKeys(key.New("name")))
	boundCounters["Delete"] = cDelete.Bind(meter.Labels(key.New("name").String(appName)))

	nGetCounter := []rune("Datastorer_Get_" + suffix)
	nGetCounter[0] = unicode.ToLower(nGetCounter[0])
	cGet := meter.NewInt64Counter(string(nGetCounter), metric.WithKeys(key.New("name")))
	boundCounters["Get"] = cGet.Bind(meter.Labels(key.New("name").String(appName)))

	nPutCounter := []rune("Datastorer_Put_" + suffix)
	nPutCounter[0] = unicode.ToLower(nPutCounter[0])
	cPut := meter.NewInt64Counter(string(nPutCounter), metric.WithKeys(key.New("name")))
	boundCounters["Put"] = cPut.Bind(meter.Labels(key.New("name").String(appName)))

	nRunCounter := []rune("Datastorer_Run_" + suffix)
	nRunCounter[0] = unicode.ToLower(nRunCounter[0])
	cRun := meter.NewInt64Counter(string(nRunCounter), metric.WithKeys(key.New("name")))
	boundCounters["Run"] = cRun.Bind(meter.Labels(key.New("name").String(appName)))

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

		methodTimeMeasure := _d.methodTimeMeasures["Close"]
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

		methodTimeMeasure := _d.methodTimeMeasures["Connect"]
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

		methodTimeMeasure := _d.methodTimeMeasures["Delete"]
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

		methodTimeMeasure := _d.methodTimeMeasures["Get"]
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

		methodTimeMeasure := _d.methodTimeMeasures["Put"]
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

		methodTimeMeasure := _d.methodTimeMeasures["Run"]
		methodTimeMeasure.Record(context.Background(), time.Since(_since).Milliseconds())
	}()
	return _d.base.Run(ctx, q)
}

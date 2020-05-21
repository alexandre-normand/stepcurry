package stepcurry

// DO NOT EDIT!
// This code is generated with http://github.com/hexdigest/gowrap tool
// using https://raw.githubusercontent.com/alexandre-normand/slackscot/master/opentelemetry.template template

//go:generate gowrap gen -p github.com/alexandre-normand/stepcurry -i Verifier -t https://raw.githubusercontent.com/alexandre-normand/slackscot/master/opentelemetry.template -o verifiermetrics.go

import (
	"context"
	"net/http"
	"time"
	"unicode"

	"go.opentelemetry.io/otel/api/key"
	"go.opentelemetry.io/otel/api/metric"
)

// VerifierWithTelemetry implements Verifier interface with all methods wrapped
// with open telemetry metrics
type VerifierWithTelemetry struct {
	base               Verifier
	methodCounters     map[string]metric.BoundInt64Counter
	errCounters        map[string]metric.BoundInt64Counter
	methodTimeMeasures map[string]metric.BoundInt64Measure
}

// NewVerifierWithTelemetry returns an instance of the Verifier decorated with open telemetry timing and count metrics
func NewVerifierWithTelemetry(base Verifier, name string, meter metric.Meter) VerifierWithTelemetry {
	return VerifierWithTelemetry{
		base:               base,
		methodCounters:     newVerifierMethodCounters("Calls", name, meter),
		errCounters:        newVerifierMethodCounters("Errors", name, meter),
		methodTimeMeasures: newVerifierMethodTimeMeasures(name, meter),
	}
}

func newVerifierMethodTimeMeasures(appName string, meter metric.Meter) (boundTimeMeasures map[string]metric.BoundInt64Measure) {
	boundTimeMeasures = make(map[string]metric.BoundInt64Measure)

	nVerifyMeasure := []rune("Verifier_Verify_ProcessingTimeMillis")
	nVerifyMeasure[0] = unicode.ToLower(nVerifyMeasure[0])
	mVerify := meter.NewInt64Measure(string(nVerifyMeasure), metric.WithKeys(key.New("name")))
	boundTimeMeasures["Verify"] = mVerify.Bind(meter.Labels(key.New("name").String(appName)))

	return boundTimeMeasures
}

func newVerifierMethodCounters(suffix string, appName string, meter metric.Meter) (boundCounters map[string]metric.BoundInt64Counter) {
	boundCounters = make(map[string]metric.BoundInt64Counter)

	nVerifyCounter := []rune("Verifier_Verify_" + suffix)
	nVerifyCounter[0] = unicode.ToLower(nVerifyCounter[0])
	cVerify := meter.NewInt64Counter(string(nVerifyCounter), metric.WithKeys(key.New("name")))
	boundCounters["Verify"] = cVerify.Bind(meter.Labels(key.New("name").String(appName)))

	return boundCounters
}

// Verify implements Verifier
func (_d VerifierWithTelemetry) Verify(header http.Header, body []byte) (err error) {
	_since := time.Now()
	defer func() {
		if err != nil {
			errCounter := _d.errCounters["Verify"]
			errCounter.Add(context.Background(), 1)
		}

		methodCounter := _d.methodCounters["Verify"]
		methodCounter.Add(context.Background(), 1)

		methodTimeMeasure := _d.methodTimeMeasures["Verify"]
		methodTimeMeasure.Record(context.Background(), time.Since(_since).Milliseconds())
	}()
	return _d.base.Verify(header, body)
}

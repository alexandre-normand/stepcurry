package stepcurry

// DO NOT EDIT!
// This code is generated with http://github.com/hexdigest/gowrap tool
// using https://raw.githubusercontent.com/alexandre-normand/slackscot/master/opentelemetry.template template

//go:generate gowrap gen -p github.com/alexandre-normand/stepcurry -i Messenger -t https://raw.githubusercontent.com/alexandre-normand/slackscot/master/opentelemetry.template -o messengermetrics.go

import (
	"context"
	"time"
	"unicode"

	"github.com/nlopes/slack"
	"go.opentelemetry.io/otel/api/key"
	"go.opentelemetry.io/otel/api/metric"
)

// MessengerWithTelemetry implements Messenger interface with all methods wrapped
// with open telemetry metrics
type MessengerWithTelemetry struct {
	base               Messenger
	methodCounters     map[string]metric.BoundInt64Counter
	errCounters        map[string]metric.BoundInt64Counter
	methodTimeMeasures map[string]metric.BoundInt64Measure
}

// NewMessengerWithTelemetry returns an instance of the Messenger decorated with open telemetry timing and count metrics
func NewMessengerWithTelemetry(base Messenger, name string, meter metric.Meter) MessengerWithTelemetry {
	return MessengerWithTelemetry{
		base:               base,
		methodCounters:     newMessengerMethodCounters("Calls", name, meter),
		errCounters:        newMessengerMethodCounters("Errors", name, meter),
		methodTimeMeasures: newMessengerMethodTimeMeasures(name, meter),
	}
}

func newMessengerMethodTimeMeasures(appName string, meter metric.Meter) (boundTimeMeasures map[string]metric.BoundInt64Measure) {
	boundTimeMeasures = make(map[string]metric.BoundInt64Measure)

	nPostMessageMeasure := []rune("Messenger_PostMessage_ProcessingTimeMillis")
	nPostMessageMeasure[0] = unicode.ToLower(nPostMessageMeasure[0])
	mPostMessage := meter.NewInt64Measure(string(nPostMessageMeasure), metric.WithKeys(key.New("name")))
	boundTimeMeasures["PostMessage"] = mPostMessage.Bind(meter.Labels(key.New("name").String(appName)))

	return boundTimeMeasures
}

func newMessengerMethodCounters(suffix string, appName string, meter metric.Meter) (boundCounters map[string]metric.BoundInt64Counter) {
	boundCounters = make(map[string]metric.BoundInt64Counter)

	nPostMessageCounter := []rune("Messenger_PostMessage_" + suffix)
	nPostMessageCounter[0] = unicode.ToLower(nPostMessageCounter[0])
	cPostMessage := meter.NewInt64Counter(string(nPostMessageCounter), metric.WithKeys(key.New("name")))
	boundCounters["PostMessage"] = cPostMessage.Bind(meter.Labels(key.New("name").String(appName)))

	return boundCounters
}

// PostMessage implements Messenger
func (_d MessengerWithTelemetry) PostMessage(channelID string, options ...slack.MsgOption) (channel string, timestamp string, err error) {
	_since := time.Now()
	defer func() {
		if err != nil {
			errCounter := _d.errCounters["PostMessage"]
			errCounter.Add(context.Background(), 1)
		}

		methodCounter := _d.methodCounters["PostMessage"]
		methodCounter.Add(context.Background(), 1)

		methodTimeMeasure := _d.methodTimeMeasures["PostMessage"]
		methodTimeMeasure.Record(context.Background(), time.Since(_since).Milliseconds())
	}()
	return _d.base.PostMessage(channelID, options...)
}

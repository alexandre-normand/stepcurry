package stepcurry

// DO NOT EDIT!
// This code is generated with http://github.com/hexdigest/gowrap tool
// using https://raw.githubusercontent.com/alexandre-normand/slackscot/master/opentelemetry.template template

//go:generate gowrap gen -p github.com/alexandre-normand/stepcurry -i Messenger -t https://raw.githubusercontent.com/alexandre-normand/slackscot/master/opentelemetry.template -o messengermetrics.go

import (
	"context"
	"time"
	"unicode"

	"github.com/slack-go/slack"
	"go.opentelemetry.io/otel/api/kv"
	"go.opentelemetry.io/otel/api/metric"
)

// MessengerWithTelemetry implements Messenger interface with all methods wrapped
// with open telemetry metrics
type MessengerWithTelemetry struct {
	base                     Messenger
	methodCounters           map[string]metric.BoundInt64Counter
	errCounters              map[string]metric.BoundInt64Counter
	methodTimeValueRecorders map[string]metric.BoundInt64ValueRecorder
}

// NewMessengerWithTelemetry returns an instance of the Messenger decorated with open telemetry timing and count metrics
func NewMessengerWithTelemetry(base Messenger, name string, meter metric.Meter) MessengerWithTelemetry {
	return MessengerWithTelemetry{
		base:                     base,
		methodCounters:           newMessengerMethodCounters("Calls", name, meter),
		errCounters:              newMessengerMethodCounters("Errors", name, meter),
		methodTimeValueRecorders: newMessengerMethodTimeValueRecorders(name, meter),
	}
}

func newMessengerMethodTimeValueRecorders(appName string, meter metric.Meter) (boundTimeValueRecorders map[string]metric.BoundInt64ValueRecorder) {
	boundTimeValueRecorders = make(map[string]metric.BoundInt64ValueRecorder)
	mt := metric.Must(meter)

	nPostMessageValRecorder := []rune("Messenger_PostMessage_ProcessingTimeMillis")
	nPostMessageValRecorder[0] = unicode.ToLower(nPostMessageValRecorder[0])
	mPostMessage := mt.NewInt64ValueRecorder(string(nPostMessageValRecorder))
	boundTimeValueRecorders["PostMessage"] = mPostMessage.Bind(kv.Key("name").String(appName))

	return boundTimeValueRecorders
}

func newMessengerMethodCounters(suffix string, appName string, meter metric.Meter) (boundCounters map[string]metric.BoundInt64Counter) {
	boundCounters = make(map[string]metric.BoundInt64Counter)
	mt := metric.Must(meter)

	nPostMessageCounter := []rune("Messenger_PostMessage_" + suffix)
	nPostMessageCounter[0] = unicode.ToLower(nPostMessageCounter[0])
	cPostMessage := mt.NewInt64Counter(string(nPostMessageCounter))
	boundCounters["PostMessage"] = cPostMessage.Bind(kv.Key("name").String(appName))

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

		methodTimeMeasure := _d.methodTimeValueRecorders["PostMessage"]
		methodTimeMeasure.Record(context.Background(), time.Since(_since).Milliseconds())
	}()
	return _d.base.PostMessage(channelID, options...)
}

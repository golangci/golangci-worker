package analytics

import (
	"context"

	"github.com/dukex/mixpanel"
	"github.com/savaki/amplitude-go"
	log "github.com/sirupsen/logrus"
)

type EventName string

const EventPRChecked EventName = "PR checked"

type Tracker interface {
	Track(ctx context.Context, event EventName)
}

type amplitudeMixpanelTracker struct{}

func (t amplitudeMixpanelTracker) Track(ctx context.Context, eventName EventName) {
	ec := ctx.Value(eventName).(eventCollector)
	userID := ec["userIDString"].(string)
	delete(ec, "userIDString")
	eventProps := ec
	log.Infof("track event %s with props %+v", eventName, eventProps)

	ac := getAmplitudeClient()
	if ac != nil {
		ac.Publish(amplitude.Event{
			UserId:          userID,
			EventType:       string(eventName),
			EventProperties: eventProps,
		})
	}

	mp := getMixpanelClient()
	if mp != nil {
		const ip = "0" // don't auto-detect
		mp.Track(userID, string(eventName), &mixpanel.Event{
			IP:         ip,
			Properties: eventProps,
		})
	}
}

func GetTracker(_ context.Context) Tracker {
	return amplitudeMixpanelTracker{}
}

type eventCollector map[string]interface{}

func ContextWithEventPropsCollector(ctx context.Context, name EventName) context.Context {
	return context.WithValue(ctx, name, eventCollector{})
}

func SaveEventProp(ctx context.Context, name EventName, key string, value interface{}) {
	ec := ctx.Value(name).(eventCollector)
	ec[key] = value
}

func SaveEventProps(ctx context.Context, name EventName, props map[string]interface{}) {
	ec := ctx.Value(name).(eventCollector)

	for k, v := range props {
		ec[k] = v
	}
}

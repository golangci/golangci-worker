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
	trackingProps := getTrackingProps(ctx)
	userID := trackingProps["userIDString"].(string)

	eventProps := map[string]interface{}{}
	for k, v := range trackingProps {
		if k != "userIDString" {
			eventProps[k] = v
		}
	}

	addedEventProps := ctx.Value(eventName).(map[string]interface{})
	for k, v := range addedEventProps {
		eventProps[k] = v
	}
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

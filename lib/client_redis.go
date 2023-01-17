package lib

import (
	"encoding/json"
	"errors"
	"fmt"

	"github.com/slack-go/slack"
)

func (c *Client) SendEvent(e ClientEvent) error {
	c.log.Debugw("sending event to redis", "event", e)

	// Marshal the event
	j, err := json.Marshal(e)
	if err != nil {
		return err
	}

	// Grab an event mutex before sending (HA)
	key := fmt.Sprintf("bot_message:%s:%s", e.ChannelUID, e.EventTimestamp)
	mutCmd := c.redis.HSetNX(c.ctx, c.mutexKey, key, "ok")
	if mutCmd == nil {
		return errors.New("failed to set mutex")
	}

	// If it's already been locked, peace out
	if !mutCmd.Val() {
		c.log.Debugw("skipping: event already sent to redis", "event", e)
		return nil
	}

	c.log.Infow("sending event to redis", "type", e.Type, "channel", e.ChannelUID, "user", e.UserUID)

	// Send the event
	return c.redis.RPush(c.ctx, c.eventsKey, string(j)).Err()
}

func (c *Client) SendMessage(m *slack.OutgoingMessage) error {
	m.ID = c.CommandID()
	c.log.Debugw("sending message to slack", "message", m)

	key := fmt.Sprintf("send_slack_message:%d", m.ID)
	mutCmd := c.redis.HSetNX(c.ctx, c.mutexKey, key, "ok")
	if mutCmd == nil {
		return errors.New("failed to set mutex")
	}

	// If it's already been locked, peace out
	if !mutCmd.Val() {
		c.log.Debugw("skipping: message already sent to slack", "message", m)
		return nil
	}

	c.log.Infow("sending message to slack", "type", m.Type, "channel", m.Channel)

	c.rtm.SendMessage(m)
	return nil
}

func (c *Client) SendRawEvent(event interface{}) error {
	return errors.New("not implemented")
}

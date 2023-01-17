package lib

import (
	"encoding/json"
	"errors"
	"fmt"
)

func (c *Client) SendEvent(e ClientEvent) error {
	// Marshal the event
	j, err := json.Marshal(e)
	if err != nil {
		return err
	}

	// Work in a Redis Transaction
	tx := c.redis.TxPipeline()
	defer tx.Close()

	// Grab an event mutex before sending (HA)
	key := fmt.Sprintf("bot_message:%s:%s", e.ChannelUID, e.EventTimestamp)
	mutCmd := tx.HSetNX(c.ctx, c.mutexKey, key, "ok")
	if mutCmd == nil {
		return errors.New("failed to set mutex")
	}

	// If it's already been locked, peace out
	if !mutCmd.Val() {
		return nil
	}

	// Send the event
	return tx.RPush(c.ctx, c.eventsKey, string(j)).Err()
}

func (c *Client) SendRawEvent(event interface{}) error {
	return errors.New("not implemented")
}

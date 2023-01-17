package lib

import (
	"encoding/json"
	"fmt"
	"time"

	"github.com/slack-go/slack"
)

func (c *Controller) RunPubSubLoop() {
	pubsub := c.redis.Subscribe(c.ctx, c.pubsubKey)
	defer pubsub.Close()

	for msg := range pubsub.Channel() {
		cmd := &ControllerCommand{}
		if err := json.Unmarshal([]byte(msg.Payload), cmd); err != nil {
			// log an error
			continue
		}

		if cmd.TeamID == "" {
			continue
		}

		switch cmd.Type {
		case "message":
			c.onMessage(cmd)
		case "team_added":
			c.onTeamAdded(cmd)
		case "team_removed":
			c.onTeamRemoved(cmd)
		}
	}
}

func (c *Controller) onMessage(cmd *ControllerCommand) {
	// get client by index
	client, ok := c.Clients.Load(cmd.ClientIndex())
	if !ok || client == nil {
		return
	}

	if cmd.Id == "" {
		cmd.Id = fmt.Sprintf("%d", time.Now().UnixNano())
	}

	m := &slack.OutgoingMessage{}
	if err := json.Unmarshal([]byte(cmd.Payload), m); err != nil {
		return
	}

	// TODO: mutex-safe send a slack message
	client.rtm.SendMessage(m)
}

func (c *Controller) onTeamAdded(cmd *ControllerCommand) {
	// if the bot already exists, stop it
	c.onTeamRemoved(cmd)

	// initialize client
	config := c.redis.HGet(c.ctx, c.botsKey, cmd.ClientIndex()).Val()
	c.InitializeClient(config)
}

func (c *Controller) onTeamRemoved(cmd *ControllerCommand) {
	// if it doesn't exist in redis, do nothing?

	// load the client
	client, ok := c.Clients.Load(cmd.ClientIndex())
	if !ok || client == nil {
		return
	}

	client.Stop()
	c.Clients.Delete(cmd.ClientIndex())
}

type ControllerCommand struct {
	Id        string `json:"id"`
	Type      string `json:"type"`
	TeamID    string `json:"team_id"`
	UserID    string `json:"user_id"`
	ChannelID string `json:"channel_id"`
	Namespace string `json:"namespace"`
	Payload   string `json:"payload"`
}

func (c ControllerCommand) ClientIndex() string {
	if c.Namespace == "" {
		return c.TeamID
	}

	return fmt.Sprintf("%s-%s", c.Namespace, c.TeamID)
}

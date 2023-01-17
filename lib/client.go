package lib

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/go-redis/redis/v8"
	"github.com/slack-go/slack"
	"github.com/spf13/viper"
)

type Client struct {
	Token     string `json:"token"`
	TeamID    string `json:"team_id"`
	Provider  string `json:"provider"`
	Namespace string `json:"namespace"`

	ctx       context.Context
	redis     *redis.Client
	rawKey    string
	mutexKey  string
	eventsKey string

	slack *slack.Client
	rtm   *slack.RTM
	self  *slack.UserDetails

	stopChan chan bool
}

func NewClient(config string, ctx context.Context, redis *redis.Client, opts ...slack.Option) (*Client, error) {
	c := &Client{
		ctx:       ctx,
		redis:     redis,
		rawKey:    viper.GetString("RELAX_RAW_KEY"),
		mutexKey:  viper.GetString("RELAX_MUTEX_KEY"),
		eventsKey: viper.GetString("RELAX_EVENTS_KEY"),
	}
	if err := json.Unmarshal([]byte(config), c); err != nil {
		return nil, err
	}

	opts = append([]slack.Option{
		slack.OptionDebug(viper.GetString("RELAX_LOG_LEVEL") == "debug"),
		// slack.OptionLog(log.New(os.Stdout, "slack-bot: ", log.Lshortfile|log.LstdFlags)),
	}, opts...)

	c.slack = slack.New(c.Token, opts...)
	c.rtm = c.slack.NewRTM(slack.RTMOptionPingInterval(5 * time.Second))

	return c, nil
}

func (c *Client) Index() string {
	if c.Namespace == "" {
		return c.TeamID
	}

	return fmt.Sprintf("%s-%s", c.Namespace, c.TeamID)
}

func (c *Client) Start() {
	c.stopChan = make(chan bool)

	go c.rtm.ManageConnection()

	for {
		select {
		case msg := <-c.rtm.IncomingEvents:
			c.HandleMessage(msg)
		case <-c.stopChan:
			return
		}
	}
}

func (c *Client) Stop() error {
	if c.stopChan != nil {
		c.stopChan <- true
		c.stopChan = nil
	}

	return c.rtm.Disconnect()
}

func (c *Client) HandleMessage(msg slack.RTMEvent) {
	e := ClientEvent{
		TeamUID:        c.TeamID,
		Namespace:      c.Namespace,
		Provider:       "slack",
		EventTimestamp: fmt.Sprintf("%d", time.Now().UnixNano()),
	}
	if c.self != nil {
		e.RelaxBotUID = c.self.ID
	}

	switch ev := msg.Data.(type) {
	case *slack.HelloEvent:
		c.self = c.rtm.GetInfo().User
	case *slack.MessageEvent:
		switch ev.SubType {
		case slack.MsgSubTypeMessageDeleted:
			e.Type = "message_deleted"
			e.UserUID = ev.User
			e.ChannelUID = ev.Channel
			e.Im = ev.Channel[0] == 'D'
			e.Text = ev.Text
			e.Timestamp = ev.DeletedTimestamp
			e.EventTimestamp = ev.Timestamp
			e.ThreadTimestamp = ev.ThreadTimestamp
		case slack.MsgSubTypeMessageChanged:
			e.Type = "message_edited"
			e.UserUID = ev.User
			e.ChannelUID = ev.Channel
			e.Im = ev.Channel[0] == 'D'
			e.Text = ev.SubMessage.Text
			e.Timestamp = ev.SubMessage.Timestamp
			e.EventTimestamp = ev.Timestamp
			e.ThreadTimestamp = ev.ThreadTimestamp
		case "":
			// Ignore messages sent from the bot itself
			if ev.User == c.self.ID || viper.GetBool("RELAX_SEND_BOT_REPLIES") {
				return
			}

			e.Type = "message_new"
			e.UserUID = ev.User
			e.ChannelUID = ev.Channel
			e.Im = ev.Channel[0] == 'D'
			e.Text = ev.Text
			e.Timestamp = ev.Timestamp
			e.EventTimestamp = ev.Timestamp
			e.ThreadTimestamp = ev.ThreadTimestamp
		}
	case *slack.ReactionAddedEvent:
		e.Type = "reaction_added"
		e.UserUID = ev.User
		e.ChannelUID = ev.Item.Channel
		e.Im = ev.Item.Channel[0] == 'D'
		e.Text = ev.Reaction
		e.Timestamp = ev.Item.Timestamp
		e.EventTimestamp = ev.EventTimestamp
		// e.Attachments
	case *slack.ReactionRemovedEvent:
		e.Type = "reaction_removed"
		e.UserUID = ev.User
		e.ChannelUID = ev.Item.Channel
		e.Im = ev.Item.Channel[0] == 'D'
		e.Text = ev.Reaction
		e.Timestamp = ev.Item.Timestamp
		e.EventTimestamp = ev.EventTimestamp
	case *slack.TeamJoinEvent:
		e.Type = "team_joined"
		e.UserUID = ev.User.ID
		// user is added to meta users
	case *slack.IMCreatedEvent:
		e.Type = "im_created"
		e.UserUID = ev.User
		e.ChannelUID = ev.Channel.ID
		e.Im = true
		// channel is added to meta channels
	case *slack.ChannelJoinedEvent:
		e.Type = "channel_joined"
		e.ChannelUID = ev.Channel.ID
		e.Im = ev.Channel.ID[0] == 'D'
		timestamp := fmt.Sprintf("channel-joined-%d-%s", (time.Now().Unix()/60)*60, ev.Channel.ID)
		e.Timestamp = timestamp
		e.EventTimestamp = timestamp
		e.ThreadTimestamp = timestamp
	default:
		// ignore other events
	}

	if e.Type != "" {
		c.SendEvent(e)
	}
}

func init() {
	viper.SetDefault("RELAX_MUTEX_KEY", "relax_mutex_key")
	viper.SetDefault("RELAX_EVENTS_KEY", "relax_events_key")
	viper.RegisterAlias("RELAX_EVENTS_KEY", "RELAX_EVENTS_QUEUE")
	viper.SetDefault("RELAX_SEND_BOT_REPLIES", false)
}

package lib

import "github.com/slack-go/slack"

type ClientEvent struct {
	Type            string             `json:"type"`
	UserUID         string             `json:"user_uid"`
	ChannelUID      string             `json:"channel_uid"`
	TeamUID         string             `json:"team_uid"`
	Im              bool               `json:"im"`
	Text            string             `json:"text"`
	RelaxBotUID     string             `json:"relax_bot_uid"`
	Timestamp       string             `json:"timestamp"`
	Provider        string             `json:"provider"`
	EventTimestamp  string             `json:"event_timestamp"`
	ThreadTimestamp string             `json:"thread_timestamp"`
	Namespace       string             `json:"namespace"`
	Attachments     []slack.Attachment `json:"attachments"`
}

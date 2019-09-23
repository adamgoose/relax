package slack

import (
	"encoding/json"
	"strings"
	"sync"
	"time"

	"github.com/zerobotlabs/relax/Godeps/_workspace/src/github.com/gorilla/websocket"
	"github.com/zerobotlabs/relax/Godeps/_workspace/src/gopkg.in/redis.v3"
)

// Channel represents a channel in Slack
type Channel struct {
	Id        string `json:"id"`
	Created   int64  `json:"created"`
	Name      string `json:"name"`
	CreatorId string `json:"creator"`

	Im bool
}

type Command struct {
	Id        string `json:"id"`
	Type      string `json:"type"`
	TeamId    string `json:"team_id"`
	UserId    string `json:"user_id"`
	ChannelId string `json:"channel_id"`
	Namespace string `json:"namespace"`
	Payload   string `json:"payload"`
}

type Payload struct {
	Message  string `json:"message"`
	ImageUrl string `json:"image_url"`
}

// Im represents an IM (Direct Message) channel on Slack
type Im struct {
	Id        string `json:"id"`
	Created   int64  `json:"created"`
	CreatorId string `json:"user"`
}

type Field struct {
	Title string `json:"title"`
	Value string `json:"value"`
	Short bool   `json:"short"`
}

type ConfirmAction struct {
	Title       string `json:"title"`
	Text        string `json:"text"`
	OkText      string `json:"ok_text"`
	DismissText string `json:"dismiss_text"`
}

type Action struct {
	Name    string        `json:"name"`
	Text    string        `json:"text"`
	Type    string        `json:"type"`
	Style   string        `json:"style"`
	Value   string        `json:"value"`
	Confirm ConfirmAction `json:"confirm"`
}

type Attachment struct {
	Fallback       string   `json:"fallback"`
	Color          string   `json:"color"`
	Pretext        string   `json:"pretext"`
	AuthorName     string   `json:"author_name"`
	AuthorLink     string   `json:"author_link"`
	AuthorIcon     string   `json:"author_icon"`
	Title          string   `json:"title"`
	TitleLink      string   `json:"title_link"`
	Text           string   `json:"text"`
	Fields         []Field  `json:"fields"`
	ImageUrl       string   `json:"image_url"`
	ThumbUrl       string   `json:"thumb_url"`
	Footer         string   `json:"footer"`
	FooterIcon     string   `json:"footer_icon"`
	Ts             int64    `json:"ts"`
	AttachmentType string   `json:"attachment_type"`
	CallbackId     string   `json:"callback_id"`
	Actions        []Action `json:"actions"`
}

type File struct {
	Id                     string   `json:"id"`
	Created                int64    `json:"created"`
	Timestamp              int64    `json:"timestamp"`
	Name                   string   `json:"name"`
	Title                  string   `json:"title"`
	Mimetype               string   `json:"mimetype"`
	Filetype               string   `json:"filetype"`
	Pretty_type            string   `json:"pretty_type"`
	User                   string   `json:"user"`
	Editable               bool     `json:"editable"`
	Size                   int64    `json:"size"`
	Mode                   string   `json:"mode"`
	IsExternal             bool     `json:"is_external"`
	ExternalType           string   `json:"external_type"`
	IsPublic               bool     `json:"is_public"`
	PublicUrlShared        bool     `json:"public_url_shared"`
	DisplayAsBot           bool     `json:"display_as_bot"`
	username               string   `json:"username"`
	UrlPrivate             string   `json:"url_private"`
	UrlPrivateDownload     string   `json:"url_private_download"`
	Thumb64                string   `json:"thumb_64"`
	Thumb80                string   `json:"thumb_80"`
	Thumb360               string   `json:"thumb_360"`
	Thumb360W              int64    `json:"thumb_360_w"`
	Thumb360H              int64    `json:"thumb_360_h"`
	Thumb480               string   `json:"thumb_480"`
	Thumb480W              int64    `json:"thumb_480_w"`
	Thumb480H              int64    `json:"thumb_480_h"`
	Thumb160               string   `json:"thumb_160"`
	ImageExifRotation      int64    `json:"image_exif_rotation"`
	OriginalW              int64    `json:"original_w"`
	OriginalH              int64    `json:"original_h"`
	Pjpeg                  string   `json:"pjpeg"`
	Permalink              string   `json:"permalink"`
	PermalinkPublic        string   `json:"permalink_public"`
	HasRichPreview         int64    `json:"has_rich_preview"`
}

// Message represents a message on Slack
type Message struct {
	Id               string       `json:"id"`
	Type             string       `json:"type"`
	Subtype          string       `json:"subtype"`
	Text             string       `json:"text"`
	Timestamp        string       `json:"ts"`
	DeletedTimestamp string       `json:"deleted_ts"`
	ThreadTimestamp  string       `json:"thread_ts"`
	Reaction         string       `json:"reaction"`
	Hidden           bool         `json:"hidden"`
	Attachments      []Attachment `json:"attachments"`
	Files            []File       `json:"files"`
	// For some events, such as message_changed, message_deleted, etc.
	// the Timestamp field contains the timestamp of the original message
	// so to make sure only one instance of the event is sent to REDIS_QUEUE_WEB
	// only once, and will be used by the `shouldSendToBot` function
	EventTimestamp string `json:"event_ts"`

	ReplyTo string `json:"reply_to"`
	User    User
	Channel Channel

	RawUser    json.RawMessage `json:"user"`
	RawChannel json.RawMessage `json:"channel"`
	RawMessage json.RawMessage `json:"message"`
	RawItem    json.RawMessage `json:"item"`
}

func (m *Message) UserId() string {
	userId := ""

	userBytes, err := m.RawUser.MarshalJSON()
	if err == nil {
		// since we know it's a string, just remove the extra quotes
		userId = strings.Trim(string(userBytes), "\"")
	}

	return userId
}

func (m *Message) ChannelId() string {
	channelId := ""

	channelBytes, err := m.RawChannel.MarshalJSON()
	if err == nil {
		// since we know it's a string, just remove the extra quotes
		channelId = strings.Trim(string(channelBytes), "\"")
	}

	return channelId
}

func (m *Message) EmbeddedMessage() *Message {
	messageBytes, err := m.RawMessage.MarshalJSON()

	if err == nil {
		var embeddedMessage Message
		err = json.Unmarshal(messageBytes, &embeddedMessage)
		if err == nil {
			return &embeddedMessage
		}
	}

	return nil
}

func (m *Message) EmbeddedItem() *Message {
	messageBytes, err := m.RawItem.MarshalJSON()

	if err == nil {
		var embeddedItem Message
		err = json.Unmarshal(messageBytes, &embeddedItem)
		if err == nil {
			return &embeddedItem
		}
	}

	return nil
}

// Metadata contains data about a Client, such as whether it has been authenticated
// for e.g. Ok == true means a connection has been made, as well as a map
// of channels, users and IMs
type Metadata struct {
	Ok           bool      `json:"ok"`
	Self         User      `json:"self"`
	Url          string    `json:"url"`
	ImsList      []Im      `json:"ims"`
	ChannelsList []Channel `json:"channels"`
	GroupsList   []Channel `json:"groups"`
	UsersList    []User    `json:"users"`
	Error        string    `json:"error"`

	Users    map[string]User
	Channels map[string]Channel
}

// Client is the backbone of this entire project and is used to make connections
// to the Slack API, and send response events back to the user
type Client struct {
	Token            string `json:"token"`
	TeamId           string `json:"team_id"`
	Provider         string `json:"provider"`
	Namespace        string `json:"namespace"`
	heartBeatsMissed int64
	heartBeatsMutex  *sync.Mutex
	data             *Metadata
	conn             *websocket.Conn
	pingTicker       *time.Ticker
	redisClient      *redis.Client
}

// User represents a user on Slack
type User struct {
	Id                  string `json:"id"`
	Name                string `json:"name"`
	Color               string `json:"color"`
	Timezone            string `json:"tz"`
	TimezoneDescription string `json:"tz_label"`
	TimezoneOffset      int64  `json:"tz_offset"`
	IsDeleted           bool   `json:"deleted"`
	IsAdmin             bool   `json:"is_admin"`
	IsBot               bool   `json:"is_bot"`
	IsOwner             bool   `json:"is_owner"`
	IsPrimaryOwner      bool   `json:"is_primary_owner"`
	IsRestricted        bool   `json:"is_restricted"`
}

// Event represents an event that is to be consumed by the user,
// for e.g. when a message is received, an emoji reaction is added, etc.
// an event is sent back to the user.
type Event struct {
	Type            string       `json:"type"`
	UserUid         string       `json:"user_uid"`
	ChannelUid      string       `json:"channel_uid"`
	TeamUid         string       `json:"team_uid"`
	Im              bool         `json:"im"`
	Text            string       `json:"text"`
	RelaxBotUid     string       `json:"relax_bot_uid"`
	Timestamp       string       `json:"timestamp"`
	Provider        string       `json:"provider"`
	EventTimestamp  string       `json:"event_timestamp"`
	ThreadTimestamp string       `json:"thread_timestamp"`
	Namespace       string       `json:"namespace"`
	Attachments     []Attachment `json:"attachments"`
	Files           []File       `json:"files"`
}

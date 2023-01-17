package lib

type ClientEvent struct {
	Type            string                  `json:"type"`
	UserUID         string                  `json:"user_uid"`
	ChannelUID      string                  `json:"channel_uid"`
	TeamUID         string                  `json:"team_uid"`
	Im              bool                    `json:"im"`
	Text            string                  `json:"text"`
	RelaxBotUID     string                  `json:"relax_bot_uid"`
	Timestamp       string                  `json:"timestamp"`
	Provider        string                  `json:"provider"`
	EventTimestamp  string                  `json:"event_timestamp"`
	ThreadTimestamp string                  `json:"thread_timestamp"`
	Namespace       string                  `json:"namespace"`
	Attachments     []ClientEventAttachment `json:"attachments"`
}

type ClientEventAttachment struct {
	Fallback       string                       `json:"fallback"`
	Color          string                       `json:"color"`
	Pretext        string                       `json:"pretext"`
	AuthorName     string                       `json:"author_name"`
	AuthorLink     string                       `json:"author_link"`
	AuthorIcon     string                       `json:"author_icon"`
	Title          string                       `json:"title"`
	TitleLink      string                       `json:"title_link"`
	Text           string                       `json:"text"`
	Fields         []ClientEventAttachmentField `json:"fields"`
	ImageURL       string                       `json:"image_url"`
	ThumbURL       string                       `json:"thumb_url"`
	Footer         string                       `json:"footer"`
	FooterIcon     string                       `json:"footer_icon"`
	TS             int64                        `json:"ts"`
	AttachmentType string                       `json:"attachment_type"`
	CallbackID     string                       `json:"callback_id"`
	Actions        []ClientEventAction          `json:"actions"`
}

type ClientEventAttachmentField struct {
	Title string `json:"title"`
	Value string `json:"value"`
	Short bool   `json:"short"`
}

type ClientEventActionConfirmAction struct {
	Title       string `json:"title"`
	Text        string `json:"text"`
	OKText      string `json:"ok_text"`
	DismissText string `json:"dismiss_text"`
}

type ClientEventAction struct {
	Name    string                         `json:"name"`
	Text    string                         `json:"text"`
	Type    string                         `json:"type"`
	Style   string                         `json:"style"`
	Value   string                         `json:"value"`
	Confirm ClientEventActionConfirmAction `json:"confirm"`
}

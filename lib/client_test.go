package lib

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/go-redis/redismock/v8"
	"github.com/slack-go/slack"
	"github.com/slack-go/slack/slacktest"
	. "github.com/smartystreets/goconvey/convey"
)

func TestClient(t *testing.T) {
	Convey("Test Client", t, func() {
		r, mock := redismock.NewClientMock()

		Convey("Without a namespace", func() {
			st := slacktest.NewTestServer()
			st.SetBotName("Relax")
			go st.Start()
			defer st.Stop()

			c, err := NewClient(ClientJSON, context.Background(), r, log, slack.OptionAPIURL(st.GetAPIURL()))
			c.self = &slack.UserDetails{
				ID: st.BotID,
			}

			So(err, ShouldBeNil)
			So(c.Token, ShouldEqual, Token)
			So(c.TeamID, ShouldEqual, TeamID)
			So(c.slack, ShouldNotBeNil)
			So(c.rtm, ShouldNotBeNil)

			Convey("It calculates an index", func() {
				So(c.Index(), ShouldEqual, TeamID)
			})

			Convey("It notifies us when it starts", func() {
				mock.Regexp().ExpectHSet(c.mutexKey, fmt.Sprintf("bot-%s-started", c.TeamID), `\d+`)

				go c.Start()

				time.Sleep(time.Second)
				So(mock.ExpectationsWereMet(), ShouldBeNil)
			})

			Convey("It responds to Slack events", func() {
				// start the client
				go c.Start()
				defer c.Stop()

				// Triger message_new
				mock.Regexp().ExpectHSetNX(c.mutexKey, `^bot_message:C99:[0-9]+$`, "ok").SetVal(true)
				mock.Regexp().ExpectRPush(c.eventsKey, `"type":"message_new"`).SetVal(1)
				st.SendMessageToChannel("C99", "hey")

				time.Sleep(time.Second)
				So(mock.ExpectationsWereMet(), ShouldBeNil)

				// Trigger message_deleted
				mock.Regexp().ExpectHSetNX(c.mutexKey, `^bot_message:C98:[0-9]+$`, "ok").SetVal(true)
				mock.Regexp().ExpectRPush(c.eventsKey, `"type":"message_deleted"`).SetVal(1)
				st.SendToWebsocket(`{"type":"message","subtype":"message_deleted","channel":"C98","user":"U99","text":"hey","ts":"123456","team":"T99","deleted_ts":"123456"}`)

				time.Sleep(time.Second)
				So(mock.ExpectationsWereMet(), ShouldBeNil)

				// Trigger message_edited
				mock.Regexp().ExpectHSetNX(c.mutexKey, `^bot_message:C97:[0-9]+$`, "ok").SetVal(true)
				mock.Regexp().ExpectRPush(c.eventsKey, `"type":"message_edited"`).SetVal(1)
				st.SendToWebsocket(`{"type":"message","subtype":"message_changed","channel":"C97","user":"U99","ts":"123456","team":"T99","message":{"text":"hey2","ts":"123456"}}`)

				time.Sleep(time.Second)
				So(mock.ExpectationsWereMet(), ShouldBeNil)

				// Trigger reaction_added
				mock.Regexp().ExpectHSetNX(c.mutexKey, `^bot_message:C96:[0-9]+$`, "ok").SetVal(true)
				mock.Regexp().ExpectRPush(c.eventsKey, `"type":"reaction_added"`).SetVal(1)
				st.SendToWebsocket(`{"type":"reaction_added","user":"U99","event_ts":"123456","team":"T99","reaction":"thumbsup","item":{"channel":"C96","ts":"123456"}}`)

				time.Sleep(time.Second)
				So(mock.ExpectationsWereMet(), ShouldBeNil)

				// Trigger reaction_removed
				mock.Regexp().ExpectHSetNX(c.mutexKey, `^bot_message:C95:[0-9]+$`, "ok").SetVal(true)
				mock.Regexp().ExpectRPush(c.eventsKey, `"type":"reaction_removed"`).SetVal(1)
				st.SendToWebsocket(`{"type":"reaction_removed","user":"U99","event_ts":"123456","team":"T99","reaction":"thumbsup","item":{"channel":"C95","ts":"123456"}}`)

				time.Sleep(time.Second)
				So(mock.ExpectationsWereMet(), ShouldBeNil)

				// Trigger team_joined
				mock.Regexp().ExpectHSetNX(c.mutexKey, `^bot_message::[0-9]+$`, "ok").SetVal(true)
				mock.Regexp().ExpectRPush(c.eventsKey, `"type":"team_joined"`).SetVal(1)
				st.SendToWebsocket(`{"type":"team_join","user":{"id":"U99","team_id":"T99"}}`)

				time.Sleep(time.Second)
				So(mock.ExpectationsWereMet(), ShouldBeNil)

				// Trigger im_created
				mock.Regexp().ExpectHSetNX(c.mutexKey, `^bot_message:C94:[0-9]+$`, "ok").SetVal(true)
				mock.Regexp().ExpectRPush(c.eventsKey, `"type":"im_created"`).SetVal(1)
				st.SendToWebsocket(`{"type":"im_created","user":"U99","channel":{"id":"C94"}}`)

				time.Sleep(time.Second)
				So(mock.ExpectationsWereMet(), ShouldBeNil)

				// Trigger channel_joined
				mock.Regexp().ExpectHSetNX(c.mutexKey, `^bot_message:C93:channel-joined-.*$`, "ok").SetVal(true)
				mock.Regexp().ExpectRPush(c.eventsKey, `"type":"channel_joined"`).SetVal(1)
				st.SendToWebsocket(`{"type":"channel_joined","channel":{"id":"C93"}}`)

				time.Sleep(time.Second)
				So(mock.ExpectationsWereMet(), ShouldBeNil)
			})
		})

		Convey("With a namespace", func() {
			c, err := NewClient(NamespacedClientJSON, context.Background(), r, log)
			So(err, ShouldBeNil)

			Convey("It calculates an index", func() {
				So(c.Index(), ShouldEqual, "deadbeef-"+TeamID)
			})
		})

	})
}

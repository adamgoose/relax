package lib

import (
	"context"
	"testing"
	"time"

	"github.com/go-redis/redismock/v8"
	"github.com/slack-go/slack"
	"github.com/slack-go/slack/slacktest"
	. "github.com/smartystreets/goconvey/convey"
	"github.com/spf13/viper"
)

const TeamID = "T01MYRKPPDK"
const Token = "xoxb-9999999999999-9999999999999-xxxxxxxxxxxxxxxxxxxxxxxx"
const ClientJSON = `{"team_id":"T01MYRKPPDK","token":"xoxb-9999999999999-9999999999999-xxxxxxxxxxxxxxxxxxxxxxxx"}`
const NamespacedClientJSON = `{"team_id":"T01MYRKPPDK","token":"xoxb-9999999999999-9999999999999-xxxxxxxxxxxxxxxxxxxxxxxx","namespace":"deadbeef"}`

func TestController(t *testing.T) {
	Convey("Test Controller", t, func() {
		r, mock := redismock.NewClientMock()
		c := NewController(r)

		Convey("It instantiates a new controller", func() {
			So(c, ShouldNotBeNil)
			So(c.ctx, ShouldNotBeNil)
			So(c.botsKey, ShouldEqual, viper.GetString("RELAX_BOTS_KEY"))
		})

		Convey("It can initialize clients", func() {
			So(c.InitializeClient(ClientJSON), ShouldBeNil)

			client, ok := c.Clients.Load(TeamID)
			So(ok, ShouldBeTrue)
			So(client, ShouldNotBeNil)
			So(client.TeamID, ShouldEqual, TeamID)

			So(client.slack, ShouldNotBeNil)
			So(client.rtm, ShouldNotBeNil)
		})

		Convey("It initializes clients from redis", func() {
			mock.ExpectHGetAll(c.botsKey).SetVal(map[string]string{
				TeamID: ClientJSON,
			})

			So(c.InitializeClients(), ShouldBeNil)

			client, ok := c.Clients.Load(TeamID)
			So(ok, ShouldBeTrue)
			So(client, ShouldNotBeNil)
			So(client.TeamID, ShouldEqual, TeamID)

			So(client.slack, ShouldNotBeNil)
			So(client.rtm, ShouldNotBeNil)
		})

		Convey("It can kill clients", func() {
			c.InitializeClient(ClientJSON)

			So(c.KillClient(TeamID), ShouldBeNil)

			_, ok := c.Clients.Load(TeamID)
			So(ok, ShouldBeFalse)
		})

		Convey("It handles a team_added command", func() {
			_, ok := c.Clients.Load(TeamID)
			So(ok, ShouldBeFalse)

			mock.ExpectHGet(c.botsKey, TeamID).SetVal(ClientJSON)
			c.onTeamAdded(&ControllerCommand{
				Type:   "team_added",
				TeamID: TeamID,
			})

			client, ok := c.Clients.Load(TeamID)
			So(ok, ShouldBeTrue)
			So(client, ShouldNotBeNil)
			So(client.TeamID, ShouldEqual, TeamID)
		})

		Convey("It handles a team_removed command", func() {
			c.InitializeClient(ClientJSON)

			_, ok := c.Clients.Load(TeamID)
			So(ok, ShouldBeTrue)

			c.onTeamRemoved(&ControllerCommand{
				Type:   "team_removed",
				TeamID: TeamID,
			})

			_, ok = c.Clients.Load(TeamID)
			So(ok, ShouldBeFalse)
		})

		Convey("It handles a message command", func() {
			st := slacktest.NewTestServer()
			st.SetBotName("relax")
			go st.Start()
			defer st.Stop()

			client, err := NewClient(ClientJSON, context.Background(), r, slack.OptionAPIURL(st.GetAPIURL()))
			So(err, ShouldBeNil)
			c.Clients.Store(TeamID, client)

			mock.ExpectHSetNX(client.mutexKey, "send_slack_message:1", "ok").SetVal(true)

			go client.Start()
			defer client.Stop()

			c.onMessage(&ControllerCommand{
				Id:      "37F29638-60F5-48F2-9996-15C9621F3E35",
				Type:    "message",
				TeamID:  TeamID,
				Payload: `{"id":"37F29638-60F5-48F2-9996-15C9621F3E35","type":"typing","channel":"C99"}`,
			})

			time.Sleep(time.Second)

			So(st.GetSeenInboundMessages(), ShouldContain, `{"id":1,"channel":"C99","type":"typing"}`)
			So(mock.ExpectationsWereMet(), ShouldBeNil)
		})
	})
}

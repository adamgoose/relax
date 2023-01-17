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
		_ = mock

		Convey("Without a namespace", func() {
			st := slacktest.NewTestServer()
			st.SetBotName("Relax")
			stOpt := slack.OptionAPIURL(fmt.Sprintf("http://%s/", st.ServerAddr))

			c, err := NewClient(ClientJSON, context.Background(), r, stOpt)

			So(err, ShouldBeNil)
			So(c.Token, ShouldEqual, Token)
			So(c.TeamID, ShouldEqual, TeamID)
			So(c.slack, ShouldNotBeNil)
			So(c.rtm, ShouldNotBeNil)

			Convey("It calculates an index", func() {
				So(c.Index(), ShouldEqual, TeamID)
			})

		})

		Convey("With a namespace", func() {
			c, err := NewClient(NamespacedClientJSON, context.Background(), r)
			So(err, ShouldBeNil)

			Convey("It calculates an index", func() {
				So(c.Index(), ShouldEqual, "deadbeef-"+TeamID)
			})
		})

	})
}

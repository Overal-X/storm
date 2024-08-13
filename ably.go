package storm

import (
	"context"
	"log"

	"github.com/ably/ably-go/ably"
)

type AblyClient struct {
	ablyRt *ably.Realtime
}

func (c *AblyClient) Connect() {
	c.ablyRt.Connection.OnAll(
		func(change ably.ConnectionStateChange) {
			log.Printf("Connection event: %s state=%s reason=%s", change.Event, change.Current, change.Reason)
		},
	)

	c.ablyRt.Connect()
}

func (c *AblyClient) Disconnect() {
	c.ablyRt.Connection.On(
		ably.ConnectionEventClosed,
		func(change ably.ConnectionStateChange) {
			log.Println("Closed the connection to Ably.")
		},
	)

	c.ablyRt.Close()
}

func (c *AblyClient) Publish(args AblyPublishArgs) error {
	channel := c.ablyRt.Channels.Get(args.Channel)

	return channel.Publish(context.Background(), args.Route, args.Content)
}

func (c *AblyClient) Subscribe(args AblySubscribeArgs) error {
	channel := c.ablyRt.Channels.Get(args.Channel)

	_, err := channel.Subscribe(
		context.Background(),
		args.Route,
		func(msg *ably.Message) { args.Callback(msg.Data) },
	)

	return err
}

func NewAblyClient() *AblyClient {
	env := NewEnv()

	client, err := ably.NewRealtime(ably.WithKey(env.Log.Config.AblyApiKey), ably.WithAutoConnect(false))
	if err != nil {
		log.Fatalln(err)
	}

	return &AblyClient{ablyRt: client}
}

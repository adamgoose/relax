package lib

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/go-redis/redis/v8"
	"github.com/lrita/cmap"
	"github.com/spf13/viper"
	"go.uber.org/zap"
)

type Controller struct {
	Clients cmap.Map[string, *Client]
	ctx     context.Context
	log     *zap.SugaredLogger

	redis     *redis.Client
	botsKey   string
	pubsubKey string
}

func NewController(redis *redis.Client, log *zap.SugaredLogger) *Controller {
	ctx := context.Background()

	log = log.Named("controller")

	return &Controller{
		ctx:       ctx,
		log:       log,
		redis:     redis,
		botsKey:   viper.GetString("RELAX_BOTS_KEY"),
		pubsubKey: viper.GetString("RELAX_BOTS_PUBSUB_KEY"),
	}
}

func (c *Controller) InitializeClients() error {
	c.log.Info("initializing clients")

	result := c.redis.HGetAll(c.ctx, c.botsKey).Val()

	for _, val := range result {
		if err := c.InitializeClient(val); err != nil {
			return err
		}
	}

	return nil
}

func (c *Controller) InitializeClient(j string) error {
	c.log.Debugw("initializing client from config", "config", j)

	client, err := NewClient(j, c.ctx, c.redis, c.log)
	if err != nil {
		return err
	}

	go client.Start()
	c.Clients.Store(client.Index(), client)
	return nil
}

func (c *Controller) KillClient(index string) error {
	client, ok := c.Clients.Load(index)
	if !ok {
		return fmt.Errorf("unable to kill client %s", index)
	}

	client.Stop()
	c.Clients.Delete(index)
	return nil
}

func (c *Controller) Run() error {
	if err := c.InitializeClients(); err != nil {
		return err
	}

	go c.RunPubSubLoop()

	time.Sleep(time.Minute)
	return errors.New(c.botsKey)
}

func init() {
	viper.SetDefault("REDIS_HOST", "localhost:6379")
	viper.SetDefault("RELAX_BOTS_KEY", "relax_bots_key")
	viper.SetDefault("RELAX_BOTS_PUBSUB_KEY", "relax_bots_pubsub_key")
	viper.RegisterAlias("RELAX_BOTS_PUBSUB_KEY", "RELAX_BOTS_PUBSUB")
}

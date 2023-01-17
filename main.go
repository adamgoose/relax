package main

import (
	"context"

	log "github.com/Sirupsen/logrus"
	"github.com/go-redis/redis/v8"
	"github.com/spf13/viper"
	"github.com/zerobotlabs/relax/lib"
)

func main() {
	r := redis.NewClient(&redis.Options{
		Addr:       viper.GetString("REDIS_HOST"),
		Password:   viper.GetString("REDIS_PASSWORD"),
		DB:         0,
		MaxRetries: 5,
	})

	_, err := r.Ping(context.Background()).Result()
	if err != nil {
		log.Fatal(err)
	}

	log.Fatal(lib.NewController(r).Run())
}

func init() {
	viper.AutomaticEnv()
	viper.SetDefault("RELAX_LOG_LEVEL", "debug")

	parsedLevel, err := log.ParseLevel(viper.GetString("RELAX_LOG_LEVEL"))
	if err != nil {
		parsedLevel = log.DebugLevel
	}

	log.SetLevel(parsedLevel)
}

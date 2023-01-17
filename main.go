package main

import (
	"context"
	stdLog "log"

	"github.com/go-redis/redis/v8"
	"github.com/spf13/viper"
	"github.com/zerobotlabs/relax/lib"
	"go.uber.org/zap"
)

func main() {
	defer log.Sync()

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

	lib.NewController(r, log).Run()
}

var log *zap.SugaredLogger

func init() {
	viper.AutomaticEnv()
	viper.SetDefault("RELAX_LOG_LEVEL", "info")

	lc := zap.NewProductionConfig()
	ll, err := zap.ParseAtomicLevel(viper.GetString("RELAX_LOG_LEVEL"))
	if err == nil {
		lc.Level = ll
	}

	l, err := lc.Build()
	if err != nil {
		stdLog.Fatal(err)
	}
	log = l.Named("relax").Sugar()
}

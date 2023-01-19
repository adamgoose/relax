package main

import (
	"context"
	stdLog "log"
	"net/url"

	"github.com/go-redis/redis/v8"
	"github.com/spf13/viper"
	"github.com/zerobotlabs/relax/lib"
	"go.uber.org/zap"
)

func main() {
	defer log.Sync()

	u, err := url.Parse(viper.GetString("REDIS_URL"))
	if err != nil {
		log.Fatalf("failed to parse redis url %s: %v", viper.GetString("REDIS_URL"), err)
	}

	ro := &redis.Options{
		Addr:       u.Host,
		DB:         0,
		MaxRetries: 5,
	}
	if u.User != nil {
		ro.Password, _ = u.User.Password()
	}
	r := redis.NewClient(ro)

	_, err = r.Ping(context.Background()).Result()
	if err != nil {
		log.Fatal(err)
	}

	lib.NewController(r, log).Run()
}

var log *zap.SugaredLogger

func init() {
	viper.AutomaticEnv()
	viper.SetDefault("RELAX_LOG_LEVEL", "info")
	viper.SetDefault("RELAX_LOG_FORMAT", "json")

	var lc zap.Config
	if viper.GetString("RELAX_LOG_FORMAT") == "json" {
		lc = zap.NewProductionConfig()
	} else {
		lc = zap.NewDevelopmentConfig()
	}
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

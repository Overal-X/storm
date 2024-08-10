package main

import (
	"log"

	"github.com/spf13/viper"
)

type Env struct {
	Log struct {
		Method string `yaml:"method"`
		Config struct {
			AblyApiKey string `yaml:"ably_api_key"`

			RedisHost string `yaml:"redis_host"`
			RedisPort string `yaml:"redis_port"`
			RedisUser string `yaml:"redis_user"`
			RedisPass string `yaml:"redis_pass"`
		} `yaml:"config"`
	} `yaml:"log"`
}

func NewEnv() *Env {
	viper.SetConfigName("env.yaml")
	viper.SetConfigType("yaml")

	viper.ReadInConfig()

	var config Env
	if err := viper.Unmarshal(&config); err != nil {
		log.Fatalf("Unable to decode into struct: %v", err)
	}

	return &config
}

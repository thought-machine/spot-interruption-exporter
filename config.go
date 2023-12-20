package main

import (
	"gopkg.in/yaml.v3"
	"os"
)

type PrometheusConfig struct {
	Path string
	Port string
}

type PubSub struct {
	InstanceCreationSubscriptionName     string `yaml:"instance_creation_subscription_name"`
	InstanceInterruptionSubscriptionName string `yaml:"instance_interruption_subscription_name"`
}

type Config struct {
	PubSub      PubSub `yaml:"pubsub"`
	Project     string `yaml:"project_name"`
	ClusterName string `yaml:"cluster_name"`
	LogLevel    string `yaml:"log_level"`
	Prometheus  PrometheusConfig
}

func LoadConfig(path string) (cfg Config, err error) {
	b, err := os.ReadFile(path)
	if err != nil {
		return
	}
	err = yaml.Unmarshal(b, &cfg)
	return
}

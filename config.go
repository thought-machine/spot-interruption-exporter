package main

import (
	"gopkg.in/yaml.v3"
	"os"
)

type GCPConfig struct {
	PubSubSubscriptionName string `yaml:"subscription_name"`
	Project                string `yaml:"project_name"`
}

type PrometheusConfig struct {
	Path string
	Port string
}

type Config struct {
	Provider    string    `yaml:"cloud_provider"`
	ClusterName string    `yaml:"cluster_name"`
	GCP         GCPConfig `yaml:"gcp"`
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

package config

import (
	"time"

	"github.com/spf13/viper"
)

type Config struct {
	WorkerPool WorkerPool `yaml:"worker_pool"`
	HttpServer Http       `yaml:"http_server"`
	Kafka      Kafka      `yaml:"kafka"`
	Redis      Redis      `yaml:"redis"`
	Cassandra  Cassandra  `yaml:"cassandra"`
}

func Init(path string) (*Config, error) {
	viper.SetConfigFile(path)

	var cfg Config

	if err := viper.Unmarshal(&cfg); err != nil {
		return nil, err
	}

	return &cfg, nil
}

type Http struct {
	Host string
	Port int
}
type Kafka struct {
	Brokers  []string
	Username string
	Password string

	ClientID string
	SASL     bool
	TLS      bool
}

type Redis struct {
	Host string
	Port int
	DB   int
}

type Cassandra struct {
	Hosts       []string
	Port        int
	Keyspace    string
	Username    string
	Password    string
	Consistency string
	Timeout     time.Duration
}

type WorkerPool struct {
	NumWorkers int
}

package config

import (
	"time"

	"github.com/spf13/viper"
)

var C *Config

type Config struct {
	Opts       Opts
	Operator   Operator   `yaml:"operator"`
	WorkerPool WorkerPool `yaml:"worker_pool"`
	HttpServer Http       `yaml:"http_server"`
	Kafka      Kafka      `yaml:"kafka"`
	Redis      Redis      `yaml:"redis"`
	Cassandra  Cassandra  `yaml:"cassandra"`
	Postgres   Postgres   `yaml:"postgres"`
}

func Init(path string) error {
	viper.SetConfigFile(path)

	var cfg Config

	if err := viper.Unmarshal(&cfg); err != nil {
		return err
	}

	C = &cfg
	return nil
}

type Http struct {
	Host string `yaml:"host"`
	Port string `yaml:"port"`
}
type Kafka struct {
	Brokers  []string `yaml:"brokers"`
	Username string   `yaml:"username"`
	Password string   `yaml:"password"`

	ClientID string `yaml:"client_id"`
	GroupID  string `yaml:"group_id"`
}

type Redis struct {
	Host string `yaml:"host"`
	Port int    `yaml:"port"`
	DB   int    `yaml:"db"`
}

type Cassandra struct {
	Hosts       []string      `yaml:"hosts"`
	Port        int           `yaml:"port"`
	Keyspace    string        `yaml:"keyspace"`
	Username    string        `yaml:"username"`
	Password    string        `yaml:"password"`
	Consistency string        `yaml:"consistency"`
	Timeout     time.Duration `yaml:"timeout"`
}

type WorkerPool struct {
	NumWorkers int `yaml:"num_workers"`
}

type Opts struct {
	CostPerSms int64 `yaml:"cost_per_sms"`
}

type Operator struct {
	BaseURL string        `yaml:"base_url"`
	APIKey  string        `yaml:"api_key"`
	Timeout time.Duration `yaml:"timeout"`
}

type Postgres struct {
	Host     string `yaml:"host"`
	Port     int    `yaml:"port"`
	Database string `yaml:"database"`
	Username string `yaml:"username"`
	Password string `yaml:"password"`
	SSLMode  string `yaml:"ssl_mode"`
}

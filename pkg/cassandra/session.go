package cassandra

import (
	"fmt"
	"strings"
	"time"

	"github.com/gocql/gocql"
)

type Config struct {
	Hosts       []string
	Port        int
	Keyspace    string
	Username    string
	Password    string
	Consistency string
	Timeout     time.Duration
}

func NewSession(cfg Config) (*gocql.Session, error) {
	if len(cfg.Hosts) == 0 {
		return nil, fmt.Errorf("at least one cassandra host must be provided")
	}

	cluster := gocql.NewCluster(cfg.Hosts...)

	cluster.Port = cfg.Port
	cluster.Keyspace = cfg.Keyspace
	cluster.Timeout = cfg.Timeout

	if cfg.Username != "" {
		cluster.Authenticator = gocql.PasswordAuthenticator{
			Username: cfg.Username,
			Password: cfg.Password,
		}
	}

	cluster.Consistency = parseConsistency(cfg.Consistency)

	session, err := cluster.CreateSession()
	if err != nil {
		return nil, fmt.Errorf("create cassandra session: %w", err)
	}

	return session, nil
}

func parseConsistency(s string) gocql.Consistency {
	switch strings.ToLower(s) {
	case "any":
		return gocql.Any
	case "one":
		return gocql.One
	case "two":
		return gocql.Two
	case "three":
		return gocql.Three
	case "quorum":
		return gocql.Quorum
	case "localquorum":
		return gocql.LocalQuorum
	case "eachquorum":
		return gocql.EachQuorum
	case "all":
		return gocql.All
	case "localone":
		return gocql.LocalOne
	default:
		return gocql.Quorum
	}
}

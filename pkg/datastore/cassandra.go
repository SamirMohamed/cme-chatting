package datastore

import (
	"fmt"
	"github.com/gocql/gocql"
)

type Cassandra struct {
	Session *gocql.Session
}

func NewCassandra(addresses []string, keyspace, username, password string) (*Cassandra, error) {
	cluster := gocql.NewCluster(addresses...)
	cluster.Keyspace = keyspace
	cluster.Consistency = gocql.One
	cluster.Authenticator = gocql.PasswordAuthenticator{
		Username: username,
		Password: password,
	}

	session, err := cluster.CreateSession()
	if err != nil {
		return nil, fmt.Errorf("failed to create Cassandra session: %v", err)
	}

	return &Cassandra{Session: session}, nil
}

func (c *Cassandra) Close() {
	if c.Session != nil {
		c.Session.Close()
	}
}

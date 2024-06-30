package datastore

import (
	"fmt"
	"github.com/gocql/gocql"
)

type Cassandra struct {
	session *gocql.Session
}

func NewCassandra(addresses []string, keyspace, username, password string) (*Cassandra, error) {
	cluster := gocql.NewCluster(addresses...)
	cluster.Keyspace = keyspace
	cluster.Authenticator = gocql.PasswordAuthenticator{
		Username: username,
		Password: password,
	}

	session, err := cluster.CreateSession()
	if err != nil {
		return nil, fmt.Errorf("failed to create Cassandra session: %v", err)
	}

	return &Cassandra{session: session}, nil
}

func (c *Cassandra) Close() {
	if c.session != nil {
		c.session.Close()
	}
}

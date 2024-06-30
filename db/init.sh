#!/bin/bash

# Start Cassandra in the background
cassandra -R &

# Wait for Cassandra to start
until cqlsh -e 'DESC KEYSPACES'; do
  echo "Waiting for Cassandra to start..."
  sleep 2
done

# Import the schema
cqlsh -f /schema.cql

# Keep the container running
tail -f /dev/null

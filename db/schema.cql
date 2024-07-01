CREATE KEYSPACE IF NOT EXISTS cme_chatting WITH replication = { 'class': 'SimpleStrategy', 'replication_factor': 1 };

USE cme_chatting;

CREATE TABLE IF NOT EXISTS users (
   username TEXT PRIMARY KEY,
   password TEXT
);

CREATE TABLE IF NOT EXISTS messages (
  id TIMEUUID,
  sender TEXT,
  recipient TEXT,
  content TEXT,
  timestamp TIMESTAMP,
  PRIMARY KEY ((sender, recipient), id)
) WITH CLUSTERING ORDER BY (id DESC);

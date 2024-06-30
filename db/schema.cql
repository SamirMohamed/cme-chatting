CREATE KEYSPACE IF NOT EXISTS cme_chatting WITH replication = { 'class': 'SimpleStrategy', 'replication_factor': 3 };

USE cme_chatting;

CREATE TABLE IF NOT EXISTS users (
   username TEXT PRIMARY KEY,
   password TEXT
);

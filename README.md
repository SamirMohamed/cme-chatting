# Prerequisites

- Install [docker](https://docs.docker.com/engine/install/) and [docker compose](https://docs.docker.com/compose/install/)

# Setup & Running

## 1. Clone the repo
```bash
git clone https://github.com/SamirMohamed/cme-chatting.git
```
## 2. Go to the directory
```bash
cd cme-chatting
```
## 3. Run docker-compose to up the services
```bash
docker compose up --build -d nginx
```
It will take sometime to the `app` service to be up as `cassandra` service takes sometimes to be up.

## 4. Validate the service is up
After executing the docker-compose command, your server will be up and running with port **8080**.

To Validate if it's running, run in your terminal:
```bash
curl http://localhost:8080/healthcheck
```
You should get ```OK``` in the response

# Requests Format

## Endpoint
```
POST /register
```
## Body
```json
{
  "username": "username",
  "password": "password"
}
```

## Response
```text
HTTP status code 201 // created
```

## Example
Run in your terminal:
```bash
curl -X POST -d '{"username":"username","password":"password"}' http://localhost:8080/register
```

---
## Endpoint
```
POST /login
```
## Body
```json
{
  "username": "username",
  "password": "password"
}
```
## Response
```json
{
  "token": "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJ1c2VybmFtZSI6InNhbWlyIiwiZXhwIjoxNzE5ODMzNjgxfQ.Q3_R6SP4jz8-4-BYG1SHSKfdxCoQpvqpNSI9w0snRmE" // JWT token
}
```

---
## Endpoint
```
POST /send
```
## Header
```json
{
  "Authorization": "Bearer <JWT Token>"
}
```
## Body
```json
{
  "sender": "samir",
  "recipient": "mohamed",
  "content": "Hello!"
}
```
## Response
```text
HTTP status code 201 // created
```

---
## Endpoint
```
GET /messages?sender=<SENDER>&recipient=<RECIPIENT>
```
## Header
```json
{
  "Authorization": "Bearer <JWT Token>"
}
```
## Response
```json
{
  "messages": [
    {
      "id":"d6e84f86-37af-11ef-ad1e-0242ac1d0004",
      "sender":"samir",
      "recipient":"mohamed",
      "content":"Hello!",
      "timestamp":"2024-07-01T13:43:03.612Z"
    },
    ...
  ]
}
```
---
# Architectural decisions and Assumptions
TODO

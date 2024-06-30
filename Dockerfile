FROM golang:1.22-alpine3.20 as build

WORKDIR /app

COPY go.mod go.sum /app/
RUN go mod download

COPY . /app
RUN go build -o chatting .

FROM alpine:3.20
COPY --from=build /app/chatting .
CMD [ "/chatting" ]

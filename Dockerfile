FROM golang:1.22-alpine3.20 as build

WORKDIR /app

COPY go.mod go.sum /app/
RUN go mod download

COPY . /app
RUN go build -C cmd/ -o /app/chatting.bin

FROM alpine:3.20
COPY --from=build /app/chatting.bin .
CMD [ "/chatting.bin" ]

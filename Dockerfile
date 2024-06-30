FROM golang:1.22-alpine3.20

WORKDIR /app
COPY . /app/
RUN go build -o chatting

FROM alpine:3.18
COPY --from=build /app/chatting .
CMD [ "/chatting" ]

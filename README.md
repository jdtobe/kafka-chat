# Kafka Chat

Group chat using Kafka pub-sub as the message exchange.

## TODO

- [x] Implement simple community chat group
- [ ] Add @{name}.{kind}
    - [ ] Add @{name}.chan properties
    - [ ] Add @{name}.user properties
    - [ ] Add @{name}.group properties
- [ ] Add #tag
    - [ ] Add #tag metrics/dashboard
- [ ] And so much more...

## Installation

```go
go install github.com/jdtobe/kafka-chat
```

## Usage

To use `kafka-chat`, you'll need to make sure the docker-compose.yaml file is running (`docker-compose up -d`), then simply execute the application with your username as the first (and only) argument.

`kafka-chat {username}`

> Note: For now, `kafka-chat` always assumes kafka is located @ localhost:9092

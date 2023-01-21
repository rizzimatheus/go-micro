# Microservices in Go

## Architecture

### Services

- **Broker** -  optional single point of entry to microservices
- **Authentication** - PostgreSQL
- **Logger** - MongoDB
- **Mail** - sends emails with a specific template
- **Listener** - consumes messages in RabbitMQ and initiates a process

### Communication

- REST API with JSON as transport
- Sending and Receiving using RPC
- Sending and Receiving using RPC
- Initiating and Responding to events using Advanced Message Queuing Protocol (AMQP)

## [📌] Broker
### Packages Used
- github.com/go-chi/chi/v5
- github.com/go-chi/chi/v5/middleware
- github.com/go-chi/cors

## [❌] Authentication

## [❌] Logger

## [❌] Mail

## [❌] Listener
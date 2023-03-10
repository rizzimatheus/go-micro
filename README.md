# Microservices in Go

## Architecture

### Services

- **Broker** -  optional single point of entry to microservices.
- **Authentication** - user authentication using PostgreSQL.
- **Logger** - event registration using MongoDB.
- **Mail** - sends emails with a specific template.
- **Listener** - consumes messages in RabbitMQ and initiates a process.

### Communication

- REST API with JSON as transport.
- Sending and Receiving using RPC.
- Sending and Receiving using gRPC.
    - Protocol Buffer Compiler: [instalation](https://grpc.io/docs/protoc-installation/) and [quick start](https://grpc.io/docs/languages/go/quickstart/).
- Initiating and Responding to events using Advanced Message Queuing Protocol (AMQP).

## [✔] Broker
Single point of entry to microservices.

### Packages Used
**Routes:**
- github.com/go-chi/chi/v5
- github.com/go-chi/chi/v5/middleware
- github.com/go-chi/cors

**RabbitMQ**
- github.com/rabbitmq/amqp091-go

**gRPC**
- google.golang.org/grpc
- google.golang.org/protobuf

## [✔] Authentication
Service to authenticate users using PostgreSQL database.

### Packages Used
**Routes:**
- github.com/go-chi/chi/v5
- github.com/go-chi/chi/v5/middleware
- github.com/go-chi/cors

**PostgreSQL Connection:**
- github.com/jackc/pgconn
- github.com/jackc/pgx/v4
- github.com/jackc/pgx/v4/stdlib

**gRPC**
- google.golang.org/grpc
- google.golang.org/protobuf

### Request
`http://localhost:8080/handle`

**Body:**
Actions: `auth-json`, `auth-rabbit`, `auth-rpc`, `auth-grpc`
```json
{
    "action": "auth-json",
    "auth": {
        "email": "admin@example.com",
        "password": "verysecret"
    }
}
```

## [✔] Logger
Service for event registration using MongoDB.

### Packages Used
**Routes:**
- github.com/go-chi/chi/v5
- github.com/go-chi/chi/v5/middleware
- github.com/go-chi/cors

**MongoDB Connection:**
- go.mongodb.org/mongo-driver/mongo
- go.mongodb.org/mongo-driver/mongo/options

**gRPC**
- google.golang.org/grpc
- google.golang.org/protobuf

### Request
`http://localhost:8080/handle`

**Body:**
Actions: `log-json`, `log-rabbit`, `log-rpc`, `log-grpc`
```json
{
    "action": "log-json",
    "log": {
        "name": "event",
        "data": "Some kind of data"
    }
}
```

## [✔] Mail
Service to send emails with a specific template.
Broker accepts commands to send mail just for testing purpose. In production this should rejected.

### Packages Used
**Routes:**
- github.com/go-chi/chi/v5
- github.com/go-chi/chi/v5/middleware
- github.com/go-chi/cors

**Mail**
- github.com/vanng822/go-premailer/premailer
- github.com/xhit/go-simple-mail/v2

**gRPC**
- google.golang.org/grpc
- google.golang.org/protobuf

### Request
`http://localhost:8080/handle`

**Body:**
Actions: `mail-json`, `mail-rabbit`, `mail-rpc`, `mail-grpc`
```json
{
    "action": "mail-json",
    "mail": {
        "from": "me@example.com",
        "to": "you@there.com",
        "subject": "Test email",
        "message": "Hello world!"
    }
}
```

## [✔] Listener
Service to consumes messages in RabbitMQ and initiates a process.

### Packages Used
**RabbitMQ**
- github.com/rabbitmq/amqp091-go

## Init Project
### Databases
- Inside `project` folder, run `make init` or create the folder `db-data/postgres`, `db-data/mongo` and `db-data/rabbitmq`.

- Run `make up-build` to start the microservices (requires Docker). Run `make down` if want stop the microservices.

- Connect to the PostgreSQL `users` database on `port 5432`, `user=postgres`, `password=password` and run the SQL:

    ```sql
    --
    -- Name: user_id_seq; Type: SEQUENCE; Schema: public; Owner: postgres
    --
    CREATE SEQUENCE public.user_id_seq
        START WITH 1
        INCREMENT BY 1
        NO MINVALUE
        NO MAXVALUE
        CACHE 1;

    ALTER TABLE public.user_id_seq OWNER TO postgres;

    SET default_tablespace = '';

    SET default_table_access_method = heap;

    --
    -- Name: users; Type: TABLE; Schema: public; Owner: postgres
    --
    CREATE TABLE public.users (
        id integer DEFAULT nextval('public.user_id_seq'::regclass) NOT NULL,
        email character varying(255),
        first_name character varying(255),
        last_name character varying(255),
        password character varying(60),
        user_active integer DEFAULT 0,
        created_at timestamp without time zone,
        updated_at timestamp without time zone
    );

    ALTER TABLE public.users OWNER TO postgres;

    --
    -- Name: user_id_seq; Type: SEQUENCE SET; Schema: public; Owner: postgres
    --
    SELECT pg_catalog.setval('public.user_id_seq', 1, true);

    --
    -- Name: users users_pkey; Type: CONSTRAINT; Schema: public; Owner: postgres
    --
    ALTER TABLE ONLY public.users
        ADD CONSTRAINT users_pkey PRIMARY KEY (id);

    INSERT INTO "public"."users"("email","first_name","last_name","password","user_active","created_at","updated_at")
    VALUES
    (E'admin@example.com',E'Admin',E'User',E'$2a$12$1zGLuYDDNvATh4RA4avbKuheAMpb1svexSzrQm7up.bnpwQHs0jNe',1,E'2022-03-14 00:00:00',E'2022-03-14 00:00:00');
    ```

- Run `make start` to start front-end. Access on `http://localhost/`. Run `make stop` if want stop the front-end.

- Can check logs in the MongoDB. URI to connect: 
    ```
    mongodb://admin:password@localhost:27017/logs?authSource=admin&readPreference=primary&directConnection=true&ssl=false
    ```
# Dark Panda Backend

Dark Panda backend API services. Please refer to [project description](https://gist.github.com/huangc28/8b6c5ff777367597c430a5fd9c6099af). It provides APIs regarding following domains:

- Register by reference code, username and phone verification.
- Accept service inquiry from man, inform girls via socket event.
- Accept chat message, emit message to channel subscriber.
- Check the geolocation of the girl when girl press service start. In order for girl to press start service, she has to be within the radius of specific location.
- TapPay one time payment. Charge man when service is booked.
- TapPay refund when service is failed.
- Features for man and girl to rate each other and give comments.
- Create / cancel / complete a service.

# Spin up docker environment

Docker compose environment consist of the following:

- postgres for development
- postgres for testing
- redis

run docker compose environment

```
docker-compose -f build/package/docker-compose.yaml --env-file build/package/.docker.env up -d
```

# Migrations and model generation

We use [go-migration](https://github.com/golang-migrate/migrate) to manage migrations for this project.

**local postgres DSN**

```
postgres://postgres:1234@127.0.0.1:5432/darkpanda?sslmode=disable
```

**create**

```
migrate create -ext sql -dir db/migrations -seq ${MIGRATION_NAME}
```

**up**

```
migrate -path=db/migrations/ -database 'postgres://postgres:1234@127.0.0.1:5432/darkpanda?sslmode=disable' up
```

**down**

```
migrate -path=db/migrations/ -database 'postgres://postgres:1234@127.0.0.1:5432/darkpanda?sslmode=disable' up
```

Run the following command to generate models in go code from migration contents:

```
go run cmd/genmodel/main.go
```

It reads and collect migration SQL from `db/migrations`. The collected content will be written to `db/schema.sql`. Moreover, it generate go code via [sqlc](https://github.com/kyleconroy/sqlc) based on SQL in `db/migrations/schema.sql`.

The content in `db/migrations/schema.sql` will be truncated everytime running the above command to ensure the model is always up to date with the latest SQL of migration **up** files.

# TODOs

- docker environment with postgres and redis
- dotenv files

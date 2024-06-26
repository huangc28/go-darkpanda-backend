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

## Environment variables

Please make sure your environment variable is setup properly before proceeding any development.
Create `env.toml` in the project directory. Paste the following content in file.

[env.toml](https://gist.github.com/huangc28/0ffa71dffefc462728e602d0919cf9bd)

## Install docker

Please make sure you have installed [docker](https://www.docker.com/get-started) on your machine

## Run on local

run docker compose for local environment

```
make run_local

```

You can then connect to local postgres and redis.

**Local postgres**

```
host     = "127.0.0.1"
port     = 5432
user     = "postgres"
password = "1234"
dbname   = "darkpanda"
```

**Local test postgres**

```
host     = "127.0.0.1"
port     = 5433
user     = "postgres"
password = "1234"
dbname   = "darkpanda"
```

**Local Redis**

```
addr = "localhost:6379"
password = ""
DB = 0
```

# Migrations and model generation

We use [go-migration](https://github.com/golang-migrate/migrate) to manage migrations for this project.

When first time host up this project, please run the migration first before proceed development.

Prompt `make migrate_up` to run all migration schemes.

**up**

```
make migrate_up
```

**down**

```
make migrate_down
```
**create**

```
make migrate_create {MIGRATION_NAME}
```

## Generate models from SQL migration

Every time you apply a migration, the data model in golang should be generated again. Run the following command to generate models in go code from migration contents:

```
make gen_model
```

It reads and collect migration SQL from `db/migrations`. The collected content will be written to `db/schema.sql`. Moreover, it generate go code via [sqlc](https://github.com/kyleconroy/sqlc) based on SQL in `db/migrations/schema.sql`.

The content in `db/migrations/schema.sql` will be truncated everytime running the above command to ensure the model is always up to date with the latest SQL of migration **up** files.

## Supported Languages

The app currently supports:
- zh
- en

You can lookup languages here: www.science.co.il/language/Locale-codes.php
## TODOs

- [] Notify male user when is at service appointment time.
- [] Remove budget column from services table.

# Miscellaneous

**local postgres DSN**

```
postgres://postgres:1234@127.0.0.1:5432/darkpanda?sslmode=disable
```

**Add TLS instructions**

```
https://www.digitalocean.com/community/tutorials/how-to-secure-nginx-with-let-s-encrypt-on-ubuntu-20-04
```

## Github Actions


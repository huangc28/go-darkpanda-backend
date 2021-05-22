# Export variable in .app.env.
ifneq (,$(wildcard ./.app.env))
	include .app.env
	export
endif


# Spin up project on local.
run_local: run_local_backend
	echo 'run successfully on local'

run_local_backend: run_local_docker
	go mod tidy && go run cmd/app/main.go

run_local_docker:
	docker-compose \
		-f build/package/docker-compose.yaml \
		--env-file build/package/.env.dev up \
		-d

# ---------- Deprecated ----------
#serve_swagger: build_swagger
	#swagger serve swagger/master.yml -p 3333 --host localhost --flavor=swagger

#build_swagger:
	#swagger flatten swagger/general.yml \
	#--output=swagger/master.yml \
	#--format=yaml
# --------------------------------

.PHONY: docker_compose
.PHONY: serve_swagger
.PHONY: build_swagger

# Generate models from migration SQL schemas. This tool uses
# `https://github.com/kyleconroy/sqlc` to parse SQL syntax
# and generate corresponding models.
gen_model:
	go run cmd/genmodel/main.go gen

MIGRATE_CMD=migrate
MIGRATE_CREATE_CMD=create
MIGRATE_UP_CMD=up
MIGRATE_UP_CMD=down

PG_DSN=postgres://$(PG_USER):$(PG_PASSWORD)@$(PG_HOST):$(PG_PORT)/darkpanda?sslmode=disable
PG_TEST_DSN=postgres://$(TEST_PG_USER):$(TEST_PG_PASSWORD)@$(TEST_PG_HOST):$(TEST_PG_PORT)/darkpanda?sslmode=disable

# Create a new migration file.
# Usage:
#   migrate_create referral_code.
migrate_create:
	$(MIGRATE_CMD) $(MIGRATE_CREATE_CMD) -ext sql -dir db/migrations -seq $(filter-out $@, $(MAKECMDGOALS))

migrate_up:
	$(MIGRATE_CMD) -path=db/migrations/ -database $(PG_DEV_DSN) $(MIGRATE_UP_CMD) && make gen_model

migrate_down:
	$(MIGRATE_CMD) -path=db/migrations/ -database $(PG_DEV_DSN) $(MIGRATE_UP_CMD)

test_migrate_up:
	ENV=test $(MIGRATE_CMD) -path=db/migrations/ -database $(PG_TEST_DSN) $(MIGRATE_UP_CMD) && make gen_model

test_migrate_down:
	ENV=test $(MIGRATE_CMD) -path=db/migrations/ -database $(PG_TEST_DSN) $(MIGRATE_UP_CMD)

prod_migrate_up:
	$(MIGRATE_CMD) -path=db/migrations/ -database $(PG_PROD_DSN) $(MIGRATE_UP_CMD)


# Build & Deploy
deploy: build
	ssh -t root@hookie.club 'cd ~/darkpanda/go-darkpanda-backend && \
		git pull https://$(GITHUB_USER):$(GITHUB_ACCESS_TOKEN)@github.com/huangc28/go-darkpanda-backend.git && \
		git reset --hard && \
		make build && \
		sudo systemctl stop darkpanda.service && \
		sudo systemctl start darkpanda.service'

build:
	echo 'building production binary...';
	cd cmd/app && GOOS=linux GOARCH=amd64 go build -o ../../bin/darkpanda_backend -v .

.PHONY: build
.PHONY: deploy

CURRENT_DIR = $(shell pwd)

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
MIGRATE_DOWN_CMD=down

PG_DSN=postgres://$(PG_USER):$(PG_PASSWORD)@$(PG_HOST):$(PG_PORT)/darkpanda?sslmode=disable
PG_TEST_DSN=postgres://$(TEST_PG_USER):$(TEST_PG_PASSWORD)@$(TEST_PG_HOST):$(TEST_PG_PORT)/darkpanda?sslmode=disable

# Create a new migration file.
# Usage:
#   migrate_create referral_code.
#
# Migrate down:
#    migrate -path=db/migrations/ -database 'postgres://darkpanda:1234@127.0.0.1:5432/darkpanda?sslmode=disable' down 1

migrate_create:
	$(MIGRATE_CMD) $(MIGRATE_CREATE_CMD) -ext sql -dir db/migrations -seq $(filter-out $@, $(MAKECMDGOALS))

migrate_up:
	$(MIGRATE_CMD) -path=${CURRENT_DIR}/db/migrations/ -database $(PG_DSN) $(MIGRATE_UP_CMD) && make gen_model

migrate_down:
	$(MIGRATE_CMD) -path=${CURRENT_DIR}/db/migrations/ -database $(PG_DSN) $(MIGRATE_DOWN_CMD) $(filter-out $@, $(MAKECMDGOALS))

test_migrate_up:
	ENV=test $(MIGRATE_CMD) -path=${CURRENT_DIR}/db/migrations/ -database $(PG_TEST_DSN) $(MIGRATE_UP_CMD) && make gen_model

test_migrate_down:
	ENV=test $(MIGRATE_CMD) -path=${CURRENT_DIR}/db/migrations/ -database $(PG_TEST_DSN) $(MIGRATE_DOWN_CMD)


# Build & Deploy
APP_SERVICE_NAME                    = darkpanda.service
SERVICE_STATUS_SCANNER_SERVICE_NAME = darkpanda_service_status_scanner.service

deploy: build
	ssh -t root@hookie.club 'cd ~/darkpanda/go-darkpanda-backend && \
		git pull https://$(GITHUB_USER):$(GITHUB_ACCESS_TOKEN)@github.com/huangc28/go-darkpanda-backend.git && \
		make build && \
		sudo systemctl stop $(APP_SERVICE_NAME) && \
		sudo systemctl start $(APP_SERVICE_NAME) && \
		sudo systemctl stop $(SERVICE_STATUS_SCANNER_SERVICE_NAME) && \
		TICK_INTERVAL_IN_SECOND=60 sudo systemctl start $(SERVICE_STATUS_SCANNER_SERVICE_NAME)'

build: build_service_status_scanner
	echo 'building production binary...'
	cd $(CURRENT_DIR)/cmd/app && GOOS=linux GOARCH=amd64 go build -o ../../bin/darkpanda_backend -v .

build_service_status_scanner:
	echo 'building worker binary...'
	cd $(CURRENT_DIR)/cmd/workers/service_status_scanner && GOOS=linux GOARCH=amd64 go build -o ../../../bin/service_status_scanner -v .

.PHONY: build
.PHONY: deploy

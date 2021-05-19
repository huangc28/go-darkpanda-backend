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
serve_swagger: build_swagger
	swagger serve swagger/master.yml -p 3333 --host localhost --flavor=swagger

build_swagger:
	swagger flatten swagger/general.yml \
	--output=swagger/master.yml \
	--format=yaml
# --------------------------------

.PHONY: docker_compose
.PHONY: serve_swagger
.PHONY: build_swagger

# Generate models from migration SQL schemas. This tool uses
# `https://github.com/kyleconroy/sqlc` to parse SQL syntax
# and generate corresponding models.
gen_model:
	go run cmd/genmodel/main.go gen

# Create a new migration file.
# Usage:
#   migrate_create referral_code.
migrate_create:
	migrate create -ext sql -dir db/migrations -seq $(filter-out $@, $(MAKECMDGOALS))

migrate_up:
	migrate -path=db/migrations/ -database 'postgres://postgres:1234@127.0.0.1:5432/darkpanda?sslmode=disable' up && make gen_model

migrate_down:
	migrate -path=db/migrations/ -database 'postgres://postgres:1234@127.0.0.1:5432/darkpanda?sslmode=disable' down

test_migrate_up:
	ENV=test migrate -path=db/migrations/ -database 'postgres://postgres:1234@127.0.0.1:5433/darkpanda?sslmode=disable' up && make gen_model

test_migrate_down:
	ENV=test migrate -path=db/migrations/ -database 'postgres://postgres:1234@127.0.0.1:5433/darkpanda?sslmode=disable' down

# Build production
build_prod:
	cd cmd/app; ENV=production go build -o ../../bin/darkpanda_backend -v .

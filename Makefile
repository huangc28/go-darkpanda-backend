# Spin up project on local.
run_local: run_local_backend
	echo 'run successfully on local'

run_local_backend: run_local_docker
	go run cmd/app/main.go

run_local_docker:
	docker-compose \
		-f build/package/docker-compose.yaml \
		--env-file build/package/.docker.env up \
		-d
docker_compose:
	docker-compose \
		-f build/package/docker-compose.yaml \
		--env-file build/package/.docker.env \
		$(filter-out $@, $(MAKECMDGOALS))

.PHONY: docker_compose

# Create a new migration file.
# Usage:
#   migrate_create referral_code.
migrate_create:
	migrate create -ext sql -dir db/migrations -seq $(filter-out $@, $(MAKECMDGOALS))

migrate_up:
	migrate -path=db/migrations/ -database 'postgres://postgres:1234@127.0.0.1:5432/darkpanda?sslmode=disable' up

migrate_down:
	migrate -path=db/migrations/ -database 'postgres://postgres:1234@127.0.0.1:5432/darkpanda?sslmode=disable' down


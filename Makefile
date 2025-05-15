-include .env

DB_URL = postgres://$(DB_USER):$(DB_PASSWORD)@$(DB_HOST):$(DB_PORT)/$(DB_NAME)?sslmode=$(DB_SSLMODE)

.PHONY: sqlc.generate
sqlc.generate:
	sqlc generate

.PHONY: sqlc.vet
sqlc.vet:
	sqlc vet

.PHONY: db.up
db.up:
	docker run --name $(CONTAINER_NAME) -p $(DB_PORT):5432 \
		-e POSTGRES_USER=$(DB_USER) \
		-e POSTGRES_PASSWORD=$(DB_PASSWORD) \
		-e POSTGRES_DB=$(DB_NAME) \
		-d postgres:latest

.PHONY: db.stop
db.stop:
	docker stop $(CONTAINER_NAME)

.PHONY: db.down
db.down: db.stop
	docker rm $(CONTAINER_NAME)

.PHONY: migrate.up
migrate.up:
	migrate -path $(MIGRATIONS_DIR) -database "$(DB_URL)" up

.PHONY: migrate.down
migrate.down:
	migrate -path $(MIGRATIONS_DIR) -database "$(DB_URL)" down 1

.PHONY: migrate.create
migrate.create:
	@read -p "Enter migration name: " name; \
	migrate create -ext sql -dir $(MIGRATIONS_DIR) -seq $${name}

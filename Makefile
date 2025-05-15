-include .env

DB_URL=postgres://$(DB_USER):$(DB_PASSWORD)@$(DB_HOST):$(DB_PORT)/$(DB_NAME)?sslmode=$(DB_SSLMODE)
DOCKER_DB_URL=postgres://$(DB_USER):$(DB_PASSWORD)@localhost:$(DB_PORT)/$(DB_NAME)?sslmode=$(DB_SSLMODE)

.PHONY: db.up
db.up:
	@echo "Starting PostgreSQL container..."
	@docker run --name $(CONTAINER_NAME) -p $(DB_PORT):5432 \
		-e POSTGRES_USER=$(DB_USER) \
		-e POSTGRES_PASSWORD=$(DB_PASSWORD) \
		-e POSTGRES_DB=$(DB_NAME) \
		-d postgres:latest
	@echo "Container started. Use 'make db.psql' to connect"

.PHONY: db.down
db.down:
	@echo "Stopping PostgreSQL container..."
	@docker stop $(CONTAINER_NAME)
	@docker rm $(CONTAINER_NAME)
	@echo "Container stopped and removed"

.PHONY: db.psql
db.psql:
	@docker exec -it $(CONTAINER_NAME) psql -U $(DB_USER) -d $(DB_NAME)

.PHONY: migrate.create 
migrate.create:
	@migrate create -ext sql -dir $(MIGRATIONS_DIR) -seq $(filter-out $@,$(MAKECMDGOALS))

.PHONY: migrate.up
migrate.up:
	@migrate -path $(MIGRATIONS_DIR) -database "$(DB_URL)" up $(if $(filter-out $@,$(MAKECMDGOALS)),$(filter-out $@,$(MAKECMDGOALS)),)

.PHONY: migrate.down
migrate.down:
	@migrate -path $(MIGRATIONS_DIR) -database "$(DB_URL)" down $(if $(filter-out $@,$(MAKECMDGOALS)),$(filter-out $@,$(MAKECMDGOALS)),)

.PHONY: migrate.force
migrate.force:
	@migrate -path $(MIGRATIONS_DIR) -database "$(DB_URL)" force $(filter-out $@,$(MAKECMDGOALS))

.PHONY: migrate.version
migrate.version:
	@migrate -path $(MIGRATIONS_DIR) -database "$(DB_URL)" version

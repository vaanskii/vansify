# Makefile for managing migrations

MIGRATE=migrate
MIGRATIONS_DIR=migrations

# Load environment variables from .env file
ifneq ("$(wildcard .env)","")
include .env
export
endif

DATABASE_URL="mysql://$(DB_USER):$(DB_PASS)@tcp($(DB_HOST):$(DB_PORT))/$(DB_NAME)?parseTime=true"

.PHONY: migrate up down

migrate:
	@echo "Running migrations..."
	$(MIGRATE) -path $(MIGRATIONS_DIR) -database $(DATABASE_URL) up

up:
	@echo "Applying up migrations..."
	$(MIGRATE) -path $(MIGRATIONS_DIR) -database $(DATABASE_URL) up

down:
	@echo "Rolling back migrations..."
	$(MIGRATE) -path $(MIGRATIONS_DIR) -database $(DATABASE_URL) down

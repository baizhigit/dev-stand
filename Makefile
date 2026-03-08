.PHONY: up-all up-all-d down-all build-all rebuild scale

up-all:
	docker compose up --remove-orphans

up-all-d:
	docker compose up --remove-orphans -d

down-all:
	docker compose down -v

build-all:
	docker compose build

# Rebuild and restart one service without touching others
# Usage: make rebuild SERVICE=api-gateway
rebuild:
	@if [ -z "$(SERVICE)" ]; then \
		echo "Error: SERVICE parameter is required"; \
		echo "Usage: make rebuild SERVICE=<name>"; \
		echo "Available: $$(docker compose config --services)"; \
		exit 1; \
	fi
	docker compose build $(SERVICE)
	docker compose up --remove-orphans -d $(SERVICE)

# Local load testing with multiple replicas
# Usage: make scale SERVICE=api-gateway N=3
scale:
	@if [ -z "$(SERVICE)" ] || [ -z "$(N)" ]; then \
		echo "Usage: make scale SERVICE=<name> N=<count>"; \
		exit 1; \
	fi
	docker compose up --scale $(SERVICE)=$(N) -d
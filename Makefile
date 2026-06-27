include .env
export 

export PROJECT_ROOT=${shell pwd}

env-up:
	@docker compose up -d trackingapp-postgres

env-down:
	@docker compose down trackingapp-postgres

env-cleanup:
	@read -p "Очистить все volume файлы окружения? [y/N]: " ans; \
	if [ "$$ans" = "y" ]; then \
		docker compose down trackingapp-postgres && \
		rm -rf out/pgdata && \
		echo "Файлы окружения очищены"; \
	else \
		echo "Очистка окружения отменена"; \
	fi

env-port-forwarder:
	@docker compose up -d port-forwarder 

env-port-close:
	@docker compose down -d port-forwarder

migrate-create:
	@if [ -z "$(seq)" ]; then \
		echo "Отсутствует необходимый параметр seq"; \
		exit 1; \
	fi; \
	docker compose run --rm trackingapp-postgres-migrate \
		create \
		-ext sql \
		-dir /migrations \
		-seq "$(seq)"

migrate-up:
	@make migrate-action action=up

migrate-down:
	@make migrate-action action=down

migrate-action:
	@if [ -z "$(action)" ]; then \
		echo "Отсутствует необходимый параметр action"; \
		exit 1; \
	fi; \
	docker compose run --rm trackingapp-postgres-migrate \
		-path /migrations \
		-database postgres://${POSTGRES_USER}:${POSTGRES_PASSWORD}@trackingapp-postgres:5432/${POSTGRES_DB}?sslmode=disable \
		"$(action)"

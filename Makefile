.PHONY: up down reset test

up:
	go -C backend mod tidy
	docker compose up --build

down:
	docker compose down

reset:
	docker compose down -v

test:
	go -C backend test -race ./...
	npm --prefix frontend test
	npm --prefix frontend run build
	docker compose up -d --wait db
	docker compose exec -T db sh -c 'dropdb -U "$$POSTGRES_USER" --if-exists --force job_search_tracker_test && createdb -U "$$POSTGRES_USER" job_search_tracker_test'
	docker compose --profile tools run --build --rm --no-deps python-tools sh -c 'DATABASE_URL="$$TEST_DATABASE_URL" python -m alembic upgrade head && python -m pytest -q'

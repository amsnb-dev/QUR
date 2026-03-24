.PHONY: migrate up up-migrate down logs ps

migrate:
@echo "▶ starting postgres..."
docker compose up -d postgres
@echo "▶ waiting for postgres to be healthy..."
@until docker compose exec postgres pg_isready -U quran -d quran_school > /dev/null 2>&1; do sleep 1; done
@echo "▶ running migrations..."
RUN_MIGRATIONS=true docker compose up --build -d api
@echo "▶ waiting for migrations to complete (watching logs)..."
@until docker compose logs --no-color --tail=200 api 2>/dev/null | grep -q "✅ migrations applied"; do sleep 1; done
@echo "▶ stopping api (migrations done)..."
docker compose stop api
@echo "✅ migrations complete — run 'make up' to start the full stack"

up:
docker compose up --build

up-migrate:
RUN_MIGRATIONS=true docker compose up --build

down:
docker compose down

logs:
docker compose logs -f api

ps:
docker compose ps

.PHONY: migrate up up-migrate down logs ps

## تشغيل migrations عبر Docker (مرة واحدة عند الإعداد الأولي أو عند إضافة migrations جديدة).
## يشغّل postgres إذا لم يكن شغّالاً، ثم يشغّل api مع RUN_MIGRATIONS=true ثم يوقفه.
migrate:
	@echo "▶ starting postgres..."
	docker compose up -d postgres
	@echo "▶ waiting for postgres to be healthy..."
	@until docker compose exec postgres pg_isready -U quran -d quran_school > /dev/null 2>&1; do sleep 1; done
	@echo "▶ running migrations..."
	RUN_MIGRATIONS=true docker compose up --build -d api
	@echo "▶ waiting for migrations to complete..."
	@sleep 5
	@echo "▶ stopping api (migrations done)..."
	docker compose stop api
	@echo "✅ migrations complete — run 'make up' to start the full stack"

## تشغيل الـ stack كاملاً — migrations=false (الافتراضي)
up:
	docker compose up --build

## تشغيل الـ stack مع migrations مفعّلة (للتشغيل الأول أو التحديثات)
up-migrate:
	RUN_MIGRATIONS=true docker compose up --build

## إيقاف جميع الخدمات
down:
	docker compose down

## عرض logs
logs:
	docker compose logs -f api

## حالة الـ containers
ps:
	docker compose ps

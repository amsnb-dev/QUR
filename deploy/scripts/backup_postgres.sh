#!/usr/bin/env bash
# ============================================================
#  deploy/scripts/backup_postgres.sh
#  pg_dump داخل container + ضغط gzip + احتفاظ 14 يوم
#
#  الاستخدام:
#    ./deploy/scripts/backup_postgres.sh
#
#  المتغيرات (من .env.prod أو environment):
#    POSTGRES_USER      (افتراضي: quran)
#    POSTGRES_DB        (افتراضي: quran_school)
#    BACKUP_DIR         (افتراضي: /opt/quran/backups)
#    RETENTION_DAYS     (افتراضي: 14)
#    CONTAINER_NAME     (افتراضي: quran_db_prod)
# ============================================================
set -euo pipefail

# ── تحميل .env.prod إذا وُجد ─────────────────────────────────
SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
ENV_FILE="$SCRIPT_DIR/../.env.prod"
if [[ -f "$ENV_FILE" ]]; then
    # shellcheck disable=SC1090
    set -a; source "$ENV_FILE"; set +a
fi

# ── متغيرات ───────────────────────────────────────────────────
PG_USER="${POSTGRES_USER:-quran}"
PG_DB="${POSTGRES_DB:-quran_school}"
BACKUP_DIR="${BACKUP_DIR:-/opt/quran/backups}"
RETENTION_DAYS="${RETENTION_DAYS:-14}"
CONTAINER="${CONTAINER_NAME:-quran_db_prod}"
TIMESTAMP=$(date +"%Y%m%d_%H%M%S")
BACKUP_FILE="$BACKUP_DIR/${PG_DB}_${TIMESTAMP}.sql.gz"

# ── التحقق من وجود مجلد الـ backups ──────────────────────────
mkdir -p "$BACKUP_DIR"

echo "▶ [$(date '+%Y-%m-%d %H:%M:%S')] starting backup → $BACKUP_FILE"

# ── تنفيذ pg_dump داخل الـ container ─────────────────────────
docker exec "$CONTAINER" \
    pg_dump -U "$PG_USER" -d "$PG_DB" \
    --format=plain \
    --no-owner \
    --no-acl \
| gzip -9 > "$BACKUP_FILE"

# ── التحقق من الحجم ───────────────────────────────────────────
BACKUP_SIZE=$(du -sh "$BACKUP_FILE" | cut -f1)
echo "✅ backup complete: $BACKUP_FILE ($BACKUP_SIZE)"

# ── حذف الـ backups القديمة (أكثر من RETENTION_DAYS يوم) ─────
DELETED=$(find "$BACKUP_DIR" \
    -name "${PG_DB}_*.sql.gz" \
    -mtime +"$RETENTION_DAYS" \
    -print -delete | wc -l)

if [[ "$DELETED" -gt 0 ]]; then
    echo "🗑  deleted $DELETED backup(s) older than $RETENTION_DAYS days"
fi

echo "✅ [$(date '+%Y-%m-%d %H:%M:%S')] backup job done"

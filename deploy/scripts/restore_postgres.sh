#!/usr/bin/env bash
# ============================================================
#  deploy/scripts/restore_postgres.sh
#  استرجاع قاعدة البيانات من ملف backup (.sql.gz)
#
#  ⚠️  DESTRUCTIVE — سيحذف البيانات الحالية ويستبدلها
#
#  الاستخدام:
#    ./deploy/scripts/restore_postgres.sh /opt/quran/backups/quran_school_20251001_030000.sql.gz
#
#  المتطلبات:
#    - الـ postgres container يجب أن يكون شغّالاً
#    - يجب إيقاف الـ api container أثناء الاسترجاع
# ============================================================
set -euo pipefail

# ── تحميل .env.prod إذا وُجد ─────────────────────────────────
SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
ENV_FILE="$SCRIPT_DIR/../.env.prod"
if [[ -f "$ENV_FILE" ]]; then
    set -a; source "$ENV_FILE"; set +a
fi

# ── التحقق من المدخلات ───────────────────────────────────────
if [[ $# -lt 1 ]]; then
    echo "Usage: $0 <backup_file.sql.gz>"
    echo ""
    echo "Available backups:"
    ls -lh "${BACKUP_DIR:-/opt/quran/backups}"/*.sql.gz 2>/dev/null || echo "  (none found)"
    exit 1
fi

BACKUP_FILE="$1"
PG_USER="${POSTGRES_USER:-quran}"
PG_DB="${POSTGRES_DB:-quran_school}"
CONTAINER="${CONTAINER_NAME:-quran_db_prod}"
API_CONTAINER="${API_CONTAINER_NAME:-quran_api_prod}"

# ── التحقق من وجود الملف ─────────────────────────────────────
if [[ ! -f "$BACKUP_FILE" ]]; then
    echo "❌ backup file not found: $BACKUP_FILE"
    exit 1
fi

echo "⚠️  WARNING: This will DESTROY and replace the current database."
echo "   DB:      $PG_DB"
echo "   From:    $BACKUP_FILE"
echo "   Size:    $(du -sh "$BACKUP_FILE" | cut -f1)"
echo ""
read -r -p "Type 'yes' to confirm: " CONFIRM
if [[ "$CONFIRM" != "yes" ]]; then
    echo "❌ aborted"
    exit 1
fi

echo "▶ [$(date '+%Y-%m-%d %H:%M:%S')] starting restore..."

# ── إيقاف الـ api لتجنب التعارض ──────────────────────────────
if docker ps --format '{{.Names}}' | grep -q "^${API_CONTAINER}$"; then
    echo "▶ stopping api container: $API_CONTAINER"
    docker stop "$API_CONTAINER"
    API_WAS_RUNNING=true
else
    API_WAS_RUNNING=false
fi

# ── إعادة إنشاء قاعدة البيانات ───────────────────────────────
echo "▶ dropping and recreating database $PG_DB ..."
docker exec "$CONTAINER" psql -U "$PG_USER" -d postgres \
    -c "SELECT pg_terminate_backend(pid) FROM pg_stat_activity WHERE datname='$PG_DB' AND pid <> pg_backend_pid();" \
    -c "DROP DATABASE IF EXISTS $PG_DB;" \
    -c "CREATE DATABASE $PG_DB OWNER $PG_USER;"

# ── استرجاع الـ dump ──────────────────────────────────────────
echo "▶ restoring from $BACKUP_FILE ..."
gunzip -c "$BACKUP_FILE" | docker exec -i "$CONTAINER" \
    psql -U "$PG_USER" -d "$PG_DB" --quiet

echo "✅ restore complete"

# ── إعادة تشغيل الـ api إذا كان شغّالاً ─────────────────────
if [[ "$API_WAS_RUNNING" == "true" ]]; then
    echo "▶ restarting api container: $API_CONTAINER"
    docker start "$API_CONTAINER"
fi

echo "✅ [$(date '+%Y-%m-%d %H:%M:%S')] restore done"

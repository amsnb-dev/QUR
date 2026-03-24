#!/usr/bin/env bash
# ============================================================
#  deploy/scripts/rotate_logs.sh
#  تنظيف / ضغط logs قديمة
#
#  ملاحظة: Docker يدير logs تلقائياً إذا حدّدت max-size في daemon.json
#  هذا السكريبت للحالات التي تحتاج تنظيفاً يدوياً إضافياً.
#
#  الاستخدام:
#    ./deploy/scripts/rotate_logs.sh
# ============================================================
set -euo pipefail

LOG_MAX_DAYS="${LOG_RETAIN_DAYS:-7}"

echo "▶ [$(date '+%Y-%m-%d %H:%M:%S')] rotating logs (retain $LOG_MAX_DAYS days)..."

# ── تنظيف nginx access logs القديمة (إذا كانت مُوجَّهة لملفات) ──
NGINX_LOG_DIR="/var/log/nginx"
if [[ -d "$NGINX_LOG_DIR" ]]; then
    find "$NGINX_LOG_DIR" -name "*.log.*" -mtime +"$LOG_MAX_DAYS" -delete 2>/dev/null || true
    echo "✅ nginx old logs cleaned (older than $LOG_MAX_DAYS days)"
fi

# ── إرسال SIGUSR1 لـ nginx ليعيد فتح ملفات الـ logs ─────────
if docker ps --format '{{.Names}}' | grep -q "quran_nginx_prod"; then
    docker kill --signal=USR1 quran_nginx_prod 2>/dev/null || true
    echo "✅ nginx log files reopened"
fi

# ── حجم logs Docker الحالية ──────────────────────────────────
echo ""
echo "Docker container log sizes:"
for container in quran_db_prod quran_api_prod quran_nginx_prod; do
    LOG_PATH=$(docker inspect --format='{{.LogPath}}' "$container" 2>/dev/null || echo "")
    if [[ -n "$LOG_PATH" && -f "$LOG_PATH" ]]; then
        SIZE=$(du -sh "$LOG_PATH" | cut -f1)
        echo "  $container: $SIZE"
    fi
done

echo ""
echo "✅ [$(date '+%Y-%m-%d %H:%M:%S')] log rotation done"

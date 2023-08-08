#!/bin/sh

# 啟用 "遇到錯誤即停止" 的模式
set -e

echo "run db migration"
/app/migrate -path /app/migration -database ${DB_SOURCE} --verbose up

echo "start the app"
# takes all parameters passed to the script and run it
exec "$@"
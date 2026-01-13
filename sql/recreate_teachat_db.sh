#!/bin/bash
set -e

# 配置
PG_BIN="/Library/PostgreSQL/13/bin"
DB_USER="robin"
DB_NAME="teachat"
PROJECT_ROOT="/Users/robin/Desktop/teachat"
BACKUP_DIR="${PROJECT_ROOT}/backups"
TIMESTAMP=$(date +%Y%m%d_%H%M%S)

echo "=== Teachat数据库重建流程 ==="

# 1. 备份当前users表
echo "1. 备份users表..."
mkdir -p "${BACKUP_DIR}"
"${PG_BIN}/pg_dump" -U "${DB_USER}" -t users --data-only --inserts "${DB_NAME}" > "${BACKUP_DIR}/users_backup_${TIMESTAMP}.sql"

echo "  备份保存至: ${BACKUP_DIR}/users_backup_${TIMESTAMP}.sql"

# 2. 重建数据库
echo "2. 重建数据库..."
"${PG_BIN}/dropdb" -U "${DB_USER}" "${DB_NAME}" || true
"${PG_BIN}/createdb" -U "${DB_USER}" "${DB_NAME}"

echo "3. 导入架构..."
"${PG_BIN}/psql" -U "${DB_USER}" -d "${DB_NAME}" -f "${PROJECT_ROOT}/sql/schema.sql"

echo "4. 导入预设数据..."
"${PG_BIN}/psql" -U "${DB_USER}" -d "${DB_NAME}" -f "${PROJECT_ROOT}/sql/seed_data.sql"

# 3. 处理备份文件并恢复
echo "5. 处理备份文件（添加ON CONFLICT子句）..."
sed -i '' 's/^INSERT INTO users/INSERT INTO users ON CONFLICT (id) DO NOTHING;/' "${BACKUP_DIR}/users_backup_${TIMESTAMP}.sql"

echo "6. 恢复用户数据..."
"${PG_BIN}/psql" -U "${DB_USER}" -d "${DB_NAME}" -f "${BACKUP_DIR}/users_backup_${TIMESTAMP}.sql"

echo "✅ 数据库重建完成！"
echo "   预设用户 (id=1,2) 已保留"
echo "   其他用户已从备份恢复"

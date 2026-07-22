#!/bin/bash
# 清空预算管理模块数据（保留总公司和 admin 账号）
# 用法: ./scripts/reset-budget-data.sh
# 前提: Docker 容器 newapi-postgres-1 正在运行

set -e

CONTAINER="newapi-postgres-1"
DB_USER="tokenjoy"
DB_NAME="tokenjoy"
COMPANY_ID=2
ADMIN_ID="m-admin"
ROOT_DEPT="dept-1"

echo "🧹 清空预算管理数据（保留总公司 + admin）..."

docker exec "$CONTAINER" psql -U "$DB_USER" -d "$DB_NAME" -c "
BEGIN;

-- 解除 platform_keys 对 project 和 member 的外键引用
UPDATE platform_keys SET project_id = NULL WHERE company_id = $COMPANY_ID;
UPDATE platform_keys SET member_id = '$ADMIN_ID' WHERE company_id = $COMPANY_ID AND member_id != '$ADMIN_ID';

-- 删除项目关联
DELETE FROM project_members WHERE company_id = $COMPANY_ID;
DELETE FROM projects WHERE company_id = $COMPANY_ID;

-- 删除告警规则
DELETE FROM alert_rule_notify_roles WHERE company_id = $COMPANY_ID;
DELETE FROM alert_rules WHERE company_id = $COMPANY_ID;

-- 删除审批记录
DELETE FROM approval_requests WHERE company_id = $COMPANY_ID;

-- 删除成员角色和成员（保留 admin）
DELETE FROM member_roles WHERE company_id = $COMPANY_ID AND member_id != '$ADMIN_ID';
DELETE FROM members WHERE company_id = $COMPANY_ID AND id != '$ADMIN_ID';

-- 删除子部门的模型白名单
DELETE FROM model_allowlist WHERE owner_type = 'org_node' AND owner_id != '$ROOT_DEPT';

-- 删除子部门节点（先叶子后父级）
DELETE FROM org_nodes WHERE company_id = $COMPANY_ID AND id != '$ROOT_DEPT'
  AND id NOT IN (SELECT parent_id FROM org_nodes WHERE company_id = $COMPANY_ID AND parent_id IS NOT NULL AND id != '$ROOT_DEPT');
DELETE FROM org_nodes WHERE company_id = $COMPANY_ID AND id != '$ROOT_DEPT';

-- 确保 admin 归属总公司
UPDATE members SET department_id = '$ROOT_DEPT' WHERE id = '$ADMIN_ID' AND company_id = $COMPANY_ID;

-- 重置所有成员的个人额度为 0
UPDATE members SET personal_budget = 0 WHERE company_id = $COMPANY_ID;

COMMIT;
"

echo "✅ 清空完成。当前仅保留："
docker exec "$CONTAINER" psql -U "$DB_USER" -d "$DB_NAME" -c "
  SELECT '部门' as type, count(*) as count FROM org_nodes WHERE company_id = $COMPANY_ID
  UNION ALL
  SELECT '成员', count(*) FROM members WHERE company_id = $COMPANY_ID
  UNION ALL
  SELECT '项目', count(*) FROM projects WHERE company_id = $COMPANY_ID;
"

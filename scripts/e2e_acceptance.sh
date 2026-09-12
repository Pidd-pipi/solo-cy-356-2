#!/usr/bin/env bash
# 端到端验收脚本：菜园地块认养申请模块（含流转规则修复回归）
# 规则：同一居民在同一地块仅允许一份进行中申请；同一居民同一时间仅允许一份待审核申请；
#       候补中的居民可申请其他空闲地块；直接认养入口已移除，地块被认养只能经由申请审核通过。
# 前置：后端已启动（默认 http://localhost:29516），数据库已迁移并含种子数据（admin/admin123）。
# 用法：BASE=http://localhost:29516/api/v1 bash scripts/e2e_acceptance.sh
# 每次运行生成唯一用户与地块（RUN 后缀），可对同一环境重复执行。
set -euo pipefail

BASE="${BASE:-http://localhost:29516/api/v1}"
RUN="${RUN:-$(date +%s)}"
PASS=0
FAIL=0

jqr() { jq -r "$1"; }

check() { # $1=描述 $2=实际 $3=期望
  if [ "$2" = "$3" ]; then
    PASS=$((PASS + 1)); echo "  ✔ $1"
  else
    FAIL=$((FAIL + 1)); echo "  ✘ $1  [got=$2 want=$3]"
  fi
}

api() { # $1=method $2=path $3=token(可空) $4=body(可空)
  local method="$1" path="$2" token="${3:-}" body="${4:-}"
  if [ -n "$body" ]; then
    curl -s -X "$method" "$BASE$path" -H 'Content-Type: application/json' ${token:+-H "Authorization: Bearer $token"} -d "$body"
  else
    curl -s -X "$method" "$BASE$path" ${token:+-H "Authorization: Bearer $token"}
  fi
}

http_code() { # 同 api，仅返回 HTTP 状态码
  local method="$1" path="$2" token="${3:-}" body="${4:-}"
  if [ -n "$body" ]; then
    curl -s -o /dev/null -w '%{http_code}' -X "$method" "$BASE$path" -H 'Content-Type: application/json' ${token:+-H "Authorization: Bearer $token"} -d "$body"
  else
    curl -s -o /dev/null -w '%{http_code}' -X "$method" "$BASE$path" ${token:+-H "Authorization: Bearer $token"}
  fi
}

echo "== 0. 准备：管理员登录 + 注册 3 位居民 + 创建 2 块空闲地块 (RUN=$RUN) =="
ADMIN=$(api POST /auth/login '' '{"username":"admin","password":"admin123"}' | jqr .data.token)
[ "$ADMIN" != "null" ] && [ -n "$ADMIN" ] || { echo "admin 登录失败"; exit 1; }
for u in u1 u2 u3; do
  api POST /auth/register '' "{\"username\":\"${u}_$RUN\",\"password\":\"pass123\",\"nickname\":\"居民$u\"}" > /dev/null
done
T1=$(api POST /auth/login '' "{\"username\":\"u1_$RUN\",\"password\":\"pass123\"}" | jqr .data.token)
T2=$(api POST /auth/login '' "{\"username\":\"u2_$RUN\",\"password\":\"pass123\"}" | jqr .data.token)
T3=$(api POST /auth/login '' "{\"username\":\"u3_$RUN\",\"password\":\"pass123\"}" | jqr .data.token)
U3ID=$(api GET /me "$T3" | jqr .data.id)
P1=$(api POST /plots "$ADMIN" "{\"name\":\"验收地块一\",\"code\":\"E2E1-$RUN\",\"area\":10,\"soil_type\":\"loam\",\"sunlight\":\"full\",\"latitude\":31.23,\"longitude\":121.47}" | jqr .data.id)
P2=$(api POST /plots "$ADMIN" "{\"name\":\"验收地块二\",\"code\":\"E2E2-$RUN\",\"area\":12,\"soil_type\":\"black\",\"sunlight\":\"partial\",\"latitude\":31.231,\"longitude\":121.472}" | jqr .data.id)
echo "  地块 P1=$P1 P2=$P2"

echo "== 1. 申请规则：同一地块/同一居民仅一份待审核 =="
R=$(api POST /applications "$T1" "{\"plot_id\":$P1,\"reason\":\"想种番茄\"}")
A1=$(echo "$R" | jqr .data.id)
check "u1 申请空闲地块 -> 待审核" "$(echo "$R" | jqr .data.status)" "pending"
check "同一地块重复提交被拦截(409)" "$(http_code POST /applications "$T1" "{\"plot_id\":$P1}")" "409"
check "已有待审核申请再申请无待审地块被拦截(409)" "$(http_code POST /applications "$T1" "{\"plot_id\":$P2}")" "409"
check "缺少 plot_id 参数校验失败(400)" "$(http_code POST /applications "$T3" '{}')" "400"

echo "== 2. 候补中的居民可申请其他空闲地块 =="
R=$(api POST /applications "$T2" "{\"plot_id\":$P1,\"reason\":\"也想种\"}")
A2=$(echo "$R" | jqr .data.id)
check "u2 申请 P1（已有待审）-> 候补" "$(echo "$R" | jqr .data.status)" "waitlisted"
R=$(api POST /applications "$T2" "{\"plot_id\":$P2}")
A3=$(echo "$R" | jqr .data.id)
check "u2 候补中再申请 P2（无待审）-> 待审核" "$(echo "$R" | jqr .data.status)" "pending"
R=$(api POST /applications "$T3" "{\"plot_id\":$P1}")
A4=$(echo "$R" | jqr .data.id)
check "u3 申请 P1 -> 候补" "$(echo "$R" | jqr .data.status)" "waitlisted"

echo "== 3. 进度查询 =="
check "u2 的申请列表同时包含候补与待审核" "$(api GET /applications/mine "$T2" | jq -r '[.data.list[].status] | sort | join(",")')" "pending,waitlisted"

echo "== 4. 撤回与候补晋升（跳过已持有待审核申请的人） =="
check "他人撤回我的申请被拦截(403)" "$(http_code POST "/applications/$A1/withdraw" "$T2")" "403"
check "u1 撤回自己的申请" "$(api POST "/applications/$A1/withdraw" "$T1" | jqr .data.status)" "withdrawn"
check "最早候补 u2 已有待审核申请被跳过（仍为候补）" "$(api GET /applications/mine "$T2" | jq -r "[.data.list[] | select(.plot_id==$P1)][0].status")" "waitlisted"
check "下一位合格候补 u3 转为待审核" "$(api GET /applications/mine "$T3" | jqr '.data.list[0].status')" "pending"

echo "== 5. 权限控制 =="
check "非管理员访问全部申请列表被拦截(403)" "$(http_code GET /applications "$T1")" "403"
check "非管理员审核被拦截(403)" "$(http_code POST "/applications/$A4/review" "$T1" '{"action":"approve"}')" "403"
check "未登录访问进度查询被拦截(401)" "$(http_code GET /applications/mine '')" "401"

echo "== 6. 审核通过同步 + 直接认养入口已移除 =="
check "管理员审核通过 u3 的 P1 申请" "$(api POST "/applications/$A4/review" "$ADMIN" '{"action":"approve","note":"欢迎认养"}' | jqr .data.status)" "approved"
check "地块状态同步为已认养" "$(api GET "/plots/$P1" '' | jqr .data.status)" "adopted"
check "地块认养人同步为申请人 u3" "$(api GET "/plots/$P1" '' | jqr .data.adopter_id)" "$U3ID"
check "重复审核同一申请被拦截(409)" "$(http_code POST "/applications/$A4/review" "$ADMIN" '{"action":"reject"}')" "409"
check "已认养地块不可再申请(409)" "$(http_code POST /applications "$T1" "{\"plot_id\":$P1}")" "409"
check "旧的直接认养入口已移除(404)" "$(http_code POST "/plots/$P2/adopt" "$T1")" "404"

echo "== 7. 驳回与候补晋升 =="
R=$(api POST /applications "$T1" "{\"plot_id\":$P2}")
A5=$(echo "$R" | jqr .data.id)
check "u1 申请 P2（u2 待审中）-> 候补" "$(echo "$R" | jqr .data.status)" "waitlisted"
check "管理员驳回 u2 的 P2 申请" "$(api POST "/applications/$A3/review" "$ADMIN" '{"action":"reject","note":"信息不完整"}' | jqr .data.status)" "rejected"
check "驳回后最早合格候补 u1 转为待审核" "$(api GET /applications/mine "$T1" | jqr '.data.list[0].status')" "pending"

echo "== 8. 地块释放后候补晋升（计划完成 -> 待释放 -> 释放） =="
R=$(api POST /applications "$T2" "{\"plot_id\":$P2}")
A6=$(echo "$R" | jqr .data.id)
check "u2 对 P2 的申请进入候补" "$(echo "$R" | jqr .data.status)" "waitlisted"
check "管理员通过 u1 的 P2 申请" "$(api POST "/applications/$A5/review" "$ADMIN" '{"action":"approve"}' | jqr .data.status)" "approved"
check "P2 同步为已认养" "$(api GET "/plots/$P2" '' | jqr .data.status)" "adopted"
PLAN=$(api POST /planting-plans "$T1" "{\"plot_id\":$P2,\"crop_name\":\"菠菜\",\"crop_type\":\"vegetable\",\"season\":\"spring\"}" | jqr .data.id)
for s in planting growing harvesting completed; do
  api POST "/planting-plans/$PLAN/status" "$T1" "{\"status\":\"$s\"}" > /dev/null
done
check "计划完成后地块进入待释放" "$(api GET "/plots/$P2" '' | jqr .data.status)" "harvested"
check "认养人释放地块" "$(api POST "/plots/$P2/release" "$T1" | jqr .data.status)" "available"
check "释放后最早合格候补 u2 转为待审核" "$(api GET /applications/mine "$T2" | jq -r "[.data.list[] | select(.plot_id==$P2)][0].status")" "pending"

echo ""
echo "=========================================="
echo "验收结果：PASS=$PASS FAIL=$FAIL"
echo "=========================================="
[ "$FAIL" -eq 0 ]

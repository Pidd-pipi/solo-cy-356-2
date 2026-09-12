#!/usr/bin/env bash
# 并发验收脚本：认养申请"同一居民仅一份待审核"不变式在并发下的回归验证
# 场景1：居民申请新地块 与 其候补被驳回晋升 并发 -> 至多一份待审核
# 场景2：居民申请新地块 与 其候补被撤回晋升 并发 -> 至多一份待审核
# 场景3：同一居民并发申请两块空闲地块 -> 恰好一份待审核
# 前置：后端已启动（默认 http://localhost:29516），种子账号 admin/admin123 可用。
# 用法：BASE=http://localhost:29516/api/v1 ROUNDS=30 bash scripts/e2e_concurrency.sh
set -euo pipefail

BASE="${BASE:-http://localhost:29516/api/v1}"
ROUNDS="${ROUNDS:-30}"
RUN="${RUN:-$(date +%s)}"
PASS=0
FAIL=0

jqr() { jq -r "$1"; }

api() { # $1=method $2=path $3=token(可空) $4=body(可空)
  local method="$1" path="$2" token="${3:-}" body="${4:-}"
  if [ -n "$body" ]; then
    curl -s -X "$method" "$BASE$path" -H 'Content-Type: application/json' ${token:+-H "Authorization: Bearer $token"} -d "$body"
  else
    curl -s -X "$method" "$BASE$path" ${token:+-H "Authorization: Bearer $token"}
  fi
}

login() { api POST /auth/login '' "{\"username\":\"$1\",\"password\":\"$2\"}" | jqr .data.token; }
register() { api POST /auth/register '' "{\"username\":\"$1\",\"password\":\"pass123\",\"nickname\":\"$1\"}" > /dev/null; }
mkplot() { api POST /plots "$ADMIN" "{\"name\":\"并发验收地块\",\"code\":\"$1\",\"area\":10,\"soil_type\":\"loam\",\"sunlight\":\"full\",\"latitude\":31.23,\"longitude\":121.47}" | jqr .data.id; }
pending_count() { api GET "/applications/mine?page_size=50" "$1" | jq '[.data.list[] | select(.status=="pending")] | length'; }

ADMIN=$(login admin admin123)
[ "$ADMIN" != "null" ] && [ -n "$ADMIN" ] || { echo "admin 登录失败"; exit 1; }

echo "== 并发场景 1/2：申请新地块 与 驳回/撤回触发晋升 并发（$ROUNDS 轮） =="
S1_FAIL=0
S2_FAIL=0
for i in $(seq 1 "$ROUNDS"); do
  register "cc1-x-$RUN-$i"; register "cc1-u-$RUN-$i"
  TX=$(login "cc1-x-$RUN-$i" pass123); TU=$(login "cc1-u-$RUN-$i" pass123)
  PA=$(mkplot "CC1A-$RUN-$i"); PB=$(mkplot "CC1B-$RUN-$i")
  # x 在 A 地块待审核，u 在 A 地块候补
  APP_X=$(api POST /applications "$TX" "{\"plot_id\":$PA}" | jqr .data.id)
  api POST /applications "$TU" "{\"plot_id\":$PA}" > /dev/null
  # 并发：u 申请 B 地块 + admin 驳回 x 的申请（晋升 u 的候补）
  api POST /applications "$TU" "{\"plot_id\":$PB}" > /tmp/cc1_apply_$i.json &
  api POST "/applications/$APP_X/review" "$ADMIN" '{"action":"reject"}' > /tmp/cc1_review_$i.json &
  wait
  CNT=$(pending_count "$TU")
  if [ "$CNT" -gt 1 ]; then S1_FAIL=$((S1_FAIL+1)); echo "  ✘ 场景1 第 $i 轮：用户出现 $CNT 份待审核"; fi

  # 场景2：换撤回触发晋升（重新搭一组）
  register "cc2-x-$RUN-$i"; register "cc2-u-$RUN-$i"
  TX2=$(login "cc2-x-$RUN-$i" pass123); TU2=$(login "cc2-u-$RUN-$i" pass123)
  PA2=$(mkplot "CC2A-$RUN-$i"); PB2=$(mkplot "CC2B-$RUN-$i")
  APP_X2=$(api POST /applications "$TX2" "{\"plot_id\":$PA2}" | jqr .data.id)
  api POST /applications "$TU2" "{\"plot_id\":$PA2}" > /dev/null
  # 并发：u 申请 B 地块 + x 撤回自己的申请（晋升 u 的候补）
  api POST /applications "$TU2" "{\"plot_id\":$PB2}" > /tmp/cc2_apply_$i.json &
  api POST "/applications/$APP_X2/withdraw" "$TX2" > /tmp/cc2_withdraw_$i.json &
  wait
  CNT2=$(pending_count "$TU2")
  if [ "$CNT2" -gt 1 ]; then S2_FAIL=$((S2_FAIL+1)); echo "  ✘ 场景2 第 $i 轮：用户出现 $CNT2 份待审核"; fi
  rm -f /tmp/cc1_apply_$i.json /tmp/cc1_review_$i.json /tmp/cc2_apply_$i.json /tmp/cc2_withdraw_$i.json
done
if [ "$S1_FAIL" -eq 0 ]; then PASS=$((PASS+1)); echo "  ✔ 场景1（申请 vs 驳回晋升）$ROUNDS 轮全部保持唯一待审核"; else FAIL=$((FAIL+1)); echo "  ✘ 场景1 失败 $S1_FAIL 轮"; fi
if [ "$S2_FAIL" -eq 0 ]; then PASS=$((PASS+1)); echo "  ✔ 场景2（申请 vs 撤回晋升）$ROUNDS 轮全部保持唯一待审核"; else FAIL=$((FAIL+1)); echo "  ✘ 场景2 失败 $S2_FAIL 轮"; fi

echo "== 并发场景 3：同一居民并发申请两块空闲地块（$ROUNDS 轮） =="
S3_FAIL=0
for i in $(seq 1 "$ROUNDS"); do
  register "cc3-u-$RUN-$i"
  TU3=$(login "cc3-u-$RUN-$i" pass123)
  PB3=$(mkplot "CC3B-$RUN-$i"); PC3=$(mkplot "CC3C-$RUN-$i")
  api POST /applications "$TU3" "{\"plot_id\":$PB3}" > /tmp/cc3_a_$i.json &
  api POST /applications "$TU3" "{\"plot_id\":$PC3}" > /tmp/cc3_b_$i.json &
  wait
  CNT3=$(pending_count "$TU3")
  if [ "$CNT3" -ne 1 ]; then S3_FAIL=$((S3_FAIL+1)); echo "  ✘ 场景3 第 $i 轮：待审核数=$CNT3（期望恰好 1）"; fi
  rm -f /tmp/cc3_a_$i.json /tmp/cc3_b_$i.json
done
if [ "$S3_FAIL" -eq 0 ]; then PASS=$((PASS+1)); echo "  ✔ 场景3（并发双申请）$ROUNDS 轮全部恰好一份待审核"; else FAIL=$((FAIL+1)); echo "  ✘ 场景3 失败 $S3_FAIL 轮"; fi

echo ""
echo "=========================================="
echo "并发验收结果：PASS=$PASS FAIL=$FAIL（每场景 $ROUNDS 轮）"
echo "=========================================="
[ "$FAIL" -eq 0 ]

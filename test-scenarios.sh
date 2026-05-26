#!/usr/bin/env bash
set -uo pipefail

BASE="http://localhost:8080"
DB="./mundoinvest.db"

check_dep() { command -v "$1" >/dev/null 2>&1 || { echo "erro: '$1' nao encontrado. instale-o primeiro."; exit 1; }; }
check_dep jq
check_dep sqlite3
PASS=0
FAIL=0

RED='\033[0;31m'
GREEN='\033[0;32m'
NC='\033[0m'

pass() {
    ((PASS++)) || true
    echo -e "  ${GREEN}✓${NC}  ${1}"
}
fail() {
    ((FAIL++)) || true
    echo -e "  ${RED}✗${NC}  ${1}"
}

assert_http() { [ "$1" -eq "$2" ] && return 0 || return 1; }
assert_json()  { echo "$1" | jq -e "$2" > /dev/null 2>&1 && return 0 || return 1; }

db_query()    { sqlite3 "$DB" "$1" 2>/dev/null; }

echo ""
echo "=== Mundo Invest — Test Scenarios ==="
echo ""

if ! curl -s -o /dev/null -w "%{http_code}" "$BASE/clientes" >/dev/null 2>&1; then
    echo "erro: servidor nao esta respondendo em $BASE"
    echo "inicie com: make clean && go run ./cmd/server"
    exit 1
fi

# ─── 1.1 Criar cliente payload válido ───
RESP=$(curl -s -o /dev/null -w "%{http_code}" -X POST "$BASE/clientes" \
    -H "Content-Type: application/json" \
    -d '{"cliente_nome":"João Silva","cliente_email":"alta@test.com","tipo_solicitacao":"Atualização cadastral","valor_patrimonio":250000}')
if assert_http "$RESP" 201; then
    pass "1.1  Criar cliente válido (João Silva, 250k)                    [201]"
else
    fail "1.1  Criar cliente válido (João Silva, 250k) — esperado 201, veio $RESP"
fi

# ─── 1.2 Email inválido ───
RESP=$(curl -s -o /dev/null -w "%{http_code}" -X POST "$BASE/clientes" \
    -H "Content-Type: application/json" \
    -d '{"cliente_nome":"X","cliente_email":"invalido","tipo_solicitacao":"X","valor_patrimonio":1000}')
BODY=$(curl -s -X POST "$BASE/clientes" \
    -H "Content-Type: application/json" \
    -d '{"cliente_nome":"X","cliente_email":"invalido","tipo_solicitacao":"X","valor_patrimonio":1000}')
if assert_http "$RESP" 400 && assert_json "$BODY" '.error.code == "VALIDATION_ERROR"'; then
    pass "1.2  Email inválido                                                    [400]"
else
    fail "1.2  Email inválido — esperado 400 VALIDATION_ERROR, veio $RESP"
fi

# ─── 1.3 Campos obrigatórios ausentes ───
RESP=$(curl -s -o /dev/null -w "%{http_code}" -X POST "$BASE/clientes" \
    -H "Content-Type: application/json" \
    -d '{"cliente_nome":"João Silva"}')
BODY=$(curl -s -X POST "$BASE/clientes" \
    -H "Content-Type: application/json" \
    -d '{"cliente_nome":"João Silva"}')
if assert_http "$RESP" 400 && assert_json "$BODY" '.error.code == "VALIDATION_ERROR"'; then
    pass "1.3  Campos obrigatórios ausentes                                      [400]"
else
    fail "1.3  Campos obrigatórios ausentes — esperado 400, veio $RESP"
fi

# ─── 1.4 valor_patrimonio = 0 ───
RESP=$(curl -s -o /dev/null -w "%{http_code}" -X POST "$BASE/clientes" \
    -H "Content-Type: application/json" \
    -d '{"cliente_nome":"X","cliente_email":"x@x.com","tipo_solicitacao":"X","valor_patrimonio":0}')
BODY=$(curl -s -X POST "$BASE/clientes" \
    -H "Content-Type: application/json" \
    -d '{"cliente_nome":"X","cliente_email":"x@x.com","tipo_solicitacao":"X","valor_patrimonio":0}')
if assert_http "$RESP" 400 && assert_json "$BODY" '.error.code == "VALIDATION_ERROR"'; then
    pass "1.4  valor_patrimonio = 0                                              [400]"
else
    fail "1.4  valor_patrimonio = 0 — esperado 400, veio $RESP"
fi

# ─── 1.5 Body vazio ───
RESP=$(curl -s -o /dev/null -w "%{http_code}" -X POST "$BASE/clientes" \
    -H "Content-Type: application/json" \
    -d '{}')
BODY=$(curl -s -X POST "$BASE/clientes" \
    -H "Content-Type: application/json" \
    -d '{}')
if assert_http "$RESP" 400 && assert_json "$BODY" '.error.code == "VALIDATION_ERROR"'; then
    pass "1.5  Body vazio                                                        [400]"
else
    fail "1.5  Body vazio — esperado 400, veio $RESP"
fi

# ─── 1.6 Email duplicado (reusa 1.1) ───
RESP=$(curl -s -o /dev/null -w "%{http_code}" -X POST "$BASE/clientes" \
    -H "Content-Type: application/json" \
    -d '{"cliente_nome":"João Silva","cliente_email":"alta@test.com","tipo_solicitacao":"Atualização cadastral","valor_patrimonio":250000}')
BODY=$(curl -s -X POST "$BASE/clientes" \
    -H "Content-Type: application/json" \
    -d '{"cliente_nome":"João Silva","cliente_email":"alta@test.com","tipo_solicitacao":"Atualização cadastral","valor_patrimonio":250000}')
if assert_http "$RESP" 409 && assert_json "$BODY" '.error.code == "EMAIL_ALREADY_EXISTS"'; then
    pass "1.6  Email duplicado (alta@test.com)                                    [409]"
else
    fail "1.6  Email duplicado — esperado 409 EMAIL_ALREADY_EXISTS, veio $RESP"
fi

# ─── 2.1 Webhook 250k → prioridade_alta (cliente do 1.1) ───
RESP=$(curl -s -o /dev/null -w "%{http_code}" -X POST "$BASE/webhooks/pipefy/card-updated" \
    -H "Content-Type: application/json" \
    -d '{"event_id":"evt_001","card_id":"card_001","cliente_email":"alta@test.com","timestamp":"2026-05-18T12:00:00Z"}')
STATUS=$(db_query "SELECT status FROM clients WHERE email='alta@test.com';")
PRIO=$(db_query   "SELECT priority FROM clients WHERE email='alta@test.com';")
if assert_http "$RESP" 200 && [ "$STATUS" = "Processado" ] && [ "$PRIO" = "prioridade_alta" ]; then
    pass "2.1  Webhook 250k → prioridade_alta                                   [200] DB: $STATUS / $PRIO"
else
    fail "2.1  Webhook 250k — esperado 200 + Processado/prioridade_alta, veio $RESP ($STATUS / $PRIO)"
fi

# ─── 2.2 Criar cliente 50k + webhook → prioridade_normal ───
curl -s -X POST "$BASE/clientes" \
    -H "Content-Type: application/json" \
    -d '{"cliente_nome":"Maria Souza","cliente_email":"normal@test.com","tipo_solicitacao":"Atualização cadastral","valor_patrimonio":50000}' \
    > /dev/null
RESP=$(curl -s -o /dev/null -w "%{http_code}" -X POST "$BASE/webhooks/pipefy/card-updated" \
    -H "Content-Type: application/json" \
    -d '{"event_id":"evt_002","card_id":"card_002","cliente_email":"normal@test.com","timestamp":"2026-05-18T12:00:00Z"}')
PRIO=$(db_query "SELECT priority FROM clients WHERE email='normal@test.com';")
if assert_http "$RESP" 200 && [ "$PRIO" = "prioridade_normal" ]; then
    pass "2.2  Webhook 50k → prioridade_normal                                  [200] DB: $PRIO"
else
    fail "2.2  Webhook 50k — esperado 200 + prioridade_normal, veio $RESP ($PRIO)"
fi

# ─── 2.3 Criar cliente 200k boundary + webhook → prioridade_alta ───
curl -s -X POST "$BASE/clientes" \
    -H "Content-Type: application/json" \
    -d '{"cliente_nome":"Carlos Silva","cliente_email":"boundary@test.com","tipo_solicitacao":"Atualização cadastral","valor_patrimonio":200000}' \
    > /dev/null
RESP=$(curl -s -o /dev/null -w "%{http_code}" -X POST "$BASE/webhooks/pipefy/card-updated" \
    -H "Content-Type: application/json" \
    -d '{"event_id":"evt_003","card_id":"card_003","cliente_email":"boundary@test.com","timestamp":"2026-05-18T12:00:00Z"}')
PRIO=$(db_query "SELECT priority FROM clients WHERE email='boundary@test.com';")
if assert_http "$RESP" 200 && [ "$PRIO" = "prioridade_alta" ]; then
    pass "2.3  Webhook 200k → prioridade_alta (boundary)                        [200] DB: $PRIO"
else
    fail "2.3  Webhook 200k boundary — esperado 200 + prioridade_alta, veio $RESP ($PRIO)"
fi

# ─── 2.4 Evento duplicado (reusa evt_001 do 2.1) ───
RESP=$(curl -s -o /dev/null -w "%{http_code}" -X POST "$BASE/webhooks/pipefy/card-updated" \
    -H "Content-Type: application/json" \
    -d '{"event_id":"evt_001","card_id":"card_001","cliente_email":"alta@test.com","timestamp":"2026-05-18T12:00:00Z"}')
BODY=$(curl -s -X POST "$BASE/webhooks/pipefy/card-updated" \
    -H "Content-Type: application/json" \
    -d '{"event_id":"evt_001","card_id":"card_001","cliente_email":"alta@test.com","timestamp":"2026-05-18T12:00:00Z"}')
if assert_http "$RESP" 409 && assert_json "$BODY" '.error.code == "EVENT_ALREADY_PROCESSED"'; then
    pass "2.4  Evento duplicado (evt_001) → 409                                  [409]"
else
    fail "2.4  Evento duplicado — esperado 409 EVENT_ALREADY_PROCESSED, veio $RESP"
fi

# ─── 2.5 Cliente não encontrado ───
RESP=$(curl -s -o /dev/null -w "%{http_code}" -X POST "$BASE/webhooks/pipefy/card-updated" \
    -H "Content-Type: application/json" \
    -d '{"event_id":"evt_004","card_id":"card_999","cliente_email":"ghost@test.com","timestamp":"2026-05-18T12:00:00Z"}')
BODY=$(curl -s -X POST "$BASE/webhooks/pipefy/card-updated" \
    -H "Content-Type: application/json" \
    -d '{"event_id":"evt_004","card_id":"card_999","cliente_email":"ghost@test.com","timestamp":"2026-05-18T12:00:00Z"}')
if assert_http "$RESP" 404 && assert_json "$BODY" '.error.code == "CLIENT_NOT_FOUND"'; then
    pass "2.5  Cliente não encontrado                                             [404]"
else
    fail "2.5  Cliente não encontrado — esperado 404, veio $RESP"
fi

# ─── 2.6 Webhook payload inválido (sem event_id) ───
RESP=$(curl -s -o /dev/null -w "%{http_code}" -X POST "$BASE/webhooks/pipefy/card-updated" \
    -H "Content-Type: application/json" \
    -d '{"card_id":"x","cliente_email":"a@a.com","timestamp":"2026-01-01T00:00:00Z"}')
BODY=$(curl -s -X POST "$BASE/webhooks/pipefy/card-updated" \
    -H "Content-Type: application/json" \
    -d '{"card_id":"x","cliente_email":"a@a.com","timestamp":"2026-01-01T00:00:00Z"}')
if assert_http "$RESP" 400 && assert_json "$BODY" '.error.code == "VALIDATION_ERROR"'; then
    pass "2.6  Webhook sem event_id → 400                                        [400]"
else
    fail "2.6  Webhook sem event_id — esperado 400, veio $RESP"
fi

# ─── Resumo ───
TOTAL=$((PASS + FAIL))
echo ""
echo "─────────────────────────────────────────"
if [ "$FAIL" -eq 0 ]; then
    echo -e "  Resultado: ${GREEN}$PASS/$TOTAL passaram${NC}"
else
    echo -e "  Resultado: $PASS/$TOTAL passaram ${RED}($FAIL falharam)${NC}"
fi
echo ""

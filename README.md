# Mundo Invest — Client Management & Pipefy Integration

API de gerenciamento de clientes com mapeamento de cards para o Pipefy, desenvolvida como teste técnico para backend.

## Estrutura do Projeto (DDD Estratégico)

```
├── cmd/server/main.go          # entrypoint: Gin router, DI, migrations
├── internal/
│   ├── cliente/                # Contexto: Gestão de Clientes
│   │   ├── handler.go          # POST /clientes
│   │   ├── service.go          # validação, persistência, mapeamento Pipefy
│   │   ├── repository.go       # SQLite CRUD (clientes)
│   │   ├── model.go            # Cliente, CriarClienteInput
│   │   └── handler_test.go     # teste integração: criação + persistência
│   ├── webhook/                # Contexto: Processamento de Webhooks
│   │   ├── handler.go          # POST /webhooks/pipefy/card-updated
│   │   ├── service.go          # idempotência, regra de prioridade, Pipefy
│   │   ├── repository.go       # SQLite (eventos_processados)
│   │   ├── model.go            # CardUpdatedInput, ProcessedEvent
│   │   └── handler_test.go     # testes: prioridade, evento duplicado
│   └── pipefy/                 # Contexto: Integração Pipefy
│       ├── client.go           # Cliente Pipefy (simulado)
│       ├── mutations.go        # GraphQL: createCard, updateCardField
│       └── models.go           # DTOs Pipefy (CreateCardInput, etc)
├── go.mod
└── README.md
```

## Stack

| Camada | Tecnologia |
|--------|-----------|
| Linguagem | Go 1.24 |
| HTTP | [Gin](https://github.com/gin-gonic/gin) |
| Banco | SQLite via [`modernc.org/sqlite`](https://pkg.go.dev/modernc.org/sqlite) (Go puro, zero CGO) |
| Testes | `testing` + `httptest` + `gin.TestMode` |

## Execução Local

```bash
# build
go build -buildvcs=false -o server ./cmd/server

# rodar servidor
./server
# Servidor iniciado em :8080

# rodar testes
go test -buildvcs=false -v ./...
```

## Exemplos de Requisição

### 1. Criar Cliente

```bash
curl -X POST http://localhost:8080/clientes \
  -H "Content-Type: application/json" \
  -d '{
    "cliente_nome": "João Silva",
    "cliente_email": "joao.silva@example.com",
    "tipo_solicitacao": "Atualização cadastral",
    "valor_patrimonio": 250000
  }'
```

**Resposta (201):**
```json
{
  "id": 1,
  "cliente_nome": "João Silva",
  "cliente_email": "joao.silva@example.com",
  "tipo_solicitacao": "Atualização cadastral",
  "valor_patrimonio": 250000,
  "status": "Aguardando Análise",
  "created_at": "2026-05-24T18:00:00Z"
}
```

### 2. Webhook — Card Atualizado

```bash
curl -X POST http://localhost:8080/webhooks/pipefy/card-updated \
  -H "Content-Type: application/json" \
  -d '{
    "event_id": "evt_123",
    "card_id": "card_456",
    "cliente_email": "joao.silva@example.com",
    "timestamp": "2026-05-18T12:00:00Z"
  }'
```

**Resposta (200):**
```json
{
  "mensagem": "evento processado com sucesso"
}
```

**Evento duplicado (409):**
```json
{
  "erro": "evento evt_123 já processado"
}
```

## Regras de Negócio

| Condição | Prioridade |
|----------|-----------|
| `valor_patrimonio >= 200.000` | `prioridade_alta` |
| `valor_patrimonio < 200.000` | `prioridade_normal` |

## Mapeamento Pipefy (GraphQL)

O pacote `internal/pipefy/` contém as mutations seguindo a [documentação oficial](https://developers.pipefy.com/reference):

### createCard ([docs](https://developers.pipefy.com/reference/cards#card-mutations))
```graphql
mutation($input: CreateCardInput!) {
  createCard(input: $input) {
    card { id title }
  }
}
```

### updateCardField ([docs](https://developers.pipefy.com/reference/fields#updating-fields-values))
```graphql
mutation($input: UpdateCardFieldInput!) {
  updateCardField(input: $input) {
    card { id }
    success
  }
}
```

O envio é simulado — o payload é logado no console. Em produção, bastaria trocar `SimulateSend` por uma chamada HTTP `POST https://api.pipefy.com/graphql`.

## Visão de Produção (AWS)

Em ambiente produtivo, a arquitetura escalaria da seguinte forma:

- **API Gateway + Lambda (Go):** Substitui o servidor Gin local. Cada endpoint vira uma função Lambda separada, com API Gateway roteando as requisições. Escala automaticamente com o volume de chamadas.
- **DynamoDB:** Substitui SQLite. Tabela `clientes` com chave primária `email` + GSI por `status` para queries. Tabela `eventos_processados` com TTL para expurgo automático de eventos antigos. DynamoDB Streams pode disparar processamento adicional em tempo real.
- **SQS + Lambda (Webhook):** O endpoint de webhook publica o evento em uma fila SQS; uma segunda Lambda consome a fila e processa (idempotência + regra de prioridade). Isso desacopla a ingestão do processamento e garante retry em caso de falha.
- **Secrets Manager:** Token de autenticação do Pipefy armazenado como secret, injetado na Lambda via variável de ambiente.
- **CloudWatch:** Logs estruturados de cada execução para tracing e alertas.

### Diagrama de fluxo

```
POST /clientes → API Gateway → Lambda CriarCliente → DynamoDB
                                                      ↓
POST /webhooks  → API Gateway → Lambda Ingestão → SQS → Lambda Processar → DynamoDB
                                                                           ↓
                                                                  Envia updateCardField → Pipefy API
```

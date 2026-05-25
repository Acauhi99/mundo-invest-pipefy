# Arquitetura Local

## Diagrama de Fluxo

```mermaid
flowchart TD
    subgraph Client["Client (curl / Postman)"]
        REQ1["POST /clientes"]
        REQ2["POST /webhooks/pipefy/card-updated"]
    end

    subgraph Gin["Gin HTTP Server (:8080)"]
        ROUTER["r.POST()
        r.POST()"]
    end

    subgraph ClienteContext["Bounded Context: Cliente"]
        direction TB
        subgraph ClienteHTTP["infrastructure/http"]
            CH["Handler.Criar()"]
        end
        subgraph ClienteApp["application"]
            CCH["CriarClienteHandler"]
            OCH["ObterClientePorEmailHandler"]
        end
        subgraph ClienteDomain["domain"]
            CL["Cliente aggregate"]
            DE["Domain Events:
            ClienteCriado
            ClienteProcessado"]
        end
        subgraph ClientePersist["infrastructure/persistence"]
            CR["SQLiteRepository"]
        end
    end

    subgraph WebhookContext["Bounded Context: Webhook"]
        direction TB
        subgraph WebhookHTTP["infrastructure/http"]
            WH["Handler.CardUpdated()"]
        end
        subgraph WebhookApp["application"]
            PCH["ProcessarCardUpdatedHandler"]
        end
        subgraph WebhookDomain["domain"]
            PE["ProcessedEvent
            CardUpdatedInput"]
        end
        subgraph WebhookPersist["infrastructure/persistence"]
            WR["SQLiteEventRepository"]
        end
    end

    subgraph ACL["Anti-Corruption Layer"]
        direction LR
        PC["pkg/pipefy/Client"]
        PM["pkg/pipefy/mutations.go
        createCard
        updateCardField"]
    end

    subgraph SharedKernel["Shared Kernel"]
        SR["pkg/shared/response.go
        APIResponse, APIError"]
    end

    subgraph DB["SQLite (mundoinvest.db)"]
        CT["clientes"]
        ET["eventos_processados"]
    end

    REQ1 --> ROUTER
    REQ2 --> ROUTER
    ROUTER --> CH
    ROUTER --> WH

    CH -->|"ShouldBindJSON() valida campos obrigatórios + email"| CCH
    CCH -->|"1. New Cliente{Status: Aguardando Análise}"| CL
    CCH -->|"2. repo.Create()"| CR
    CR -->|"INSERT INTO clientes"| CT
    CCH -->|"3. buildCreateCardPayload()"| PC
    PC -->|"SimulateSend() retorna card_sim_xxx"| CCH
    CCH -->|"4. repo.UpdateCardID()"| CR

    WH -->|"ShouldBindJSON() valida campos"| PCH
    PCH -->|"1. IsEventProcessed(event_id)"| WR
    WR -->|"SELECT COUNT(*) FROM eventos_processados"| ET
    PCH -->|"2. ObterClientePorEmail.Handle()"| OCH
    OCH -->|"repo.FindByEmail()"| CR
    CR -->|"SELECT FROM clientes WHERE email=?"| CT
    PCH -->|"3. Regra: >=200k → prioridade_alta"| PCH
    PCH -->|"4. UpdateStatusAndPriority('Processado', prioridade)"| CR
    CR -->|"UPDATE clientes SET status=?, prioridade=?"| CT
    PCH -->|"5. buildUpdateCardFieldPayload()"| PC
    PC -->|"SimulateSend() loga card_id"| PCH
    PCH -->|"6. MarkEventProcessed(event_id)"| WR
    WR -->|"INSERT INTO eventos_processados"| ET

    SR -.-> CH
    SR -.-> WH
    PM -.-> PC

    style Client fill:#f0f0f0,stroke:#999
    style Gin fill:#90EE90,stroke:#333
    style DB fill:#87CEEB,stroke:#333
    style ACL fill:#FFD700,stroke:#333
    style SharedKernel fill:#DDA0DD,stroke:#333
```

## Camadas e Responsabilidades

| Camada | Responsabilidade | Exemplos de Arquivos |
|--------|-----------------|---------------------|
| `domain/` | Aggregate root, value objects, domain events, erros de domínio | `cliente.go`, `evento.go`, `errors.go` |
| `application/` | Commands (mutações) + Queries (leituras). Orquestra o fluxo, define ports | `commands.go`, `queries.go` |
| `infrastructure/http/` | Adapter HTTP — bind JSON, chama command, mapeia HTTP status | `handler.go` |
| `infrastructure/persistence/` | Adapter de banco — implementa a port definida em application | `repository.go` |
| `pkg/pipefy/` | Anti-corruption layer — mutations GraphQL, payload builder, simulação de envio | `client.go`, `mutations.go`, `models.go` |
| `pkg/shared/` | Shared kernel — formato de resposta padronizado da API | `response.go` |

## Fluxo de Dados — Criar Cliente

```
curl POST /clientes
  → Gin Router
    → HTTP Handler: ShouldBindJSON() → valida campos obrigatórios, email, valor>0
      → CriarClienteHandler.Handle()
        1. Constrói Cliente{Status: "Aguardando Análise"}
        2. repo.Create(cliente) → INSERT INTO clientes → retorna ID + created_at
        3. buildCreateCardPayload() → mutation createCard via pkg/pipefy
        4. pipefy.SimulateSend() → loga "card_sim_xxx" no console
        5. repo.UpdateCardID(email, cardID) → UPDATE clientes SET card_id=?
        Retorna Cliente completo com ID, CardID, Status
      → HTTP Handler: shared.Success(cliente) → 201 Created
```

## Fluxo de Dados — Webhook Card Updated

```
curl POST /webhooks/pipefy/card-updated
  → Gin Router
    → HTTP Handler: ShouldBindJSON() → valida event_id, card_id, email, timestamp
      → ProcessarCardUpdatedHandler.Handle()
        1. eventRepo.IsEventProcessed(event_id)
           → Se já processado → ErrEventAlreadyProcessed → 409 Conflict
        2. clienteQry.Handle(email) → repo.FindByEmail()
           → Se não encontrado → ErrClientNotFound → 404 Not Found
        3. Regra de prioridade:
           - valor_patrimonio >= 200000 → prioridade_alta
           - valor_patrimonio < 200000  → prioridade_normal
        4. clienteUpd.UpdateStatusAndPriority("Processado", prioridade)
           → UPDATE clientes SET status=?, prioridade=?
        5. buildUpdateCardFieldPayload(cardID, prioridade) → mutations updateCardField
        6. pipefy.SimulateSend() → loga no console
        7. eventRepo.MarkEventProcessed(event_id)
           → INSERT INTO eventos_processados
        Retorna nil (sucesso)
      → HTTP Handler: shared.Success({"message": "event processed successfully"}) → 200 OK
```

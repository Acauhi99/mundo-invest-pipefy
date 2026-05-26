# CONTEXT — Mundo Invest Pipefy

## Domínio

Sistema interno de gestão de clientes e seus patrimônios investidos, com mapeamento de cards para o Pipefy (ferramenta de controle de processos). O Pipefy é simulado localmente — as mutations GraphQL estão estruturadas no código seguindo a documentação oficial, mas o envio real é substituído por um log no console.

## Glossário

| Termo | Definição |
|-------|----------|
| **Cliente** | Pessoa com patrimônio investido. Aggregate root do bounded context `client`. Atributos: Nome, Email, TipoSolicitacao, ValorPatrimonio, Status, Prioridade, CardID. |
| **TipoSolicitacao** | Classificação da solicitação do cliente (ex: "Atualização cadastral"). Campo livre, validado como required. |
| **ValorPatrimonio** | Patrimônio total investido do cliente, em reais (float64). Dispara a regra de prioridade: >= 200.000 → prioridade_alta; < 200.000 → prioridade_normal. |
| **Status** | Estado do cliente no fluxo. Valores: `"Aguardando Análise"` (inicial, após criação), `"Processado"` (após webhook de card updated). |
| **Prioridade** | Calculada na chegada do webhook com base no ValorPatrimonio. Valores: `"prioridade_alta"`, `"prioridade_normal"`. |
| **CardID** | Identificador do card correspondente no Pipefy. Gerado pela simulação (`card_sim_<timestamp>`) durante a criação do cliente. |
| **Card** | Unidade de trabalho no Pipefy. Criado via mutation `createCard`, atualizado via `updateCardField`. |
| **Webhook Card Updated** | Evento recebido do Pipefy simulando que um card foi alterado. Dispara: idempotência, regra de prioridade, atualização de status, mutation de update. |
| **Evento Processado** | Registro de idempotência. Cada `event_id` de webhook é armazenado na tabela `processed_events` para evitar processamento duplicado. |

## Bounded Contexts

### client

Gerencia o ciclo de vida do cliente: criação, consulta por email, atualização de status e card_id.

- **domain/** — `Client` (aggregate root), `ClientCreated` / `ClientProcessed` (domain events), `ErrClientNotFound`
- **application/** — `CreateClientHandler` (command), `GetClientByEmailHandler` (query). Define a port `Repository`.
- **infrastructure/persistence/** — `SQLiteRepository` (adapter que implementa `Repository`)
- **infrastructure/http/** — `Handler.Create()` (adapter HTTP, bind JSON → chama command → responde)

### webhook

Gerencia o processamento de eventos webhook recebidos: idempotência, regra de prioridade, atualização de cliente, simulação de update no Pipefy.

- **domain/** — `CardUpdatedInput` (payload), `ProcessedEvent` (registro), `ErrEventAlreadyProcessed`
- **application/** — `ProcessCardUpdatedHandler` (command). Define ports `EventRepository`, `ClientQuerier`, `ClientUpdater`.
- **infrastructure/persistence/** — `SQLiteEventRepository` (adapter para eventos processados)
- **infrastructure/http/** — `Handler.CardUpdated()` (adapter HTTP)

### pipefy (Anti-Corruption Layer)

Isola as mutations GraphQL do domínio. Contém:
- `CreateCardMutation` — mutation GraphQL `createCard`
- `UpdateCardFieldMutation` — mutation GraphQL `updateCardField`
- `Client` — implementa `PipefyClient` (interface). `SimulateSend()` retorna card_id fake; `BuildCreateCardPayload()` / `BuildUpdateCardFieldPayload()` montam o payload.
- `CreateCardInput`, `FieldAttribute`, `UpdateCardFieldInput` — modelos de entrada das mutations

### shared (Shared Kernel)

Formato padronizado de resposta da API: `APIResponse { Success, Data, Error }`, `APIError { Code, Message }`. Funções helper: `Success()`, `ValidationError()`, `NotFoundError()`, `ConflictError()`, `InternalError()`.

## Fluxos

### Criação de Cliente

```
POST /clientes
  → Handler.Create() → ShouldBindJSON (valida required+email+gt=0)
    → CreateClientHandler.Handle()
      1. Constrói Client{Status: "Aguardando Análise"}
      2. repo.Create() → INSERT clients
      3. buildCreateCardPayload() → mutation createCard
      4. pipefy.SimulateSend() → card_sim_xxx
      5. repo.UpdateCardID()
    → 201 { success: true, data: Client }
```

### Webhook Card Updated

```
POST /webhooks/pipefy/card-updated
  → Handler.CardUpdated() → ShouldBindJSON
    → ProcessCardUpdatedHandler.Handle()
      1. eventRepo.IsEventProcessed(event_id) → se sim, ErrEventAlreadyProcessed → 409
      2. clientQry.Handle(email) → se não encontrado, ErrClientNotFound → 404
      3. valor_patrimonio >= 200000 → prioridade_alta, senão prioridade_normal
      4. clientUpd.UpdateStatusAndPriority("Processado", priority)
      5. buildUpdateCardFieldPayload() → mutation updateCardField
      6. pipefy.SimulateSend()
      7. eventRepo.MarkEventProcessed(event_id)
    → 200 { success: true, data: { message: "event processed successfully" } }
```

## Regras de Negócio

| Condição | Resultado |
|----------|----------|
| `valor_patrimonio >= 200000` | `prioridade_alta` |
| `valor_patrimonio < 200000` | `prioridade_normal` |
| `event_id` já existe em `processed_events` | Bloqueia processamento → 409 Conflict |
| Email não encontrado em `clients` | 404 Not Found |

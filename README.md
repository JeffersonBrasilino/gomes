# 🎯 Gomes - Plugin de Mensageria Avançada

> Sistema completo de mensageria com suporte a **CQRS**, **Event-Driven Architecture** e múltiplos message brokers

---

## 📌 Objetivo

**Gomes** é um framework de mensageria para Go que implementa padrões de arquitetura moderna, permitindo construir sistemas distribuídos escaláveis e resilientes com:

- 📤 **Comandos Síncronos** e Assincronos
- 🔍 **Queries** para consultas de dados
- 📢 **Eventos** com múltiplos consumidores
- 👁️ **Consumidores Assincronos** com processamento paralelo
- 🔧 **Múltiplos Brokers** (Kafka, RabbitMQ, etc)
- 📊 **Observabilidade Nativa** (OpenTelemetry)
- 🛡️ **Padrões Enterprise** (Dead Letter, Retry, etc)

Gomes abstrai a complexidade de integração com message brokers, permitindo que você foque na lógica de negócio.

---

## 🚀 Resumo de Uso Rápido

### 1️⃣ Instalar & Importar

```go
import "github.com/jeffersonbrasilino/gomes"
```

### 2️⃣ Configurar & Inicializar

```go
// Registrar componentes
gomes.AddChannelConnection(
    kafka.NewConnection("kafka", []string{"localhost:9092"}),
)

gomes.AddPublisherChannel(
    kafka.NewPublisherChannelAdapterBuilder("kafka", "order.events"),
)

gomes.AddActionHandler(&CreateOrderHandler{})

// Inicializar sistema
if err := gomes.Start(); err != nil {
    panic(err)
}
defer gomes.Shutdown()
```

### 3️⃣ Usar CommandBus (Síncrono)

```go
bus, _ := gomes.CommandBus()

result, err := bus.Send(ctx, &CreateOrderCommand{
    OrderID: "ORD-001",
    Total:   199.99,
})
```

### 4️⃣ Usar EventBus (Assíncrono)

```go
eventBus, _ := gomes.EventBusByChannel("order.events")

err := eventBus.Publish(ctx, &OrderCreatedEvent{
    OrderID: "ORD-001",
})
```

### 5️⃣ Usar EventDrivenConsumer (Processor Pool)

```go
consumer, _ := gomes.EventDrivenConsumer("order-consumer-group")

err := consumer.
    WithAmountOfProcessors(4).
    WithMessageProcessingTimeout(30000).
    Run(ctx)
```

---

## 📚 Documentação Completa

### 🎓 Guias Principais

| Documento                                                      | Descrição                                                         | Para Quem            |
| -------------------------------------------------------------- | ----------------------------------------------------------------- | -------------------- |
| [**Gomes Bootstrap**](docs/gomes-bootstrap.md)                 | Como inicializar o sistema, registrar componentes e ciclo de vida | Todos que usam Gomes |
| [**Command Processing Flow**](docs/command-processing-flow.md) | Guia completo de comandos síncronos e assincronos                 | Quem usa CommandBus  |
| [**Event Processing Flow**](docs/event-processing-flow.md)     | Guia completo de eventos e broadcast para múltiplos consumidores  | Quem usa EventBus    |
| [**Event-Driven Consumer**](docs/event-driven-consumer.md)     | Configuração e tuning de consumidores com processamento paralelo  | Quem consome eventos |
| [**Kafka Channel Adapters**](docs/kafka.md)                    | Integração com Apache Kafka para publicar e consumir mensagens    | Quem usa Kafka       |
| [**RabbitMQ Channel Adapters**](docs/rabbitmq.md)              | Integração com RabbitMQ com roteamento avançado (Fanout, Topic)   | Quem usa RabbitMQ    |

---

## 🏗️ Arquitetura

```
┌─────────────────────────────────────────────────────────┐
│                     APLICAÇÃO                           │
└──────────────┬──────────────────────────────────────────┘
               │
               ├─→ [1] CommandBus (Sync/Async)
               │        └─→ Handler → Resultado
               │
               ├─→ [2] QueryBus (Sync)
               │        └─→ Handler → Dados
               │
               ├─→ [3] EventBus (Async Publish)
               │        └─→ Message Broker
               │
               └─→ [4] EventDrivenConsumer (Async Subscribe)
                        └─→ Processor Pool → Handlers


┌─────────────────────────────────────────────────────────┐
│              MESSAGE BROKERS                            │
│  ┌──────────────┐  ┌──────────────┐  ┌──────────────┐  │
│  │ Kafka        │  │ RabbitMQ     │  │ Custom       │  │
│  └──────────────┘  └──────────────┘  └──────────────┘  │
└─────────────────────────────────────────────────────────┘
```

---

## 💡 Casos de Uso

### ✅ CommandBus - Ações Síncronas

- Criar pedidos
- Transferências de dinheiro
- Alterações críticas que precisam de resposta imediata
- Validações antes de commit

**Quando usar**: Necessário resposta imediata

```go
result, err := commandBus.Send(ctx, &CreateOrderCommand{...})
if err != nil {
    return err
}
```

### ✅ EventBus - Publicação de Eventos

- Notificações de mudanças
- Sincronização entre microsserviços
- Histórico de alterações (Event Sourcing)
- Processamento em background

**Quando usar**: Pode processar assincronamente

```go
eventBus.Publish(ctx, &OrderCreatedEvent{...})
```

### ✅ EventDrivenConsumer - Processamento em Paralelo

- Processar milhares de mensagens
- Workers para diferentes tipos de evento
- Garantir entrega com retry automático
- Escalar horizontalmente

**Quando usar**: Volume alto ou operações lentas

```go
consumer.WithAmountOfProcessors(8).Run(ctx)
```

### ✅ QueryBus - Consultas de Dados

- Buscar informações
- Gerar relatórios
- Leitura de cache
- Padrão CQRS separação De Leitura/Escrita

**Quando usar**: Precisa ler dados sem alterar

```go
result, err := queryBus.Execute(ctx, &GetOrderQuery{ID: "123"})
```

---

## 🎁 Principais Características

### 1. **CQRS (Command Query Responsibility Segregation)**

- Separa **comandos** (escrita) de **queries** (leitura)
- Permite otimizar cada path independentemente
- Facilita evolução da arquitetura

### 2. **Event-Driven Architecture**

- Eventos publicados assincronamente
- Múltiplos consumidores por evento
- Desacoplamento entre componentes

### 3. **Multiple Message Brokers**

- Suporte out-of-the-box para Kafka, RabbitMQ
- Abstrair broker específico da lógica de negócio
- Trocar broker sem alterar código

### 4. **Processamento Paralelo Configurável**

- Ajuste quantidade de workers
- Timeouts por mensagem
- Estratégias de erro (continue ou stop)

### 5. **Enterprise Patterns**

- Dead Letter Queue para mensagens com erro
- Retry automático com backoff
- Idempotência garantida
- Ack/Nack configurável

### 6. **Observabilidade**

- OpenTelemetry integrado
- Traces automáticos de ponta a ponta
- CorrelationId para rastreamento
- Logging estruturado

### 7. **Type Safety**

- Generics do Go para máxima segurança
- Handlers type-safe
- Compilação garante ciclos corretos

---

## 📖 Exemplos de Padrões

### Pattern 1: Comando com Resposta Imediata

```go
// Enviar comando
commandBus, _ := gomes.CommandBus()
result, err := commandBus.Send(ctx, &CreateOrderCommand{
    OrderID: "ORD-001",
    Items:   []Item{...},
})
if err != nil {
    // Falhou, informar cliente imediatamente
    return err
}
// ✅ Sucesso, responder ao cliente
```

### Pattern 2: Evento Assincronamente Processado

```go
// Publicar evento (instantâneo)
eventBus, _ := gomes.EventBusByChannel("order.events")
err := eventBus.Publish(ctx, &OrderCreatedEvent{OrderID: "ORD-001"})
// ✅ Evento publicado, retornar ao cliente

// Em background - múltiplos consumidores processam
// Consumer 1: Verificar estoque
// Consumer 2: Enviar email
// Consumer 3: Atualizar dashboard
```

### Pattern 3: Processamento de Alta Carga

```go
// Consumer com múltiplos workers
consumer, _ := gomes.EventDrivenConsumer("payment-processor")
consumer.
    WithAmountOfProcessors(16).           // 16 workers paralelos
    WithMessageProcessingTimeout(60000).  // Timeout 60s
    WithStopOnError(false).               // Continuar em erro
    Run(ctx)

// Processa 16 mensagens em paralelo
```

### Pattern 4: CQRS Completo

```go
// ========== LADO DE ESCRITA (Commands) ==========
commandBus.Send(ctx, &CreateOrderCommand{...})

// ========== LADO DE LEITURA (Queries) ==========
queryBus.Execute(ctx, &GetOrdersQuery{UserID: "123"})

// Podem usar diferentes stores (DB, Cache, ElasticSearch)
```

---

## 🔄 Fluxo Típico de Uma Aplicação

```
1. INICIALIZAÇÃO (Boot)
   ↓
   gomes.AddChannelConnection(...)
   gomes.AddPublisherChannel(...)
   gomes.AddConsumerChannel(...)
   gomes.AddActionHandler(...)
   gomes.Start()
   ↓

2. OPERAÇÃO
   ↓
   [HTTP Request] → CommandBus.Send() → [Response]
            ↓
   [Evento Publicado] → EventBus.Publish() → [Fire & Forget]
            ↓
   [Background] ← EventDrivenConsumer → [Message Broker]
   ↓

3. ENCERRAMENTO (Graceful Shutdown)
   ↓
   gomes.Shutdown()
   └─ Stop consumers
   └─ Close channels
   └─ Disconnect brokers
```

---

## 📊 Comparação: Quando Usar Cada Bus

| Aspecto                    | CommandBus     | QueryBus    | EventBus     | Consumer              |
| -------------------------- | -------------- | ----------- | ------------ | --------------------- |
| **Síncrono**               | ✅ Sim         | ✅ Sim      | ❌ Não       | ❌ Não                |
| **Retorna resultado**      | ✅ Sim         | ✅ Sim      | ❌ Não       | ❌ Baseado no handler |
| **Múltiplos handlers**     | ❌ Não         | ❌ Não      | ✅ Sim       | ✅ Sim                |
| **Com lado em background** | ❌ Síncrono    | ❌ Síncrono | ✅ Sim       | ✅ Sim                |
| **Melhor para**            | Ações críticas | Consultas   | Notificações | Alto volume           |
| **Timeout padrão**         | ~5s            | ~5s         | ~30s         | Configurável          |

---

## ⚙️ Configuração Básica

### Setup Mínimo

```go
func main() {
    // Conectar
    gomes.AddChannelConnection(
        kafka.NewConnection("kafka", []string{"localhost:9092"}),
    )

    // Publicar
    gomes.AddPublisherChannel(
        kafka.NewPublisherChannelAdapterBuilder("kafka", "events"),
    )

    // Consumir
    gomes.AddConsumerChannel(
        kafka.NewConsumerChannelAdapterBuilder("kafka", "events", "group"),
    )

    // Handler
    gomes.AddActionHandler(&MyHandler{})

    // Start
    gomes.Start()
    defer gomes.Shutdown()

    // Usar
    bus, _ := gomes.CommandBus()
    bus.Send(ctx, &MyCommand{})
}
```

### Setup com Observabilidade

```go
func main() {
    // Habilitar traces ANTES de Start()
    gomes.EnableOtelTrace()

    // ... resto da configuração ...

    gomes.Start()
    defer gomes.Shutdown()

    // Todos os comandos e eventos são rastreados automaticamente
}
```

---

## 🆘 Troubleshooting Rápido

| Problema                            | Causa                   | Solução                               |
| ----------------------------------- | ----------------------- | ------------------------------------- |
| "failed to get default command bus" | Não chamou Start()      | Chamar `gomes.Start()`                |
| "handler for X already exists"      | Registrou handler 2x    | Registrar apenas 1x antes de Start()  |
| "consumer already exists"           | Consumer nome duplicado | Usar nomes únicos                     |
| "connection refused"                | Broker offline          | Iniciar Kafka/RabbitMQ                |
| "handler not found"                 | Handler não registrado  | `AddActionHandler()` antes de Start() |

---

## 📚 Mapa de Documentação

```
📂 docs/
│
├─ 🎯 gomes-bootstrap.md
│  │  Inicializar sistema, ciclo de vida, setup
│  │  👥 Para: Todos
│  └─ 📖 22 seções, exemplos de setup completo
│
├─ 📤 command-processing-flow.md
│  │  Comandos síncronos e assincronos
│  │  👥 Para: Quem envia comandos
│  └─ 📖 Fluxo detalhado, 9 exemplos, troubleshooting
│
├─ 📢 event-processing-flow.md
│  │  Eventos, broadcast, múltiplos consumidores
│  │  👥 Para: Quem publica e consome eventos
│  └─ 📖 Fluxo detalhado, 5 exemplos, padrões
│
├─ 👁️ event-driven-consumer.md
│  │  Configurar consumer, processamento paralelo, tuning
│  │  👥 Para: Quem processa eventos em background
│  └─ 📖 Métodos, 6 exemplos, tuning por workload
│
├─ 🔗 kafka.md
│  │  Integração com Apache Kafka
│  │  👥 Para: Quem usa Kafka como message broker
│  └─ 📖 Componentes, 20+ métodos, configuração, troubleshooting
│
└─ 🔗 rabbitmq.md
   │  Integração com RabbitMQ
   │  👥 Para: Quem faz roteamento avançado
   └─ 📖 4 Exchange types, 15+ métodos, Fanout/Topic/Direct, patterns
```

---

## 🎓 Roteiro de Aprendizado

### 👶 Iniciante

1. Leia [**Gomes Bootstrap**](docs/gomes-bootstrap.md) - Entender ciclo de vida
2. Leia [**Command Processing Flow**](docs/command-processing-flow.md) - Começar simples
3. Rode o exemplo em `examples/cqrs/`

### 👨‍💻 Intermediário

1. Leia [**Event Processing Flow**](docs/event-processing-flow.md) - Entender eventos
2. Leia [**Event-Driven Consumer**](docs/event-driven-consumer.md) - Processar em paralelo
3. Leia [**Kafka Channel Adapters**](docs/kafka.md) ou [**RabbitMQ Channel Adapters**](docs/rabbitmq.md) - Integrar com broker
4. Rode exemplos em `examples/event_driven_consumer/`

### 🚀 Avançado

1. Combine todos os padrões (Bootstrap + CQRS + Kafka/RabbitMQ)
2. Implemente CQRS completo com broker escolhido
3. Configure observabilidade end-to-end
4. Otimize consumer para seus workloads
5. Implemente padrões enterprise (DLQ, Retry, Event Sourcing, etc)

---

## 📦 Estrutura do Plugin

```
gomes/
├── bus/                      # Buses (Command, Query, Event)
│   ├── command_bus.go
│   ├── query_bus.go
│   └── event_bus.go
│
├── channel/                  # Message Brokers
│   ├── kafka/
│   └── rabbitmq/
│
├── message/                  # Mensagens e Handlers
│   ├── message.go
│   ├── handler/              # Action handlers
│   ├── endpoint/             # Consumers, Gateways
│   ├── adapter/              # Channel adapters
│   ├── channel/              # Point-to-point, PubSub
│   └── router/               # Message routers
│
├── container/                # Injeção de dependências
│
├── otel/                     # OpenTelemetry
│
├── docs/                     # 📚 Documentação
│   ├── gomes-bootstrap.md
│   ├── command-processing-flow.md
│   ├── event-processing-flow.md
│   └── event-driven-consumer.md
│
└── examples/                 # 💡 Exemplos
    ├── cqrs/
    ├── event_driven_consumer/
    └── message_publisher/
```

---

## 🔗 Recursos Úteis

### Documentação Interna

- [GoDoc](gomes.go): Documentação código-fonte
- [Gomes Bootstrap](docs/gomes-bootstrap.md): Inicialização do sistema, ciclo de vida e setup
- [Command Bus Doc](docs/command-bus.md): Especificação e guia de CommandBus
- [Query Bus Doc](docs/query-bus.md): Especificação e guia de QueryBus
- [Event Bus Doc](docs/event-bus.md): Especificação e guia de EventBus
- [Event-Driven Consumer](docs/event-driven-consumer.md): Configuração, tuning e padrões de consumo
- [Kafka Channel Adapters](docs/kafka.md): Integração com Apache Kafka para publicação e consumo
- [RabbitMQ Channel Adapters](docs/rabbitmq.md): Integração com RabbitMQ com roteamento avançado

### Recursos Externos

- [CQRS Pattern](https://martinfowler.com/bliki/CQRS.html) - Martin Fowler
- [Event Sourcing](https://martinfowler.com/eaaDev/EventSourcing.html) - Martin Fowler
- [Event-Driven Architecture](https://www.aws.amazon.com/event-driven-architecture/) - AWS
- [Enterprise Integration Patterns](https://www.enterpriseintegrationpatterns.com/) - Hohpe & Woolf
- [OpenTelemetry](https://opentelemetry.io/docs/getting-started/) - OTEL Docs
- [Go Context](https://go.dev/blog/context) - Official Go Blog

---

## 💭 FAQ

### P: Qual broker devo usar, Kafka ou RabbitMQ?

**R**: Depende do caso:

- **Kafka**: Alto volume, reprocessamento, Event Sourcing
- **RabbitMQ**: Simplicidade, routing complexo, baixa latência

Gomes abstrai isso - mude de broker sem alterar código.

### P: Posso usar CommandBus sem Message Broker?

**R**: Sim! CommandBus padrão usa canal interno. Use `gomes.CommandBus()` sem broker.
Para escalar, registre um publisher channel com Kafka/RabbitMQ.

### P: Como garantir que evento foi processado?

**R**: Use handles com Ack explícito. Gomes rastreia com OpenTelemetry.
Para garantia forte, implemente idempotência no handler.

### P: Quantos processors (workers) devo usar?

**R**: Regra geral:

- **I/O bound**: 2-4x quantidade de CPUs
- **CPU bound**: 1x quantidade de CPUs
- **Teste**: Use `WithAmountOfProcessors(n)` e monitore

Veja [Event-Driven Consumer Tuning](docs/event-driven-consumer.md#-tuning-por-workload) para detalhes.

### P: Como debugar se mensagem não foi processada?

**R**: 1. Ative logs 2. Use ShowActiveEndpoints() 3. Habilite traces OTEL 4. Verifique se handler está registrado 5. Verifique Dead Letter Queue

---

## 📞 Contato & Suporte

- **Documentação**: Veja pasta [docs/](docs/)
- **Exemplos**: Veja pasta [examples/](examples/)
- **Issues**: Reporte problemas no repositório
- **Especialista Backend**: @jeffersonbrasilino

---

**Versão**: 1.0+  
**Última Atualização**: 16 de fevereiro de 2026  
**Status**: ✅ Produção

---

## 🎉 Próximos Passos

1. **Iniciante**: Leia [Gomes Bootstrap](docs/gomes-bootstrap.md)
2. **Experimente**: Rode exemplo em `examples/cqrs/`
3. **Integre**: Use Gomes na sua aplicação
4. **Evolua**: Explore padrões avançados

**Bem-vindo ao Gomes!** 🚀

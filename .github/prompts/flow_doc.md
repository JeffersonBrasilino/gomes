---
promptType: documentation
purpose: "Padronizar a criação de documentações de fluxos do Gomes"
targetAudience: "Desenvolvedores, com foco em desenvolvedores júnior"
outputLocation: "docs/{flow-name-dash-case}.md"
---

# 📋 Prompt: Documentação de Fluxos - Gomes Plugin

## 📌 Objetivo

Gerar documentação padronizada e de alta qualidade para fluxos do plugin **Gomes**, mantendo consistência visual, estrutural e didática em todas as documentações.

---

## 🎯 Instruções de Entrada

O usuário fornecerá:

1. **Nome do Fluxo** (ex: "event-driven-consumer", "retry-handler", "dead-letter-channel")
2. **Contexto** (breve descrição do que o fluxo faz)
3. **Objetivo** (para que serve, quando usar)

---

## 📂 Estrutura do Arquivo

### Localização

- **Pasta**: `docs/`
- **Nome do Arquivo**: `{flow-name-dash-case}.md` (ex: `event-driven-consumer.md`)
- **Encoding**: UTF-8

### Exemplo de Conventions

```
event-driven-consumer.md  ✅ Correto
EventDrivenConsumer.md    ❌ Incorreto
event_driven_consumer.md  ❌ Incorreto
```

---

## 📄 Estrutura de Conteúdo

### 1️⃣ Cabeçalho e Introdução

```markdown

## 📖 O que é?

[Descrição detalhada do que é o fluxo, seu propósito e quando deve ser utilizado. 2-3 parágrafos explicativos, ajustado para desenvolvedores júnior.]

### Quando Usar

- ✅ Caso 1: [Descrição]
- ✅ Caso 2: [Descrição]
- ✅ Caso 3: [Descrição]

### Quando NÃO Usar

- ❌ Caso 1: [Descrição]
- ❌ Caso 2: [Descrição]
```

---

### 2️⃣ Características Principais

```markdown
## 🎁 Características Principais

| Característica       | Descrição       |
| -------------------- | --------------- |
| **Característica 1** | Breve descrição |
| **Característica 2** | Breve descrição |
| **Característica 3** | Breve descrição |
| **Característica 4** | Breve descrição |
```

---

### 3️⃣ Descrição Detalhada da Implementação

```markdown
## 🔧 Implementação Detalhada

### Arquitetura

[Explicação clara de como o fluxo é implementado internamente. Descreva:

- Componentes principais envolvidos
- Fluxo de dados
- Responsabilidades de cada parte
- Interações entre componentes]

### Características Técnicas

- **Thread-Safe**: [Sim/Não] - [Explicação]
- **Assíncrono**: [Sim/Não] - [Explicação]
- **Idempotente**: [Sim/Não] - [Explicação]
- **Configurável**: [Sim/Não] - [Explicação]
```

---

### 4️⃣ Documentação de Métodos Públicos

````markdown
## 📚 Métodos Públicos

[Para cada método público, incluir:]

### WithConfigParam(context.Context, config interface{}) error

**Descrição**: [Extraída/Melhorada da documentação GoDoc]

**Parâmetros**:

- `context.Context`: Contexto para cancelamento
- `config`: Configuração do componente

**Retorno**:

- `error`: Erro se alguma validação falhar

**Exemplo**:

```go
consumer.WithAmountOfProcessors(5)
```
````

---

### RunFlows(ctx context.Context) error

**Descrição**: Inicia o processamento do fluxo

**Parâmetros**:

- `ctx context.Context`: Contexto para controle de ciclo de vida

**Retorno**:

- `error`: Erro durante execução

**Exemplo**:

```go
if err := consumer.Run(ctx); err != nil {
    log.Fatal(err)
}
```

````

---

### 5️⃣ Diagrama de Componentes

```markdown
## 🏗️ Diagrama de Componentes

[Diagrama Mermaid mostrando a arquitetura e interação entre componentes]

\`\`\`mermaid
graph TB
    Client["Cliente"]
    Bus["CommandBus<br/>(Orquestrador)"]
    Router["Router<br/>(Roteador)"]
    Handler["Handler<br/>(Processador)"]
    Channel["Channel<br/>(Transmissão)"]

    Client -->|Comando| Bus
    Bus -->|Roteia| Router
    Router -->|Encontra| Handler
    Handler -->|Processa| Função["Função de Negócio"]
    Função -->|Resultado| Handler
    Handler -->|Publica| Channel
    Channel -->|Entrega| Output["Saída"]

    style Client fill:#e1f5e1
    style Bus fill:#e3f2fd
    style Router fill:#fff3cd
    style Handler fill:#f8d7da
    style Channel fill:#e2e3e5
\`\`\`

**Componentes Principais**:

- **Componente A**: [Descrição breve da responsabilidade]
- **Componente B**: [Descrição breve da responsabilidade]
- **Componente C**: [Descrição breve da responsabilidade]
````

---

### 6️⃣ Diagrama de Execução

```markdown
## 🔄 Diagrama de Execução

[Diagrama Mermaid mostrando o fluxo de execução passo a passo]

\`\`\`mermaid
sequenceDiagram
actor User as Usuário
participant App as Aplicação
participant Fluxo as Fluxo
participant Handler as Handler
participant Result as Resultado

    User->>App: Inicia Fluxo
    App->>Fluxo: Start()
    Fluxo->>Fluxo: Valida Configuração
    Fluxo->>Handler: Processa
    Handler->>Handler: Executa Lógica
    Handler-->>Fluxo: Retorna Resultado
    Fluxo-->>App: Resultado
    App-->>User: Sucesso/Erro

\`\`\`

**Fluxo de Execução**:

1. **Passo 1**: [Descrição]
2. **Passo 2**: [Descrição]
3. **Passo 3**: [Descrição]
4. **Passo N**: [Descrição]
```

---

### 7️⃣ Exemplo de Uso Prático

````markdown
## 💡 Exemplo de Uso Prático

[Referência ao arquivo de exemplo do projeto e adaptação para documentação]

### Setup Básico

```go
package main

import (
    "context"
    "log/slog"
    "github.com/jeffersonbrasilino/gomes"
    kafka "github.com/jeffersonbrasilino/gomes/channel/kafka"
)

// Definir a estrutura de dados
type MeuComando struct {
    ID    string
    Dados string
}

func (c *MeuComando) Name() string {
    return "meuComando"
}

// Definir o handler
type MeuHandler struct{}

func (h *MeuHandler) Handle(
    ctx context.Context,
    cmd *MeuComando,
) (any, error) {
    slog.Info("Processando", "id", cmd.ID)
    return map[string]string{"status": "sucesso"}, nil
}

func main() {
    // 1. Registrar componentes
    gomes.AddChannelConnection(
        kafka.NewConnection("kafka", []string{"localhost:9092"}),
    )

    gomes.AddPublisherChannel(
        kafka.NewPublisherChannelAdapterBuilder("kafka", "meu-topico"),
    )

    gomes.AddActionHandler(&MeuHandler{})

    // 2. Iniciar sistema
    if err := gomes.Start(); err != nil {
        panic(err)
    }
    defer gomes.Shutdown()

    // 3. Usar o fluxo
    bus, _ := gomes.CommandBus()
    result, _ := bus.Send(context.Background(), &MeuComando{
        ID:    "123",
        Dados: "teste",
    })

    slog.Info("Resultado:", "result", result)
}
```
````

\`\`\`

### Configuração Avançada

\`\`\`go
consumer, \_ := gomes.EventDrivenConsumer("meu-consumer")

consumer.
WithAmountOfProcessors(10).
WithMessageProcessingTimeout(30000).
WithStopOnError(false).
Run(ctx)
\`\`\`

````

---

### 8️⃣ Boas Práticas

```markdown
## ✅ Boas Práticas

- ✅ [Prática 1]: [Descrição e exemplo]
- ✅ [Prática 2]: [Descrição e exemplo]
- ✅ [Prática 3]: [Descrição e exemplo]

### Erros Comuns a Evitar

- ❌ [Erro 1]: [Por que evitar e exemplo correto]
- ❌ [Erro 2]: [Por que evitar e exemplo correto]
- ❌ [Erro 3]: [Por que evitar e exemplo correto]
````

---

### 9️⃣ Troubleshooting

````markdown
## 🔍 Troubleshooting

### Problema: [Descrição do Problema]

**Sintomas**:

- Sintoma 1
- Sintoma 2

**Causa**: [Explicação]

**Solução**:

```go
// Código de solução
```
````

### Problema: [Outro Problema]

...

````

---

### 🔟 Referências e Links

```markdown
## 📚 Referências

- [Link 1](url): Descrição
- [Link 2](url): Descrição
- [Documentação GoDoc](url): Link para GoDoc
- [Exemplo Completo](../../examples/flow-name/main.go): Arquivo de exemplo

---

**Última Atualização**: [Data]
**Status**: ✅ Produção
**Versão do Gomes**: v1.0+
````

---

## 📏 Diretrizes de Escrita

### Linguagem e Tom

- **Linguagem**: Clara, concisa e amigável
- **Público**: Desenvolvedores, especialmente juniores
- **Tom**: Didático, explicativo e não condescendente
- **Terminologia**: Explicar termos técnicos quando necessário

### Exemplos de Código

✅ **Bom**:

```markdown
Para usar o fluxo, você precisa registrar o handler:

\`\`\`go
gomes.AddActionHandler(&MeuHandler{})
\`\`\`

Isso permite que o Gomes saiba qual handler executar quando o comando é enviado.
```

❌ **Ruim**:

```markdown
Use AddActionHandler.
```

### Diagramas

- Usar **Mermaid** para todos os diagramas
- Incluir legendas explicativas
- Manter consistência visual com outros diagramas
- Adicionar cores: verde (sucesso), vermelho (erro), azul (processo), amarelo (decisão)

### Estrutura Visual

- Usar emojis em títulos (🎯, 🔧, 💡, ✅, ❌, 📚, etc.)
- Usar negrito `**texto**` para destacar conceitos-chave
- Usar listas com bullets `-` ou números `1.`
- Usar tabelas para comparações
- Usar código inline com backticks para nomes de funções/variáveis

---

## 🎨 Template de Início

```markdown
# 🎯 Nome do Fluxo

**Tipo**: [Padrão/Componente]  
**Objetivo**: Uma linha descrevendo o propósito  
**Status**: ✅ Produção

## 📖 O que é?

[2-3 parágrafos explicativos dirigidos a desenvolvedores júnior]

### Quando Usar

- ✅ [Caso de uso 1]
- ✅ [Caso de uso 2]

## 🎁 Características Principais

| Característica | Descrição |
| -------------- | --------- |
| Feature 1      | Descrição |
| Feature 2      | Descrição |

[Continue com os outros tópicos...]
```

---

## 📋 Checklist de Qualidade

Antes de considerar a documentação completa:

- [ ] Título claro e descritivo
- [ ] Objetivo explicado em uma linha
- [ ] Seção "O que é?" com 2-3 parágrafos
- [ ] Casos de uso (quando usar / não usar)
- [ ] Características principais tabeladas
- [ ] Implementação detalhada
- [ ] Todos os métodos públicos documentados
- [ ] Diagrama de componentes em Mermaid
- [ ] Diagrama de execução em Mermaid
- [ ] Exemplo de uso prático completo
- [ ] Código bem comentado nos exemplos
- [ ] Boas práticas identificadas
- [ ] Erros comuns evitados
- [ ] Troubleshooting para problemas comuns
- [ ] Links e referências úteis
- [ ] Verificação de ortografia e gramática
- [ ] Consistência com outras documentações
- [ ] Tom didático e amigável mantido
- [ ] Emojis usados consistentemente
- [ ] Formatação markdown correta

---

## 🚀 Como Usar Este Prompt

1. **Forneça o nome do fluxo** que deseja documentar
2. **Especifique o contexto/objetivo** do fluxo
3. **Indique se há um arquivo de exemplo** (`examples/` ou `cmd/`)
4. **Revise os GoDoc** dos métodos públicos
5. **Gere a documentação** usando esta estrutura padronizada
6. **Salve em** `docs/{flow-name-dash-case}.md`

---

## 📝 Exemplo de Uso do Prompt

**Input do Usuário:**

```
Quero documentar o fluxo de "Event-Driven Consumer"
Existe um exemplo em examples/event_driven_consumer/main.go
Objetivo: Explicar como usar o consumer assíncrono para processar mensagens
```

**Output Esperado:**

```
File: docs/event-driven-consumer.md

# 🎯 Event-Driven Consumer

**Tipo**: Padrão de Consumo
**Objetivo**: Processar mensagens de forma assíncrona com workers paralelos
**Status**: ✅ Produção

[Documentação completa seguindo a estrutura acima]
```

---

**Versão do Prompt**: 1.0  
**Data de Criação**: 16 de fevereiro de 2026  
**Mantido por**: Especialista em Desenvolvimento Backend (Gomes)

---
agent: agent
description: Refactor the provided code files to improve readability, maintainability, and performance.
---

# Refatoração de Código Go: Formatação e Documentação

**Objetivo**: Analisar e ajustar arquivos de código Go para seguir rigorosamente as melhores práticas da linguagem, focando em legibilidade, manutenibilidade e conformidade com convenções oficiais.

## 1. Análise Inicial

- Leia o arquivo fornecido completamente.
- Identifique violações às regras abaixo.
- Priorize ajustes em elementos públicos (funções, tipos, constantes) para impacto máximo.

## 2. Formatação de Código

### Regras Principais

- **Go Proverbs**: Aplique princípios como "Clarity is better than cleverness" e "Less is more". Evite código complexo; prefira simplicidade e clareza.
- **Ordem de Componentes**: Estruture o arquivo nesta sequência exata:
  1. Declaração do pacote (`package`)
  2. Imports (organizados: stdlib, third-party, locais; use `goimports`)
  3. Constantes
  4. Variáveis globais
  5. Tipos (structs, interfaces, etc.)
  6. Funções (públicas primeiro, privadas depois)
- **Limite de Linha**: Máximo de **100 caracteres** por linha. Quebre linhas longas em chamadas de função, strings ou operadores, mantendo indentação consistente.
- **Outros**: Use `go fmt` para alinhamento automático. Evite tabs misturados com espaços.

### Exemplos

- **Antes**: `fmt.Errorf("erro longo demais para caber em uma linha")`
- **Depois**:
  ```go
  fmt.Errorf(
      "erro longo demais para caber em uma linha",
  )
  ```

## 3. Documentação GoDoc

### Estrutura Geral

- Use comentários `//` (não `/* */`).
- Inicie com o nome do elemento (função, tipo, etc.).
- Seja conciso, mas completo.

### Documentação de Arquivo (Package Comment)

- Primeiro comentário do arquivo, logo após `package`.
- Descreva:
  - **Intenção**: Propósito geral do pacote.
  - **Objetivo**: Funcionalidades principais e contexto de uso.
- Exemplo:
  ```go
  // Package container provides a generic container implementation for managing
  // key-value pairs with thread-safe operations. It supports setting, getting,
  // replacing, and removing items, with error handling for duplicate keys and
  // missing items.
  package container
  ```

### Documentação de Funções e Métodos Públicos

- Para cada `func` ou método exportado (iniciando com maiúscula), adicione comentário antes.
- Estrutura:
  - **Intenção**: O que faz (frase curta).
  - **Parâmetros**: Descrição de cada um (tipo e propósito).
  - **Retorno**: Descrição do valor retornado e erros possíveis.
  - **Comportamento**: Side effects, panics ou condições especiais.
- Exemplo:
  ```go
  // NewGenericContainer creates a new instance of a generic container.
  // It initializes an empty map for storing key-value pairs.
  // Returns a pointer to the new container.
  func NewGenericContainer[K comparable, T any]() *genericContainer[K, T]
  ```

### Documentação de Tipos e Interfaces

- Para structs, interfaces e tipos exportados, descreva propósito e uso.
- Exemplo:
  ```go
  // Container defines the interface for a generic key-value container.
  // It provides methods to manage items with error handling.
  type Container[K comparable, T any] interface {
      // Set adds a new item. Returns error if key exists.
      Set(key K, item T) error
  }
  ```

## 4. Diretrizes Gerais

- **Idioma**: Toda documentação em **inglês** (padrão GoDoc global).
- **Padrão GoDoc**: Siga https://go.dev/blog/godoc. Use frases completas, pontuação adequada.
- **Clareza e Objetividade**: Evite jargões; seja direto. Não repita informações óbvias.
- **Completude**: Documente 100% dos elementos públicos. Ignore privados (iniciam com minúscula).
- **Validação**: Após ajustes, execute `go doc` para verificar se gera documentação correta.
- **Performance**: Embora o foco seja formatação/doc, remova código redundante se identificado (e.g., loops desnecessários).

## 5. Processo de Aplicação

1. Identifique arquivos a ajustar (fornecidos no contexto).
2. Aplique formatação primeiro (`go fmt`).
3. Adicione/Atualize documentação.
4. Teste compilação (`go build`).
5. Revise manualmente para conformidade.

Este prompt garante código Go profissional, legível e bem-documentado, facilitando manutenção e colaboração.

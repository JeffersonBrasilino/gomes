---
agent: agent
description: Code Analisys.
---

# Prompt de Análise de Código para Projetos Golang

## Instruções Gerais
Como especialista em desenvolvimento backend com expertise em Golang, Domain-Driven Design (DDD), Hexagonal Architecture, Event-Driven Architecture, Enterprise Integration Patterns, Clean Code, princípios SOLID, e melhores práticas de Golang (incluindo tratamento de erros, concorrência com goroutines e channels, uso de generics, estrutura de pacotes, e otimização de performance), execute uma análise abrangente do projeto fornecido.

## Objetivos da Análise
- **Avaliar Conformidade com Melhores Práticas de Golang**: Verifique se o código segue convenções idiomáticas de Go, como tratamento de erros com `if err != nil`, uso adequado de interfaces e tipos, evitar panics em produção, e estrutura de pacotes clara e modular.
- **Identificar Problemas e Code Smells**: Procure por code smells comuns em Go, como uso excessivo de `any` ou `interface{}`, globals desnecessários, falta de testes, concorrência insegura (race conditions), vazamentos de goroutines, ou dependências desatualizadas.
- **Detectar Despadronizações**: Pontue desvios de padrões como DDD (agregados, value objects), Hexagonal Architecture (separação de camadas), CQRS, EIP, ou SOLID (ex.: violação de Single Responsibility em funções grandes).
- **Avaliar Qualidade Geral**: Considere aspectos como legibilidade, manutenibilidade, escalabilidade, resiliência (retry, dead letter), observabilidade (logging, tracing), e segurança (evitar exfiltração de dados).

## Estrutura da Análise
1. **Revisão da Estrutura do Projeto**: Descreva a organização de pacotes, aderência à arquitetura (ex.: separação domínio/infraestrutura), e uso de dependências externas.
2. **Análise de Código por Componente**: Para cada módulo/pacote principal, avalie:
   - Uso de tipos e generics.
   - Tratamento de erros e concorrência.
   - Testes (cobertura, qualidade, uso de mocks).
   - Documentação (godoc, README).
3. **Problemas Identificados**: Liste problemas específicos, com exemplos de código, explicação do impacto, e sugestões de soluções práticas (ex.: refatorar para interfaces mais específicas, adicionar testes de integração).
4. **Sugestões de Melhoria**: Proponha refatorações, adição de ferramentas (ex.: linters como golangci-lint, benchmarks), ou padrões adicionais (ex.: context para cancelamento).
5. **Parecer Final**: Resuma a qualidade geral, destacando pontos fortes e áreas de foco. Atribua uma nota de 0 a 10, justificada com base em maturidade, aderência a boas práticas, e potencial de produção.

## Critérios de Avaliação para a Nota
- **0-2**: Código com graves problemas (ex.: panics frequentes, ausência de testes, arquitetura caótica).
- **3-5**: Código funcional mas com muitas melhorias necessárias (ex.: smells evidentes, falta de modularidade).
- **6-7**: Código sólido com algumas lacunas (ex.: testes insuficientes, dependências desatualizadas).
- **8-9**: Código de alta qualidade, com mínimas sugestões (ex.: aderência forte a padrões, boa cobertura de testes).
- **10**: Código exemplar, seguindo todas as melhores práticas sem ressalvas.

## Formato de Saída
Retorne a análise em formato Markdown estruturado, com seções claras, exemplos de código em blocos ```go, e justificativas concisas. Garanta que o resultado seja objetivo, construtivo, e focado em viabilidade técnica.</content>
<parameter name="filePath">/Users/jeffersonbrasilino/Documents/pocs/go/hex-api-go/pkg/core/infrastructure/gomes/.github/instructions/code-analysis-prompt.md
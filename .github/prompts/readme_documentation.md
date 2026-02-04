---
agent: agent
description: Criar ou atualizar documentação README.md completa e estruturada
---


# Documentação README.md

Crie ou atualize a documentação README.md do projeto seguindo as especificações abaixo.

## Processo Inicial

**Sempre pergunte** ao usuário se deseja:
- **Atualizar ou criar o arquivo completo**
- **Criar ou atualizar apenas um tópico específico**

## Documentação Completa

Quando a resposta for **completo**:

### Criação/Atualização do Arquivo
- Se o arquivo `README.md` **não existir**, gere-o do zero
- Se o arquivo `README.md` **já existir**, atualize-o mantendo o conteúdo relevante

### Estrutura do Índice

A documentação deve conter um índice completo com os seguintes tópicos:

#### 1. Visão Geral
- Resumo do objetivo do plugin
- Características principais
- Padrões e abordagens utilizadas
- Estrutura de pastas do plugin

#### 2. Bootstrap
- Como registrar componentes
- Como iniciar o plugin gomes
- Como encerrar o plugin gomes

#### 3. Componentes Principais
- **Diagrama de Fluxo**: Expressar a relação entre os componentes
- **Diagrama de Execução**: Mostrar o fluxo de execução dos componentes

#### 4. CQRS
- Descrição detalhada da implementação do padrão CQRS
- Exemplos de uso práticos
- Diagrama de componentes
- Diagrama de execução

#### 5. Padrões de Publicação
Crie sub-índices para:
- **Comandos**: Descrição e exemplos
- **Queries**: Descrição e exemplos
- **Eventos**: Descrição e exemplos

**Referências**:
- Use a pasta do plugin como referência
- Use o arquivo `cmd/message_publisher/main.go` para entender e criar exemplos
- Descreva com base na documentação GoDoc o que cada método do fluxo de publicação faz

#### 6. Padrões de Consumo
Crie sub-índices para:
- **Event-Driven**: Descrição detalhada
- **Polling**: Descrição detalhada

**Referências**:
- Use a pasta do plugin como referência
- Use o arquivo `cmd/event_driven_consumer/main.go` para entender e criar exemplos
- Faça uma **comparação entre as duas abordagens** destacando prós e contras de cada uma
- Descreva com base na documentação GoDoc o que cada método do fluxo de consumo faz

#### 7. Resiliência
Crie sub-índices para:
- **Retry**: Descrição detalhada do funcionamento
- **Dead Letter**: Descrição detalhada do funcionamento

**Para cada um**:
- Use diagramas de fluxo para expressar o funcionamento
- Use diagramas de execução para auxiliar na explicação

#### 8. Kafka
- Descrição detalhada de como a configuração funciona
- Descrição detalhada de como o uso do driver Kafka implementado no plugin funciona

#### 9. RabbitMQ
- Descrição detalhada de como a configuração funciona
- Descrição detalhada de como o uso do driver RabbitMQ implementado no plugin funciona

## Documentação por Tópico

Quando a resposta for **tópico**:

1. **Pergunte** em que posição do índice o novo tópico deve ser criado
2. **Atualize o índice** incluindo o novo tópico
3. **Crie o tópico** na posição especificada

## Diretrizes de Escrita

- **Linguagem**: Clara e amigável
- **Público-alvo**: Desenvolvedores, com foco em desenvolvedores júnior
- **Tom**: Didático e explicativo
- **Exemplos**: Sempre inclua exemplos práticos de código quando relevante
- **Diagramas**: Use diagramas para facilitar o entendimento quando apropriado
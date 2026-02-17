---
agent: agent
description: Generate unit tests for the provided code files.
---

# Geração de Testes Unitários

Analise o arquivo fornecido e gere um arquivo de teste unitário completo seguindo as especificações abaixo.

## Cobertura de Testes

- **Cobertura Total**: O teste deve cobrir **100% das linhas de código**
- **Validação Completa**: Cada método e função deve ter seus parâmetros de entrada e saída validados
- **Casos de Sucesso e Falha**: Gere testes para ambos os cenários quando aplicável

## Estrutura de Arquivos

- **Um arquivo de teste por arquivo de código**: Gere um único arquivo `*_test.go` por arquivo analisado
- **Nomenclatura**: Use o nome do arquivo analisado como base, adicionando o sufixo `_test.go`
- **Namespace**: O namespace do arquivo de teste deve ser o mesmo do arquivo analisado seguido de `_test`

## Restrições e Regras

- **Pacotes Permitidos**: Use **apenas** o pacote `testing` - **não use** o pacote `reflect`
- **Escopo**: Use somente as structs e funções que estão no arquivo analisado
- **Sem Imports Desnecessários**: Não importe outros pacotes além do necessário
- **Sem Reescrever Código**: Não reescreva as structs e funções, apenas as chame
- **Apenas Elementos do Arquivo**: Não gere testes para funções que não estão no arquivo analisado

## Estrutura de Testes

### Organização de Casos de Teste

Cada método ou função a ser testada deve ter **um único método de teste** agrupando todos os casos usando `t.Parallel()`:

- Use `t.Run()` para cada caso de teste individual
- Cada caso deve usar `t.Parallel()` para execução paralela
- Use `t.Cleanup()` para limpeza de recursos quando necessário

### Exemplo de Estrutura

Use este exemplo como referência para a estrutura de testes:
```go
func TestPointToPoint_Send(t *testing.T) {
	t.Run("should send message successfully", func(t *testing.T) {
		t.Parallel()
		msg := &message.Message{}
		ctx := context.Background()
		ch := channel.NewPointToPointChannel("chan1")
		go ch.Send(ctx, msg)
		ch.Receive()
		t.Cleanup(func() {
			ch.Close()
		})
	})
	t.Run("should error when send message with context cancel", func(t *testing.T) {
		t.Parallel()
		ch := channel.NewPointToPointChannel("chan1")
		msg := &message.Message{}
		ctx, cancel := context.WithCancel(context.Background())
		cancel()
		err := ch.Send(ctx, msg)
		if err.Error() != "context cancelled while sending message: context canceled" {
			t.Errorf("Send should return nil error, got: %v", err)
		}
		t.Cleanup(func() {
			ch.Close()
		})
	})

	t.Run("shoud channel has been closed", func(t *testing.T) {
		t.Parallel()
		ch := channel.NewPointToPointChannel("chan1")
		msg := &message.Message{}
		ctx := context.Background()
		ch.Close()
		err := ch.Send(ctx, msg)
		if err.Error() != "channel has not been opened" {
			t.Error("Send should return error if channel is closed")
		}
	})
}
```
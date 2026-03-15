---
name: make_unit_tests
description: Generate comprehensive unit tests in Go with 100% coverage following specific structural standards.
---

# Go Unit Test Generation Skill

## 🎯 Objective

Analyze a given Go file and generate a complete unit test file following Go standard practices, ensuring maximum resilience and aiming for 100% test coverage.

## 📂 File Structure Guidelines

- **One test file per source file**: Generate a single `*_test.go` file for each analyzed source file.
- **Naming Convention**: The file name must match the analyzed file with the suffix `_test.go` (e.g., `service.go` -> `service_test.go`).
- **Namespace/Package**: Use the same package name as the analyzed file appended with `_test` (e.g., if package is `user`, use `package user_test`) to ensure black-box testing.

## 🎯 Coverage & Validation Specifications

- **Target Coverage**: Ensure your test scenarios cover **100% of the lines of code** present in the target module.
- **Thorough Validation**: Validate all input parameters and output values for each method and function.
- **Path Verification**: Include test cases that validate both **Success** and **Failure** scenarios.

## 🛑 Constraints & Rules

- **Allowed Packages**: Use **ONLY** the standard library `testing` package. Do **NOT** import the `reflect` package or external mocking libraries unless configured.
- **Scope Limitation**: Utilize only the structs and functions that are present natively in the analyzed file. Do not generate extraneous tests.
- **Clean Imports**: Keep imports strictly to what is necessary for the tests.
- **Code Preservation**: Do not rewrite existing structs or functions. Your sole objective is executing and testing them.

## 🏗️ Test Structure Implementation

Each tested method or function must have **one primary test method** that groups all cases securely.

### Test Organization

1. Use `t.Run()` for defining individual sub-test cases.
2. Every sub-test case must begin with `t.Parallel()` to enable parallel execution.
3. Use `t.Cleanup()` appropriately for resource sanitization, teardown, or channel closing.

### Example Reference Structure

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

	t.Run("should return error if channel has been closed", func(t *testing.T) {
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

---
description: Formatar código Go e gerar documentação GoDoc seguindo boas práticas
globs:
alwaysApply: false
---

# Go Code Formatting and GoDoc Documentation

Analyze the provided Go file and adjust its formatting and documentation according to Go best practices.

## Formatting Rules

Apply the following formatting rules:

- **Go Proverbs**: Adhere to Go's principles and proverbs.
- **Component Order**: Organize code components as per Go best practices:
  1. Package declaration
  2. Imports
  3. Constants
  4. Variables
  5. Types
  6. Functions
- **Line Length**: Each line of code must not exceed **100 characters**.
- **`gofmt` Style**: Ensure the code formatting matches `gofmt` output.

## GoDoc Documentation

Add comprehensive documentation following the GoDoc standard:

### File-Level Documentation

- At the beginning of the file, add a package-level comment specifying:
  - **Intent**: The overall purpose of the file.
  - **Objective**: What the implemented functionality achieves.

### Function and Method Documentation

- For each public function or method, add documentation specifying:
  - **Intent**: What the method/function does.
  - **Parameters**: Description of each input parameter (if any).
  - **Return Type**: Description of the value(s) returned (if any).
  - **Behavior**: Any special behaviors or side effects (if any).

## Guidelines

- **Language**: All documentation must be written in **English**.
- **GoDoc Standard**: Follow official GoDoc conventions.
- **Clarity**: Be clear and concise in the documentation.
- **Completeness**: Document all public elements (types, functions, methods, constants).
- **No API Changes**: Do not change public API names or signatures unless explicitly instructed.
- **Minimal Refactoring**: Preserve existing code style where possible and avoid unrelated refactoring.
- **Output**: If changes are made, produce an `apply_patch` style output and list modified files. For each modified file, include a brief summary of changes and the GoDoc added.
- **Static Checks**: Recommend running `go vet` and `go test` (if present) after modifications and report any failures.
- **Edge Cases**:
  - If existing GoDoc is present, improve its clarity and ensure it is in English.
  - If a function is too large, add a concise GoDoc and _suggest_ an internal refactor, but do not perform major refactors without confirmation.
  - If a line exceeds 100 characters, wrap or refactor it while preserving semantics.


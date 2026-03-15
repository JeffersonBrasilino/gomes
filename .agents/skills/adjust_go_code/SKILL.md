---
name: adjust_go_code
description: Format Go code and generate GoDoc documentation adhering to Go best practices.
---

# Go Code Formatting and GoDoc Structure Skill

## 🎯 Objective

Analyze Go files, format them cleanly, and generate comprehensive documentation according to Go best practices and standardized patterns.

## 🛠️ Formatting Rules

1. **Go Proverbs**: Ensure code logic adheres to Go's principles and proverbs.
2. **Component Organization**: Structure code components strictly as follows:
   1. Package declaration
   2. Imports
   3. Constants
   4. Variables
   5. Types
   6. Functions
3. **Line Constraints**: Maintain a maximum line length of **100 characters**. Wrap or lightly refactor lines that exceed this limit appropriately, while perfectly preserving semantics.
4. **Style Consistency**: Code must match standard `gofmt` output styling.

## 📚 GoDoc Documentation Guidelines

Implement comprehensive documentation complying with the official GoDoc standard. All documentation must be written in **English**.

### File-Level Documentation

- Insert a package-level comment at the very beginning of the file.
- **Intent**: Describe the overarching purpose of the file.
- **Objective**: Describe functionality achievements and behaviors mapped in this file.

### Function and Method Documentation

Document every public function, method, struct, and exported constant. Include:

- **Intent**: A detailed description of what the method or function does.
- **Parameters**: A detailed breakdown of each input parameter.
- **Return Type**: Description of the returned value(s), including errors.
- **Behavior**: Any side effects, special states, or specific panics the function might cause.

## 🛑 Constraints & Edge Cases

- **No API Changes**: Do NOT change public API names or signatures unless explicitly instructed by the user.
- **Minimal Refactoring**: Preserve the existing code style and avoid unrelated structural refactoring.
- **Existing Documentation**: If existing GoDoc is present, improve its clarity, correctly format it, and ensure it is in English.
- **Large Functions**: If a function is exceptionally large, add concise GoDoc and _suggest_ an internal refactor in your response, but do not perform major refactors automatically without confirmation.

## 📋 Output Expectations

- Apply requested changes using file editing tools directly.
- Briefly summarize the updates and GoDoc changes applied.
- Recommend running `go vet` and `go test` (if applicable) as a post-verification step to ensure code integrity.

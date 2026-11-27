# AGENTS.md

## Project

Sitka is a no-nonsense, simple, and joyful programming language. The repository contains an early Go tree-walking interpreter with a lexer, parser, AST, REPL, command-line expression evaluation, and source-file execution. Its syntax and built-ins are documented in `docs/language.md`.

Do not silently invent language syntax or semantics. Preserve the README's concise, welcoming tone.

## Repository setup

- Inspect existing files and manifests before choosing tools or dependencies.
- Make small, focused changes and avoid framework or dependency-heavy scaffolding.
- Keep the executable in `cmd/sitka/` and language packages under `internal/`.
- Keep tests alongside their Go packages; put runnable programs in `examples/` and language/contributor guidance in `docs/`.
- Preserve the separation between AST, lexer, parser, interpreter, and command-line responsibilities.

## Implementation conventions

- Favor clarity, straightforward behavior, descriptive names, and consistent formatting.
- Keep interfaces small and avoid circular dependencies.
- Design useful diagnostics and a pleasant first-use experience.
- Add focused automated tests for language behavior and boundary cases.
- Follow `docs/language.md`; update it, tests, and examples when syntax or behavior changes.
- Keep user-facing errors specific and cap built-in operations that can allocate large outputs.

## Validation

- Run `gofmt`, `go test ./...`, and `go vet ./...` for Go changes.
- Smoke-test affected examples and CLI behavior; keep documented commands verified and runnable.
- Ignore generated artifacts and local configuration rather than committing them.

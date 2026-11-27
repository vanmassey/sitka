# Sitka project instructions

## Project state

- Sitka is a no-nonsense, simple, and joyful programming language.
- The repository contains an early Go tree-walking interpreter. The executable is in `cmd/sitka/`; AST, lexer, parser, interpreter, and tests are under `internal/`.
- `docs/language.md` is the current language reference. Keep it, the examples, tests, and implementation consistent.
- Preserve the project's concise, welcoming tone. Do not silently add undocumented language syntax or semantics.

## Working with the repository

- Inspect existing files and manifests before choosing tools or adding dependencies.
- Make focused changes and avoid introducing a framework or large dependency stack without need.
- Keep the executable in `cmd/sitka/` and language packages under `internal/`.
- Keep tests alongside their Go packages; put runnable programs in `examples/` and language/contributor guidance in `docs/`.
- Preserve the separation between AST, lexer, parser, interpreter, and CLI concerns. Keep interfaces small and avoid circular dependencies.
- When syntax or behavior changes, update the language guide, focused tests, and affected examples.

## Implementation principles

- Favor clarity and straightforward behavior over cleverness.
- Optimize for a pleasant first-use experience and specific, useful error messages.
- Use descriptive names, consistent formatting, and Go idioms.
- Add focused tests for language behavior, boundary cases, and runtime errors.
- Bound built-in operations that can allocate large outputs.

## Documentation and validation

- Keep `README.md` concise, accurate, and runnable.
- Run `gofmt`, `go test ./...`, and `go vet ./...` for Go changes.
- Smoke-test affected examples and CLI behavior before considering work complete.
- Ignore generated artifacts and local configuration rather than committing them.

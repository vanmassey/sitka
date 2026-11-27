# sitka

A no-nonsense, simple, and joyful programming language.

## Current status

This repository contains an early Go implementation of the Sitka programming
language. It supports variables, integers, floats, strings, booleans, lists,
maps, functions, closures, conditionals, loops, and a growing standard library.

See [the language guide](docs/language.md) and the programs in [examples/](examples/).

## Build and run

```sh
go run ./cmd/sitka
```

This starts the interactive REPL:

```text
sitka> 2 + 3 * 4
14
```

You can evaluate a single expression or run a Sitka source file:

```sh
go run ./cmd/sitka "12 / (2 + 2)"
3
go run ./cmd/sitka examples/fibonacci.sitka
```

## Test

```sh
go test ./...
```

## License

Licensed under either of [Apache License, Version 2.0](LICENSE-APACHE) or
[MIT License](LICENSE-MIT), at your option.

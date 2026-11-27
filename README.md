# sitka

A no-nonsense, simple, and joyful programming language.

## Current status

This repository contains an early Go implementation of the Sitka programming
language. It supports variables, integers, floats, strings, booleans, lists,
maps, structs, functions, closures, modules, conditionals, loops, error
handling, static checks, and a growing standard library.

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

Check a source file or every `.sitka` file under a directory without running it:

```sh
go run ./cmd/sitka check examples/data-pipeline.sitka
go run ./cmd/sitka check examples
```

The standard library includes collection transforms, CSV and JSON conversion,
regular expressions, text and byte file operations, path helpers, date/time
utilities, and string encodings. See [the language guide](docs/language.md)
for the complete reference.

## Test

```sh
go test ./...
```

## License

Licensed under either of [Apache License, Version 2.0](LICENSE-APACHE) or
[MIT License](LICENSE-MIT), at your option.

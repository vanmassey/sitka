# Sitka language guide

Sitka is a small, dynamically typed language focused on readable programs and
friendly errors.

## Values

```sitka
let count = 42
let ratio = 3.14
let name = "Ada"
let enabled = true
let empty = nil
let numbers = [1, 2, 3]
let person = {"name": "Ada", "language": "Sitka"}
```

Numbers may be integers or floats. Strings use double or single quotes.
Comments begin with `//`; block comments use `/* ... */`.

## Variables

Use `let` to declare a variable. Assignment uses `=`:

```sitka
let score = 10
score = score + 5
```

Variables can be reassigned, but cannot be declared in one scope and then
implicitly created in another.

## Operators

Arithmetic:

- `+`, `-`, `*`, `/`, `%`

Comparisons:

- `==`, `!=`, `<`, `<=`, `>`, `>=`

Logical:

- `&&`, `||`, `!`

Strings can be concatenated with `+`, repeated with `*`, and compared
lexically. Lists can be concatenated with `+`.

## Printing

`print` accepts a comma-separated list of values:

```sitka
print "The answer is", 42
```

## Lists and maps

Negative list indexes count from the end:

```sitka
let numbers = [10, 20, 30]
print numbers[0]
print numbers[-1]
numbers[1] = 25
```

Maps use string keys:

```sitka
let colors = {"sky": "blue", "grass": "green"}
print colors["sky"]
print colors.sky
```

## Functions

Functions use `fun`, curly-braced blocks, and `return`:

```sitka
fun square(value) {
    return value * value
}

print square(12)
```

Functions close over the environment in which they were declared.
Anonymous functions use the same syntax without a name:

```sitka
let squares = map([1, 2, 3], fun(value) {
    return value * value
})
```

## Conditionals

```sitka
if score >= 10 {
    print "Winner"
} else if score >= 5 {
    print "Almost"
} else {
    print "Try again"
}
```

## Loops

Use `while` for a condition-based loop:

```sitka
let count = 3
while count > 0 {
    print count
    count = count - 1
}
```

Use `for ... in` to iterate over a list:

```sitka
for value in [10, 20, 30] {
    print value
}
```

`range(end)`, `range(start, end)`, and `range(start, end, step)` produce lists.
Use `break` and `continue` to control loop execution.

## Built-in functions

- `len(value)` - string, list, or map length
- `str(value)` - convert a value to display text
- `int(value)` - convert a value to an integer
- `float(value)` - convert a value to a float
- `type(value)` - return the value's type name
- `range(...)` - produce an integer list
- `push(list, value)` - return a list with a value appended
- `pop(list)` - return a list without its final value
- `abs(number)` - absolute value
- `min(...)` - lowest integer argument
- `max(...)` - highest integer argument
- `contains(container, value)` - test membership
- `clock()` - seconds since the Unix epoch
- `map(list, function)` - transform each item; callbacks receive one item
- `filter(list, function)` - keep items whose callback result is truthy
- `reduce(list, function, initial)` - fold values with a two-argument callback
- `sorted(list)` - return an ordered copy of a list
- `reversed(list_or_string)` - return values in reverse order
- `enumerate(list, start?)` - pair each item with its index
- `zip(list, list, ...)` - pair items up to the shortest list
- `keys(map)` / `values(map)` - return map contents in key order
- `join(list, separator)` / `split(string, separator)` - convert between text and lists
- `upper(string)`, `lower(string)`, `trim(string)` - transform text
- `replace(string, old, new)`, `starts_with(string, prefix)`, `ends_with(string, suffix)`
- `sum(list)`, `all(list)`, `any(list)` - aggregate list values
- `first(list)`, `last(list)`, `slice(list, start, end?)` - retrieve list sections
- `unique(list)`, `flatten(list)`, `chars(string)` - reshape collections and text
- `assert(condition, message?)` - fail with a useful message when a condition is false

Collection transforms return new lists and leave their input lists unchanged.
`range` and string repetition have output limits to avoid accidental enormous
allocations.

## Command line

Start the REPL:

```sh
go run ./cmd/sitka
```

The REPL keeps variables and functions between lines. Open blocks can span
multiple lines; the prompt changes to `...>` until the block is complete.

Evaluate an expression:

```sh
go run ./cmd/sitka "1 + 2 * 3"
```

Run a source file:

```sh
go run ./cmd/sitka examples/fibonacci.sitka
```

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

Named structs describe fixed sets of properties:

```sitka
struct Person {
    name,
    age
}

let person = Person("Ada", 37)
person.age = 38
print person.name
```

Constructors take one positional value per declared field. Struct fields can be
read and updated using dot notation or string-key indexes. Struct equality is
structural for values of the same struct type.

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

## Error handling

Use `try` and `catch` to handle runtime errors. The catch variable contains the
error message as a string. `throw` raises a language error:

```sitka
try {
    throw "the operation could not continue"
} catch problem {
    print "Handled:", problem
}
```

The try and catch blocks have their own scopes. Return, break, and continue
control flow passes through a try block when no error occurs.

## Modules

Import a Sitka file with a quoted path and an explicit alias:

```sitka
import "math.sitka" as math
print math.double(21)
```

Paths are resolved relative to the importing file. Every top-level name not
starting with `_` is exported as a property of the module alias. Imported
modules are cached for the duration of a program; circular imports are
reported as errors.

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
- `count(container, value)`, `index_of(list, value)` - search values
- `take(list, count)`, `drop(list, count)`, `chunk(list, size)` - divide lists
- `partition(list, function)`, `find(list, function)` - select by callback
- `flat_map(list, function)`, `group_by(list, function)`, `count_by(list, function)`
- `sort_by(list, function)`, `min_by(list, function)`, `max_by(list, function)`
- `take_while(list, function)`, `drop_while(list, function)`, `find_index(list, function)`
- `key_by(list, function)` - index values using unique string callback results
- `assert(condition, message?)` - fail with a useful message when a condition is false
- `pi`, `e` - mathematical constants
- `sqrt`, `pow`, `floor`, `ceil`, `round`, `clamp`, `sign` - numeric helpers
- `sin`, `cos`, `tan`, `asin`, `acos`, `atan`, `atan2`, `radians`, `degrees`, `is_finite`
- `exp`, `ln`, `log2`, `log10`, `hypot`, `trunc`, `gcd`, `lcm`, `divmod`, `factorial`
- `json_encode(value)`, `json_decode(string)` - convert JSON values
- `read_text(path)`, `write_text(path, string)`, `append_text(path, string)` - text files
- `file_exists(path)`, `make_dir(path)`, `list_dir(path)` - basic filesystem operations
- `read_lines(path)`, `write_lines(path, lines)` - UTF-8 line-oriented file I/O
- `read_bytes(path)`, `write_bytes(path, bytes)`, `append_bytes(path, bytes)` - byte lists
- `csv_decode(text)`, `csv_encode(rows)` - convert CSV to and from lists of rows
- `csv_decode_maps(text)`, `csv_encode_maps(rows, headers?)` - map rows by header
- `regex_match(text, pattern)`, `regex_find(text, pattern)`, `regex_find_all(text, pattern)`
- `regex_replace(text, pattern, replacement)`, `regex_split(text, pattern)`, `regex_captures(text, pattern)`
- `path_join(...)`, `path_clean(path)`, `path_base(path)`, `path_dir(path)`, `path_ext(path)`
- `path_is_abs(path)`, `path_abs(path)`, `path_rel(base, target)`, `path_separator`, `cwd()`
- `time_unix(year, month, day, hour?, minute?, second?)`, `time_parse(text, layout)`
- `time_format(timestamp, layout)`, `time_utc(timestamp)`, `time_add(timestamp, seconds)`
- `time_diff(start, end)`, `time_parts(timestamp)` - work with Unix-second timestamps
- `substring(text, start, end?)`, `rune_at(text, index)`, `ord(char)`, `chr(codepoint)`
- `repeat(text, count)`, `pad_left(text, width, char)`, `pad_right(text, width, char)`
- `split_lines(text)`, `join_lines(lines)` - newline-normalized text helpers
- `base64_encode/decode`, `hex_encode/decode`, `url_encode/decode`
- `query_encode(map)`, `query_decode(string)` - URL query maps with repeated values
- `url_parse(string)`, `url_build(map)` - inspect or construct a URL
- `sha256(string)`, `sha256_bytes(bytes)` - SHA-256 digests as lowercase hex

Collection transforms return new lists and leave their input lists unchanged.
`range` and string repetition have output limits to avoid accidental enormous
allocations.
Filesystem paths are relative to the currently running source file (or the
current working directory in the REPL). File and directory errors are reported
to the program; these built-ins do not silently ignore failures.
`write_lines` writes a newline after each supplied line. `read_bytes` and
`write_bytes` use integer lists with values from 0 through 255. Text encoding
helpers reject decoded byte sequences that are not valid UTF-8.
CSV conversion follows standard quoted-field rules; map conversion uses the
first CSV row as headers and sorts inferred headers when encoding. Regular
expressions use Go's RE2-compatible syntax. Time formatting uses Go's reference
date layout (for example, `2006-01-02`); timestamps are seconds since the Unix
epoch, and calendar operations use UTC unless the parsed layout includes a zone.

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

Pass arguments to a source file after its path. Programs can read them from the
global `args` list:

```sh
go run ./cmd/sitka examples/hello.sitka Ada
```

`args` contains strings and is empty in the REPL and when evaluating an
expression directly. See `examples/arguments.sitka` for a runnable example.

Run the static checker without executing the program:

```sh
go run ./cmd/sitka check examples/data-pipeline.sitka
go run ./cmd/sitka check examples
```

The checker reports undefined names, duplicate declarations, statically known
function arity mismatches, and invalid `return`, `break`, or `continue`
statements. It is an early error checker, not a complete type checker.

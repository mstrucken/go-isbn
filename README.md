# go-isbn

[![Go Reference](https://pkg.go.dev/badge/github.com/mstrucken/go-isbn.svg)](https://pkg.go.dev/github.com/mstrucken/go-isbn)
[![Go Report Card](https://goreportcard.com/badge/github.com/mstrucken/go-isbn)](https://goreportcard.com/report/github.com/mstrucken/go-isbn)
[![License: MIT](https://img.shields.io/badge/License-MIT-yellow.svg)](https://opensource.org/licenses/MIT)

A Go library for parsing, validating, converting, and formatting ISBN-10 and ISBN-13 identifiers.

## Installation

```bash
go get github.com/mstrucken/go-isbn
```

Requires Go 1.22 or later.

## Usage

### Parsing

`Parse` validates the input and returns a fully populated `ISBN` struct. It accepts ISBN-10 or ISBN-13, with or without hyphens.

```go
b, err := isbn.Parse("978-0-306-40615-7")
if err != nil {
    log.Fatal(err)
}

fmt.Println(b.Digits)                    // 9780306406157
fmt.Println(b.Prefix)                    // 978
fmt.Println(b.RegistrationGroup)         // 0
fmt.Println(b.RegistrationGroupAgency)   // English language
fmt.Println(b.Registrant)                // 306
fmt.Println(b.Publication)               // 40615
fmt.Println(b.CheckDigit)                // 7
fmt.Println(b.Version)                   // 13
```

`Raw` preserves the original input string exactly as provided, while `Digits` holds the normalised digit-only form:

```go
b, _ := isbn.Parse("978-0-306-40615-7")
fmt.Println(b.Raw)    // 978-0-306-40615-7
fmt.Println(b.Digits) // 9780306406157
```

Use `ParseAsISBN13` when you always want an ISBN-13 result regardless of input form:

```go
b, err := isbn.ParseAsISBN13("0306406152")
fmt.Println(b.Digits) // 9780306406157
```

### Validation

```go
err := isbn.Validate("9780306406157") // nil
err := isbn.Validate("12345")         // ErrInvalidLength

ok := isbn.IsValid("978-0-306-40615-7") // true
```

### Conversion

Convert between versions using the methods on `ISBN`:

```go
b, _ := isbn.Parse("9780306406157")

b10, err := b.ToISBN10()
fmt.Println(b10.Digits) // 0306406152

b13 := b10.ToISBN13()
fmt.Println(b13.Digits) // 9780306406157
```

Or use the string convenience functions:

```go
s, err := isbn.ConvertToISBN13("0306406152")
fmt.Println(s) // 9780306406157

s, err := isbn.ConvertToISBN10("9780306406157")
fmt.Println(s) // 0306406152
```

Note: only ISBN-13 values with the `978` prefix have an ISBN-10 equivalent. Attempting to convert a `979` prefix returns `ErrNoISBN10Equivalent`.

### Formatting

`Hyphenate` inserts hyphens at the correct positions according to the official ISBN range data. The `ISBN` struct implements `fmt.Stringer`, so hyphenated output is also the default in `fmt.Sprintf` and log output.

```go
b, _ := isbn.Parse("9780306406157")
fmt.Println(b.Hyphenate()) // 978-0-306-40615-7
fmt.Println(b.String())    // 978-0-306-40615-7
fmt.Printf("%v\n", b)      // 978-0-306-40615-7

// Use Digits directly when you need the bare digit sequence
fmt.Println(b.Digits) // 9780306406157
```

Or use the string convenience function:

```go
s, err := isbn.HyphenateString("9780306406157")
fmt.Println(s) // 978-0-306-40615-7
```

When the registration group or registrant cannot be determined from the range data, the library falls back to minimal hyphenation rather than returning an error.

## Error Handling

All functions return typed errors that can be matched with `errors.Is`:

```go
_, err := isbn.Parse("9780306406158")
if errors.Is(err, isbn.ErrInvalidCheckDigit) {
    // handle bad check digit
}
```

Available sentinel errors:

| Sentinel                   | Meaning                                          |
|----------------------------|--------------------------------------------------|
| `ErrInvalidLength`         | Input is not 10 or 13 digits after cleaning      |
| `ErrInvalidCharacter`      | Non-digit character found in an invalid position |
| `ErrInvalidCheckDigit`     | Check digit does not match the computed value    |
| `ErrNoISBN10Equivalent`    | ISBN-13 with `979` prefix has no ISBN-10 form    |

Use `errors.As` to retrieve the full error message:

```go
var isbnErr *isbn.ISBNError
if errors.As(err, &isbnErr) {
    fmt.Println(isbnErr.Error())
}
```

## The ISBN Struct

| Field                      | Description                                                             |
|----------------------------|-------------------------------------------------------------------------|
| `Raw`                      | Original input, preserved exactly as provided                           |
| `Digits`                   | Normalised digit-only form; use this for programmatic access            |
| `Prefix`                   | GS1 prefix (`"978"` or `"979"`); empty for ISBN-10                     |
| `RegistrationGroup`        | Language/country group element (e.g. `"0"`, `"3"`, `"99937"`)         |
| `RegistrationGroupAgency`  | Human-readable group name (e.g. `"English language"`)                  |
| `Registrant`               | Publisher element                                                       |
| `Publication`              | Title-specific element                                                  |
| `CheckDigit`               | Final character (`0`–`9` or `X` for ISBN-10)                           |
| `Version`                  | `V10` or `V13`                                                          |

## Range Data

Hyphenation and group/registrant splitting rely on the official ISBN range message published by the International ISBN Agency. To regenerate the embedded range data from an updated `RangeMessage.xml`:

```bash
go generate ./...
```

## License

MIT

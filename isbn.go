// Package isbn provides functions for parsing, validating, converting,
// and formatting ISBN-10 and ISBN-13 identifiers.
//
//go:generate go run ./cmd/gen/main.go -input RangeMessage.xml -output ranges.go
package isbn

import (
	"strconv"
	"strings"
)

// Version identifies whether an ISBN was parsed from a 10- or 13-digit form.
type Version uint8

const (
	// V10 indicates an ISBN-10.
	V10 Version = 10
	// V13 indicates an ISBN-13.
	V13 Version = 13
)

type isbnErrCode int

const (
	errCodeInvalidLength isbnErrCode = iota + 1
	errCodeInvalidCharacter
	errCodeInvalidCheckDigit
	errCodeNoISBN10Equivalent
)

// ISBNError is the error type returned by all isbn functions.
// Use errors.Is against the sentinel variables (ErrInvalidLength, etc.) to
// distinguish error categories without string matching.
type ISBNError struct {
	code    isbnErrCode
	message string
}

// Error implements the error interface.
func (e *ISBNError) Error() string { return e.message }

// Is reports whether e belongs to the same error category as target,
// enabling errors.Is comparisons against the sentinel variables.
func (e *ISBNError) Is(target error) bool {
	t, ok := target.(*ISBNError)
	if !ok {
		return false
	}
	return e.code == t.code
}

// Sentinel errors for use with errors.Is.
var (
	ErrInvalidLength      = &ISBNError{code: errCodeInvalidLength}
	ErrInvalidCharacter   = &ISBNError{code: errCodeInvalidCharacter}
	ErrInvalidCheckDigit  = &ISBNError{code: errCodeInvalidCheckDigit}
	ErrNoISBN10Equivalent = &ISBNError{code: errCodeNoISBN10Equivalent}
)

// ISBN holds the fully parsed components of an ISBN-10 or ISBN-13.
//
// For ISBN-10, Prefix is empty and all digit positions are relative to the
// 10-digit form. For ISBN-13, Prefix is "978" or "979".
type ISBN struct {
	// Raw is the original input string, preserved exactly as provided
	// (hyphens and spaces included). For ISBNs produced by conversion
	// (ToISBN10, ToISBN13) rather than direct parsing, Raw equals Digits.
	Raw string

	// Digits is the normalised form: digits only, no hyphens or spaces.
	// May end in 'X' for ISBN-10. Use this field for all programmatic access.
	Digits string

	// Prefix is the GS1 prefix ("978" or "979") for ISBN-13.
	// Empty for ISBN-10.
	Prefix string

	// RegistrationGroup is the language/country group element (e.g. "0", "3", "99937").
	// Empty string when the range is undefined/reserved.
	RegistrationGroup string

	// RegistrationGroupAgency is the human-readable name of the group
	// (e.g. "English language"). Empty when group is undefined.
	RegistrationGroupAgency string

	// Registrant is the publisher element.
	// Empty string when the range is undefined/reserved.
	Registrant string

	// Publication is the title-specific element.
	Publication string

	// CheckDigit is the final character ('0'–'9', or 'X' for ISBN-10).
	CheckDigit string

	// Version is V10 or V13 depending on the form the ISBN was parsed from.
	Version Version
}

// Parse normalises the input (strips hyphens and spaces), determines whether it
// is ISBN-10 or ISBN-13, validates the check digit, and returns a fully
// populated ISBN struct. ISBN-10 inputs are parsed as-is and are not converted
// to ISBN-13.
//
// Returns an error when:
//   - the input length is not 10 or 13 (after cleaning)
//   - non-digit characters remain (except 'X' as the last char of ISBN-10)
//   - the check digit is invalid
func Parse(s string) (ISBN, error) {
	c := canonicalize(s)
	switch len(c) {
	case 10:
		if err := validateISBN10(c); err != nil {
			return ISBN{}, err
		}
		b := split10(c)
		b.Raw = s
		return b, nil
	case 13:
		if err := validateISBN13(c); err != nil {
			return ISBN{}, err
		}
		b := split13(c)
		b.Raw = s
		return b, nil
	default:
		return ISBN{}, &ISBNError{code: errCodeInvalidLength, message: "invalid length: must be 10 or 13 digits"}
	}
}

// ParseAsISBN13 normalises the input (strips hyphens and spaces), determines whether it
// is ISBN-10 or ISBN-13, validates the check digit, and returns a fully
// populated ISBN struct. In contrast to Parse, ISBN-10 inputs are converted
// to their ISBN-13 equivalent; in that case Raw and Digits will reflect the
// converted form rather than the original input.
func ParseAsISBN13(s string) (ISBN, error) {
	b, err := Parse(s)
	if err != nil {
		return ISBN{}, err
	}
	if b.Version == V10 {
		return b.ToISBN13(), nil
	}
	return b, nil
}

// Validate returns nil when s is a syntactically valid ISBN-10 or ISBN-13.
// It does NOT require hyphens in any particular position.
func Validate(s string) error {
	c := canonicalize(s)
	switch len(c) {
	case 10:
		return validateISBN10(c)
	case 13:
		return validateISBN13(c)
	default:
		return &ISBNError{code: errCodeInvalidLength, message: "invalid length: must be 10 or 13 digits"}
	}
}

// IsValid returns true when s is a syntactically valid ISBN-10 or ISBN-13.
// It does NOT require hyphens in any particular position.
func IsValid(s string) bool {
	return Validate(s) == nil
}

// ToISBN10 returns the ISBN-10 representation of b as a parsed ISBN struct.
// If b is already an ISBN-10 it is returned unchanged, meaning Raw will
// preserve the original input (e.g. "0-306-40615-2") rather than the bare
// digit sequence. For ISBNs that go through conversion, Raw and Digits will
// be identical.
// Returns an error when b is an ISBN-13 with a "979" prefix, which has no
// ISBN-10 equivalent.
func (b ISBN) ToISBN10() (ISBN, error) {
	if b.Version == V10 {
		return b, nil
	}
	if b.Prefix != "978" {
		return ISBN{}, &ISBNError{code: errCodeNoISBN10Equivalent, message: "cannot convert to ISBN-10: prefix is not \"978\""}
	}
	base9 := b.Digits[3:12]
	check := isbn10CheckDigit(base9)
	var digits string
	if check == 10 {
		digits = base9 + "X"
	} else {
		digits = base9 + strconv.Itoa(check)
	}
	return split10(digits), nil
}

// ToISBN13 returns the ISBN-13 representation of b as a parsed ISBN struct.
// If b is already an ISBN-13 it is returned unchanged, meaning Raw will
// preserve the original input (e.g. "978-0-306-40615-7") rather than the bare
// digit sequence. For ISBNs that go through conversion from ISBN-10, Raw and
// Digits will be identical. ISBN-10 always maps to the "978" prefix space.
func (b ISBN) ToISBN13() ISBN {
	if b.Version == V13 {
		return b
	}
	return split13(convertTo13digits(b.Digits))
}

// ConvertToISBN10 converts an ISBN string to its ISBN-10 equivalent.
// Input may be ISBN-10 or ISBN-13 (with or without hyphens).
// Only the "978" prefix can be represented as ISBN-10.
// Returns an error if the input is invalid or has a "979" prefix.
func ConvertToISBN10(s string) (string, error) {
	b, err := Parse(s)
	if err != nil {
		return "", err
	}
	b10, err := b.ToISBN10()
	if err != nil {
		return "", err
	}
	return b10.Digits, nil
}

// ConvertToISBN13 converts an ISBN string to its ISBN-13 equivalent.
// Input may be ISBN-10 or ISBN-13 (with or without hyphens).
// The only possible error is a parse failure on the input string itself;
// once parsed, conversion to ISBN-13 is always successful.
func ConvertToISBN13(s string) (string, error) {
	b, err := Parse(s)
	if err != nil {
		return "", err
	}
	// ToISBN13 is infallible: ISBN-13 is returned unchanged, and
	// ISBN-10 always has a "978" equivalent, so no error check is needed.
	return b.ToISBN13().Digits, nil
}

// HyphenateString returns the ISBN with hyphens inserted according to the range data.
// ISBN-13 is formatted as e.g. "978-0-306-40615-7".
// ISBN-10 is formatted as e.g. "0-306-40615-2".
// Falls back to minimal hyphenation when the registration group or registrant
// cannot be determined. Input may be ISBN-10 or ISBN-13 (with or without hyphens).
func HyphenateString(s string) (string, error) {
	b, err := Parse(s)
	if err != nil {
		return "", err
	}
	return b.Hyphenate(), nil
}

// IsISBN10 reports whether this ISBN was parsed from a 10-digit input.
func (b ISBN) IsISBN10() bool { return b.Version == V10 }

// IsISBN13 reports whether this ISBN was parsed from a 13-digit input.
func (b ISBN) IsISBN13() bool { return b.Version == V13 }

// String implements fmt.Stringer and returns the hyphenated ISBN, making it
// the default representation in fmt.Sprintf, log output, and similar contexts.
// Note that this means %v and %s produce a hyphenated string, not the bare
// digit sequence. Use the Digits field directly when you need the unhyphenated
// form for programmatic access.
func (b ISBN) String() string {
	return b.Hyphenate()
}

// Hyphenate returns the ISBN with hyphens inserted according to the range data.
// ISBN-13 is formatted as e.g. "978-0-306-40615-7".
// ISBN-10 is formatted as e.g. "0-306-40615-2".
func (b ISBN) Hyphenate() string {
	if b.Prefix == "" {
		// ISBN-10
		if b.RegistrationGroup == "" {
			return b.Digits[:9] + "-" + b.CheckDigit
		}
		if b.Registrant == "" {
			return b.RegistrationGroup + "-" + b.Digits[len(b.RegistrationGroup):9] + "-" + b.CheckDigit
		}
		return b.RegistrationGroup + "-" + b.Registrant + "-" + b.Publication + "-" + b.CheckDigit
	}
	// ISBN-13
	if b.RegistrationGroup == "" {
		return b.Prefix + "-" + b.Digits[3:12] + "-" + b.CheckDigit
	}
	if b.Registrant == "" {
		offset := 3 + len(b.RegistrationGroup)
		return b.Prefix + "-" + b.RegistrationGroup + "-" + b.Digits[offset:12] + "-" + b.CheckDigit
	}
	return b.Prefix + "-" + b.RegistrationGroup + "-" + b.Registrant + "-" + b.Publication + "-" + b.CheckDigit
}

// cleaner is a package-level replacer to avoid reallocating on every call to canonicalize.
var cleaner = strings.NewReplacer("-", "", " ", "")

// canonicalize strips hyphens and spaces from s and uppercases the result,
// normalising lowercase 'x' check digits to 'X'.
func canonicalize(s string) string {
	s = strings.ToUpper(s)
	return cleaner.Replace(s)
}

// convertTo13digits converts a clean 10-digit ISBN string to a 13-digit string.
func convertTo13digits(isbn10 string) string {
	base := "978" + isbn10[:9]
	check := isbn13CheckDigit(base)
	return base + strconv.Itoa(check)
}

// split10 parses a validated 10-digit ISBN-10 string into an ISBN struct.
// Both Raw and Digits are set to s; Parse overwrites Raw with the original input.
// Range lookups use the implicit "978" GS1 prefix.
func split10(s string) ISBN {
	b := ISBN{
		Raw:        s,
		Digits:     s,
		Prefix:     "",
		CheckDigit: s[9:10],
		Version:    V10,
	}

	// Use the first 7 digits for group lookup against the "978" prefix ranges.
	val, err := strconv.ParseUint(s[0:7], 10, 32)
	if err != nil {
		return b
	}
	v := uint32(val)

	groupLen := 0
	for _, pr := range gs1Prefixes["978"] {
		if v >= pr.lo && v <= pr.hi {
			groupLen = pr.length
			break
		}
	}
	if groupLen == 0 {
		return b
	}

	b.RegistrationGroup = s[0:groupLen]

	// Find registrant length from registrationGroups using the "978-<group>" key.
	groupKey := "978-" + b.RegistrationGroup
	registrantLen := 0
	var agency string
	for _, rg := range registrationGroups {
		if rg.prefix != groupKey {
			continue
		}
		agency = rg.agency
		end := min(groupLen+7, 9)
		afterGroup := s[groupLen:end]
		padded := afterGroup + strings.Repeat("0", max(0, 7-len(afterGroup)))
		rv, err := strconv.ParseUint(padded, 10, 32)
		if err != nil {
			break
		}
		rv32 := uint32(rv)
		for _, gr := range rg.ranges {
			if rv32 >= gr.lo && rv32 <= gr.hi {
				registrantLen = gr.length
				break
			}
		}
		break
	}

	b.RegistrationGroupAgency = agency

	if registrantLen == 0 {
		return b
	}

	registrantEnd := groupLen + registrantLen
	b.Registrant = s[groupLen:registrantEnd]
	b.Publication = s[registrantEnd:9]

	return b
}

// split13 parses a validated 13-digit ISBN-13 string into an ISBN struct.
// Both Raw and Digits are set to s; Parse overwrites Raw with the original input.
func split13(s string) ISBN {
	b := ISBN{
		Raw:        s,
		Digits:     s,
		Prefix:     s[:3],
		CheckDigit: s[12:13],
		Version:    V13,
	}

	sevenDigits := s[3:10]
	val, err := strconv.ParseUint(sevenDigits, 10, 32)
	if err != nil {
		return b
	}
	v := uint32(val)

	groupLen := 0
	for _, pr := range gs1Prefixes[b.Prefix] {
		if v >= pr.lo && v <= pr.hi {
			groupLen = pr.length
			break
		}
	}
	if groupLen == 0 {
		return b
	}

	b.RegistrationGroup = s[3 : 3+groupLen]

	groupKey := b.Prefix + "-" + b.RegistrationGroup
	registrantLen := 0
	var agency string
	for _, rg := range registrationGroups {
		if rg.prefix != groupKey {
			continue
		}
		agency = rg.agency
		afterGroup := s[3+groupLen : 10]
		padded := afterGroup + strings.Repeat("0", max(0, 7-len(afterGroup)))
		rv, err := strconv.ParseUint(padded, 10, 32)
		if err != nil {
			break
		}
		rv32 := uint32(rv)
		for _, gr := range rg.ranges {
			if rv32 >= gr.lo && rv32 <= gr.hi {
				registrantLen = gr.length
				break
			}
		}
		break
	}

	b.RegistrationGroupAgency = agency

	if registrantLen == 0 {
		return b
	}

	registrantStart := 3 + groupLen
	registrantEnd := registrantStart + registrantLen
	b.Registrant = s[registrantStart:registrantEnd]
	b.Publication = s[registrantEnd:12]

	return b
}

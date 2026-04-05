package isbn

import "strconv"

// isbn13CheckDigit computes the check digit for a 12-digit ISBN-13 base string.
// The input must be exactly 12 ASCII digit characters.
func isbn13CheckDigit(base12 string) int {
	sum := 0
	for i := range 12 {
		d := int(base12[i] - '0')
		if i%2 == 0 {
			sum += d
		} else {
			sum += d * 3
		}
	}
	rem := sum % 10
	if rem == 0 {
		return 0
	}
	return 10 - rem
}

// validateISBN13 checks that s is a valid ISBN-13 string.
// Precondition: s must be exactly 13 characters; callers are responsible
// for ensuring the correct length before calling this function.
func validateISBN13(s string) error {
	for _, c := range s {
		if c < '0' || c > '9' {
			return &ISBNError{code: errCodeInvalidCharacter, message: "invalid character: " + string(c)}
		}
	}
	expected := isbn13CheckDigit(s[:12])
	got := int(s[12] - '0')
	if expected != got {
		return &ISBNError{code: errCodeInvalidCheckDigit, message: "invalid check digit"}
	}
	return nil
}

// isbn10CheckDigit computes the check digit for a 9-digit ISBN-10 base string.
// The input must be exactly 9 ASCII digit characters
// Returns a value in [0, 10], where 10 represents 'X'.
func isbn10CheckDigit(base9 string) int {
	sum := 0
	for i := range 9 {
		sum += int(base9[i]-'0') * (10 - i)
	}
	rem := sum % 11
	if rem == 0 {
		return 0
	}
	return 11 - rem
}

// validateISBN10 checks that s is a valid ISBN-10 string.
// Precondition: s must be exactly 10 characters; callers are responsible
// for ensuring the correct length before calling this function.
func validateISBN10(s string) error {
	sum := 0
	for i := range 9 {
		c := s[i]
		if c < '0' || c > '9' {
			return &ISBNError{code: errCodeInvalidCharacter, message: "invalid character at position " + strconv.Itoa(i) + ": " + string(c)}
		}
		sum += int(c-'0') * (10 - i)
	}
	last := s[9]
	if last == 'X' {
		sum += 10
	} else if last >= '0' && last <= '9' {
		sum += int(last - '0')
	} else {
		return &ISBNError{code: errCodeInvalidCharacter, message: "invalid check digit character: " + string(last)}
	}
	if sum%11 != 0 {
		return &ISBNError{code: errCodeInvalidCheckDigit, message: "invalid check digit"}
	}
	return nil
}

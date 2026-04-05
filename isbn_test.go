package isbn_test

import (
	"errors"
	"testing"

	"github.com/mstrucken/go-isbn"
)

func TestValidate(t *testing.T) {
	tests := []struct {
		name    string
		input   string
		wantErr error // nil means no error expected
	}{
		{"valid ISBN-13", "9780306406157", nil},
		{"valid ISBN-13 with hyphens", "978-0-306-40615-7", nil},
		{"valid ISBN-10", "0-306-40615-2", nil},
		{"ISBN-10 with X check", "080442957X", nil},
		{"invalid check digit", "9780306406158", isbn.ErrInvalidCheckDigit},
		{"wrong length", "12345", isbn.ErrInvalidLength},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := isbn.Validate(tt.input)
			if tt.wantErr == nil {
				if err != nil {
					t.Errorf("Validate(%q) unexpected error: %v", tt.input, err)
				}
				return
			}
			if !errors.Is(err, tt.wantErr) {
				t.Errorf("Validate(%q) error = %v, want %v", tt.input, err, tt.wantErr)
			}
		})
	}
}

func TestConvertTo13(t *testing.T) {
	tests := []struct {
		name    string
		input   string
		want    string
		wantErr error
	}{
		{"standard conversion", "0306406152", "9780306406157", nil},
		{"already ISBN-13", "9780306406157", "9780306406157", nil},
		{"invalid check digit", "0306406153", "", isbn.ErrInvalidCheckDigit},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := isbn.ConvertToISBN13(tt.input)
			if tt.wantErr == nil {
				if err != nil {
					t.Errorf("ConvertToISBN13(%q) unexpected error: %v", tt.input, err)
					return
				}
			} else if !errors.Is(err, tt.wantErr) {
				t.Errorf("ConvertToISBN13(%q) error = %v, want %v", tt.input, err, tt.wantErr)
				return
			}
			if got != tt.want {
				t.Errorf("ConvertToISBN13(%q) = %q, want %q", tt.input, got, tt.want)
			}
		})
	}
}

func TestHyphenate(t *testing.T) {
	tests := []struct {
		name    string
		input   string
		want    string
		wantErr error
	}{
		{"ISBN-13 digits", "9780306406157", "978-0-306-40615-7", nil},
		{"ISBN-13 with hyphens", "978-0-306-40615-7", "978-0-306-40615-7", nil},
		{"ISBN-10 input", "0306406152", "0-306-40615-2", nil},
		{"ISBN-10 including x", "344249401x", "3-442-49401-X", nil},
		{"wrong length", "12345", "", isbn.ErrInvalidLength},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := isbn.HyphenateString(tt.input)
			if tt.wantErr == nil {
				if err != nil {
					t.Errorf("Hyphenate(%q) unexpected error: %v", tt.input, err)
					return
				}
			} else if !errors.Is(err, tt.wantErr) {
				t.Errorf("Hyphenate(%q) error = %v, want %v", tt.input, err, tt.wantErr)
				return
			}
			if got != tt.want {
				t.Errorf("Hyphenate(%q) = %q, want %q", tt.input, got, tt.want)
			}
		})
	}
}

func TestParse(t *testing.T) {
	t.Run("struct fields for 9780306406157", func(t *testing.T) {
		b, err := isbn.Parse("9780306406157")
		if err != nil {
			t.Fatalf("Parse error: %v", err)
		}
		if b.Digits != "9780306406157" {
			t.Errorf("Digits = %q, want %q", b.Digits, "9780306406157")
		}
		if b.Prefix != "978" {
			t.Errorf("Prefix = %q, want %q", b.Prefix, "978")
		}
		if b.RegistrationGroup != "0" {
			t.Errorf("RegistrationGroup = %q, want %q", b.RegistrationGroup, "0")
		}
		if b.RegistrationGroupAgency != "English language" {
			t.Errorf("RegistrationGroupAgency = %q, want %q", b.RegistrationGroupAgency, "English language")
		}
		if b.Registrant != "306" {
			t.Errorf("Registrant = %q, want %q", b.Registrant, "306")
		}
		if b.Publication != "40615" {
			t.Errorf("Publication = %q, want %q", b.Publication, "40615")
		}
		if b.CheckDigit != "7" {
			t.Errorf("CheckDigit = %q, want %q", b.CheckDigit, "7")
		}
	})

	t.Run("Raw preserves original input with hyphens", func(t *testing.T) {
		const original = "978-0-306-40615-7"
		b, err := isbn.Parse(original)
		if err != nil {
			t.Fatalf("Parse error: %v", err)
		}
		if b.Raw != original {
			t.Errorf("Raw = %q, want %q", b.Raw, original)
		}
		if b.Digits != "9780306406157" {
			t.Errorf("Digits = %q, want %q", b.Digits, "9780306406157")
		}
	})

	t.Run("ISBN-10 parsed as ISBN-10", func(t *testing.T) {
		b, err := isbn.Parse("0306406152")
		if err != nil {
			t.Fatalf("Parse error: %v", err)
		}
		if b.Digits != "0306406152" {
			t.Errorf("Digits = %q, want %q", b.Digits, "0306406152")
		}
		if b.Prefix != "" {
			t.Errorf("Prefix = %q, want empty for ISBN-10", b.Prefix)
		}
		if b.Version != isbn.V10 {
			t.Errorf("Version = %v, want V10", b.Version)
		}
		if b.RegistrationGroup != "0" {
			t.Errorf("RegistrationGroup = %q, want %q", b.RegistrationGroup, "0")
		}
		if b.RegistrationGroupAgency != "English language" {
			t.Errorf("RegistrationGroupAgency = %q, want %q", b.RegistrationGroupAgency, "English language")
		}
		if b.Registrant != "306" {
			t.Errorf("Registrant = %q, want %q", b.Registrant, "306")
		}
		if b.Publication != "40615" {
			t.Errorf("Publication = %q, want %q", b.Publication, "40615")
		}
		if b.CheckDigit != "2" {
			t.Errorf("CheckDigit = %q, want %q", b.CheckDigit, "2")
		}
	})

	t.Run("invalid check digit", func(t *testing.T) {
		_, err := isbn.Parse("9780306406158")
		if !errors.Is(err, isbn.ErrInvalidCheckDigit) {
			t.Errorf("Parse(%q) error = %v, want %v", "9780306406158", err, isbn.ErrInvalidCheckDigit)
		}
	})

	t.Run("invalid character", func(t *testing.T) {
		_, err := isbn.Parse("97X0306406157")
		if !errors.Is(err, isbn.ErrInvalidCharacter) {
			t.Errorf("Parse(%q) error = %v, want %v", "97X0306406157", err, isbn.ErrInvalidCharacter)
		}
	})

	t.Run("wrong length", func(t *testing.T) {
		_, err := isbn.Parse("12345")
		if !errors.Is(err, isbn.ErrInvalidLength) {
			t.Errorf("Parse(%q) error = %v, want %v", "12345", err, isbn.ErrInvalidLength)
		}
	})
}

func TestToISBN10(t *testing.T) {
	t.Run("from ISBN-13 978 prefix", func(t *testing.T) {
		b, _ := isbn.Parse("9780306406157")
		got, err := b.ToISBN10()
		if err != nil {
			t.Fatalf("ToISBN10() unexpected error: %v", err)
		}
		if got.Digits != "0306406152" {
			t.Errorf("ToISBN10() Digits = %q, want %q", got.Digits, "0306406152")
		}
		if !got.IsISBN10() {
			t.Errorf("ToISBN10() Version = %v, want V10", got.Version)
		}
	})

	t.Run("already ISBN-10", func(t *testing.T) {
		b, _ := isbn.Parse("0306406152")
		got, err := b.ToISBN10()
		if err != nil {
			t.Fatalf("ToISBN10() unexpected error: %v", err)
		}
		if got.Digits != "0306406152" {
			t.Errorf("ToISBN10() Digits = %q, want %q", got.Digits, "0306406152")
		}
	})

	t.Run("979 prefix has no ISBN-10 equivalent", func(t *testing.T) {
		b, _ := isbn.Parse("9791032301951")
		_, err := b.ToISBN10()
		if !errors.Is(err, isbn.ErrNoISBN10Equivalent) {
			t.Errorf("ToISBN10() error = %v, want %v", err, isbn.ErrNoISBN10Equivalent)
		}
	})
}

func TestToISBN13(t *testing.T) {
	t.Run("from ISBN-10", func(t *testing.T) {
		b, _ := isbn.Parse("0306406152")
		got := b.ToISBN13()
		if got.Digits != "9780306406157" {
			t.Errorf("ToISBN13() Digits = %q, want %q", got.Digits, "9780306406157")
		}
		if !got.IsISBN13() {
			t.Errorf("ToISBN13() Version = %v, want V13", got.Version)
		}
	})

	t.Run("already ISBN-13", func(t *testing.T) {
		b, _ := isbn.Parse("9780306406157")
		got := b.ToISBN13()
		if got.Digits != "9780306406157" {
			t.Errorf("ToISBN13() Digits = %q, want %q", got.Digits, "9780306406157")
		}
	})
}

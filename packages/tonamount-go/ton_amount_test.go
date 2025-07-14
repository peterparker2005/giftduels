package tonamount_test

import (
	"encoding/json"
	"testing"

	"github.com/peterparker2005/giftduels/packages/tonamount-go"
)

func TestNewTonAmountFromString(t *testing.T) {
	cases := []struct {
		input       string
		wantString  string
		wantNano    uint64
		expectError bool
	}{
		{"1.23", "1.23", 1230000000, false},
		{"1.2", "1.2", 1200000000, false},
		{"2.00", "2", 2000000000, false},
		{"0", "0", 0, false},
		{"-1.23", "", 0, true},
		{"foo", "", 0, true},
	}

	for _, c := range cases {
		amt, err := tonamount.NewTonAmountFromString(c.input)
		if c.expectError {
			if err == nil {
				t.Errorf("NewTonAmountFromString(%q) expected error, got none", c.input)
			}
			continue
		}
		if err != nil {
			t.Errorf("NewTonAmountFromString(%q) unexpected error: %v", c.input, err)
			continue
		}
		if got := amt.String(); got != c.wantString {
			t.Errorf("String() = %q; want %q", got, c.wantString)
		}
		nano, err := amt.ToNano()
		if err != nil {
			t.Errorf("ToNano(%q) unexpected error: %v", c.input, err)
			continue
		}
		if nano != c.wantNano {
			t.Errorf("ToNano(%q) = %d; want %d", c.input, nano, c.wantNano)
		}
	}
}

func TestNewTonAmountFromFloat64(t *testing.T) {
	cases := []struct {
		input      float64
		wantString string
		wantNano   uint64
	}{
		{1.235, "1.24", 1240000000},
		{1.234, "1.23", 1230000000},
		{0.0, "0", 0},
	}

	for _, c := range cases {
		amt, err := tonamount.NewTonAmountFromFloat64(c.input)
		if err != nil {
			t.Errorf("NewTonAmountFromFloat64(%v) unexpected error: %v", c.input, err)
			continue
		}
		if got := amt.String(); got != c.wantString {
			t.Errorf("String() = %q; want %q", got, c.wantString)
		}
		nano, _ := amt.ToNano()
		if nano != c.wantNano {
			t.Errorf("ToNano(%v) = %d; want %d", c.input, nano, c.wantNano)
		}
	}
}

func TestNewTonAmountFromNano(t *testing.T) {
	cases := []struct {
		inputNano  uint64
		wantString string
	}{
		{1230000000, "1.23"},
		{1200000000, "1.2"},
		{2000000000, "2"},
		{0, "0"},
	}

	for _, c := range cases {
		amt, err := tonamount.NewTonAmountFromNano(c.inputNano)
		if err != nil {
			t.Errorf("NewTonAmountFromNano(%d) unexpected error: %v", c.inputNano, err)
			continue
		}
		if got := amt.String(); got != c.wantString {
			t.Errorf("String() = %q; want %q", got, c.wantString)
		}
	}
}

func TestAddSubNegateZero(t *testing.T) {
	a, _ := tonamount.NewTonAmountFromString("1.5")
	b, _ := tonamount.NewTonAmountFromString("0.75")
	sum := a.Add(b)
	if got := sum.String(); got != "2.25" {
		t.Errorf("Add: got %q; want %q", got, "2.25")
	}
	diff := a.Sub(b)
	if got := diff.String(); got != "0.75" {
		t.Errorf("Sub: got %q; want %q", got, "0.75")
	}
	neg := a.Negate()
	if got := neg.String(); got != "-1.5" {
		t.Errorf("Negate: got %q; want %q", got, "-1.5")
	}
	zero, _ := tonamount.NewTonAmountFromString("0")
	if !zero.IsZero() {
		t.Errorf("IsZero: expected true for zero amount")
	}
}

func TestJSONMarshalling(t *testing.T) {
	orig, _ := tonamount.NewTonAmountFromString("3.50")
	data, err := json.Marshal(orig)
	if err != nil {
		t.Fatalf("MarshalJSON error: %v", err)
	}
	// Expect quoted string
	want := `"3.5"`
	if string(data) != want {
		t.Errorf("MarshalJSON = %s; want %s", data, want)
	}
	var unmarshaled tonamount.TonAmount
	if err = json.Unmarshal(data, &unmarshaled); err != nil {
		t.Fatalf("UnmarshalJSON error: %v", err)
	}
	if got := unmarshaled.String(); got != "3.5" {
		t.Errorf("After Unmarshal, String() = %q; want %q", got, "3.5")
	}
}

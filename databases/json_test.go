package databases

import (
	"database/sql/driver"
	"encoding/json"
	"reflect"
	"testing"
)

type testObj struct {
	Foo string `json:"foo"`
	Bar int    `json:"bar"`
}

func TestJSON_Scan_Object(t *testing.T) {
	var j JSON[testObj]

	// NULL -> zero value, Valid=false (default)
	if err := j.Scan(nil); err != nil {
		t.Fatalf("Scan(nil) error = %v", err)
	}
	if got := j.V; !reflect.DeepEqual(got, testObj{}) {
		t.Fatalf("Scan(nil) V = %+v, want zero", got)
	}
	if j.Valid {
		t.Fatalf("Scan(nil) Valid = true, want false")
	}

	// "" -> zero value, Valid=false
	if err := j.Scan(""); err != nil {
		t.Fatalf(`Scan("") error = %v`, err)
	}
	if got := j.V; !reflect.DeepEqual(got, testObj{}) {
		t.Fatalf(`Scan("") V = %+v, want zero`, got)
	}
	if j.Valid {
		t.Fatalf("Scan(\"\") Valid = true, want false")
	}

	// "null" -> zero value, Valid=false
	if err := j.Scan("null"); err != nil {
		t.Fatalf(`Scan("null") error = %v`, err)
	}
	if got := j.V; !reflect.DeepEqual(got, testObj{}) {
		t.Fatalf(`Scan("null") V = %+v, want zero`, got)
	}
	if j.Valid {
		t.Fatalf(`Scan("null") Valid = true, want false`)
	}

	// Valid JSON -> terisi & Valid=true
	raw := `{"foo":"x","bar":42}`
	if err := j.Scan(raw); err != nil {
		t.Fatalf("Scan(valid JSON) error = %v", err)
	}
	want := testObj{Foo: "x", Bar: 42}
	if !reflect.DeepEqual(j.V, want) {
		t.Fatalf("Scan(valid JSON) V = %+v, want %+v", j.V, want)
	}
	if !j.Valid {
		t.Fatalf("Scan(valid JSON) Valid = false, want true")
	}
}

func TestJSON_Scan_ObjectPtr(t *testing.T) {
	var j JSON[*testObj]

	// NULL -> zero value, Valid=false (default)
	if err := j.Scan(nil); err != nil {
		t.Fatalf("Scan(nil) error = %v", err)
	}
	if j.V != nil {
		t.Fatalf("Scan(nil) V = %+v, want nil", j.V)
	}
	if j.Valid {
		t.Fatalf("Scan(nil) Valid = true, want false")
	}

	// "" -> zero value, Valid=false
	if err := j.Scan(""); err != nil {
		t.Fatalf(`Scan("") error = %v`, err)
	}
	if j.V != nil {
		t.Fatalf("Scan(nil) V = %+v, want nil", j.V)
	}
	if j.Valid {
		t.Fatalf("Scan(\"\") Valid = true, want false")
	}

	// "null" -> zero value, Valid=false
	if err := j.Scan("null"); err != nil {
		t.Fatalf(`Scan("null") error = %v`, err)
	}
	if j.V != nil {
		t.Fatalf("Scan(nil) V = %+v, want nil", j.V)
	}
	if j.Valid {
		t.Fatalf(`Scan("null") Valid = true, want false`)
	}

	// Valid JSON -> terisi & Valid=true
	raw := `{"foo":"x","bar":42}`
	if err := j.Scan(raw); err != nil {
		t.Fatalf("Scan(valid JSON) error = %v", err)
	}
	want := testObj{Foo: "x", Bar: 42}
	if !reflect.DeepEqual(j.V, &want) {
		t.Fatalf("Scan(valid JSON) V = %+v, want %+v", j.V, want)
	}
	if !j.Valid {
		t.Fatalf("Scan(valid JSON) Valid = false, want true")
	}
}

func TestJSON_Value_Object(t *testing.T) {
	{
		j := JSON[*testObj]{}
		val, err := j.Value()
		if err != nil {
			t.Fatalf("Value() error = %v", err)
		}
		if val != nil {
			t.Fatalf("must be nil")
		}
	}

	// struct zero value -> bukan nil (akan di-marshal jadi "{}")
	{
		j := JSON[testObj]{} // V zero
		val, err := j.Value()
		if err != nil {
			t.Fatalf("Value() error = %v", err)
		}
		// Value harus bertipe string (atau []byte) JSON
		switch v := val.(type) {
		case string:
			if v != "{\"foo\":\"\",\"bar\":0}" {
				t.Fatalf("Value() = %q, want {}", v)
			}
		case []byte:
			if string(v) != "{\"foo\":\"\",\"bar\":0}" {
				t.Fatalf("Value() = %q, want {}", v)
			}
		default:
			t.Fatalf("Value() type = %T, want string/[]byte", val)
		}
	}

	// non-zero struct -> JSON berisi field
	{
		j := JSON[testObj]{V: testObj{Foo: "a", Bar: 7}}
		val, err := j.Value()
		if err != nil {
			t.Fatalf("Value() error = %v", err)
		}
		var back testObj
		switch v := val.(type) {
		case string:
			if err := json.Unmarshal([]byte(v), &back); err != nil {
				t.Fatalf("unmarshal back error = %v", err)
			}
		case []byte:
			if err := json.Unmarshal(v, &back); err != nil {
				t.Fatalf("unmarshal back error = %v", err)
			}
		default:
			t.Fatalf("Value() type = %T, want string/[]byte", val)
		}
		if !reflect.DeepEqual(back, j.V) {
			t.Fatalf("roundtrip back = %+v, want %+v", back, j.V)
		}
	}

	// Untuk tipe slice nil: bergantung pada utils/reflection.IsNil(V)
	// Jika IsNil(nilSlice) == true maka Value() -> (nil,nil)
	{
		var j JSON[[]string] // V == nil slice
		val, err := j.Value()
		if err != nil {
			t.Fatalf("Value() error = %v", err)
		}
		if val != nil {
			t.Fatalf("Value() for nil slice = %v, want nil (SQL NULL)", val)
		}
	}

	// Untuk tipe slice empty non-nil: Value() -> "[]"
	{
		j := JSON[[]string]{V: []string{}}
		val, err := j.Value()
		if err != nil {
			t.Fatalf("Value() error = %v", err)
		}
		switch v := val.(type) {
		case string:
			if v != "[]" {
				t.Fatalf("Value() = %q, want []", v)
			}
		case []byte:
			if string(v) != "[]" {
				t.Fatalf("Value() = %q, want []", v)
			}
		default:
			t.Fatalf("Value() type = %T, want string/[]byte", val)
		}
	}
}

func TestJSON_MarshalFrom_Into_Unmarshal(t *testing.T) {
	var j JSON[testObj]
	j.MarshalFrom(testObj{Foo: "k", Bar: 9})

	// Into
	got := j.Into()
	want := testObj{Foo: "k", Bar: 9}
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("Into() = %+v, want %+v", got, want)
	}

	// Value -> Scan roundtrip
	val, err := j.Value()
	if err != nil {
		t.Fatalf("Value() error = %v", err)
	}
	var raw []byte
	switch v := val.(type) {
	case string:
		raw = []byte(v)
	case []byte:
		raw = v
	default:
		t.Fatalf("Value() type = %T, want string/[]byte", val)
	}
	var j2 JSON[testObj]
	if err := j2.Scan(raw); err != nil {
		t.Fatalf("Scan(roundtrip) error = %v", err)
	}
	if !reflect.DeepEqual(j2.V, want) || !j2.Valid {
		t.Fatalf("Scan(roundtrip) got %+v (Valid=%v), want %+v (Valid=true)", j2.V, j2.Valid, want)
	}
}

func TestJSON_IntoPtr(t *testing.T) {
	// Case 1: V adalah nil slice -> IntoPtr harus mengembalikan pointer non-nil ke nilai nil slice (sesuai implementasi)
	{
		var j JSON[[]string] // V == nil
		p := j.IntoPtr()
		if p == nil {
			t.Fatalf("IntoPtr(nil slice) = nil, want non-nil pointer to nil slice")
		}
		if *p != nil {
			t.Fatalf("IntoPtr(nil slice) points to %v, want nil slice", *p)
		}
	}

	// Case 2: V adalah zero struct -> IsZero true -> IntoPtr() mengembalikan nil
	{
		j := JSON[testObj]{V: testObj{}}
		if got := j.IntoPtr(); got != nil {
			t.Fatalf("IntoPtr(zero struct) = %v, want nil", got)
		}
	}

	// Case 3: V non-zero -> pointer ke salinan
	{
		j := JSON[testObj]{V: testObj{Foo: "ok", Bar: 1}}
		p := j.IntoPtr()
		if p == nil {
			t.Fatalf("IntoPtr(non-zero) = nil, want non-nil")
		}
		if !reflect.DeepEqual(*p, j.V) {
			t.Fatalf("IntoPtr value = %+v, want %+v", *p, j.V)
		}
		// pastikan bukan alias langsung (opsional)
		p.Foo = "changed"
		if reflect.DeepEqual(*p, j.V) && p == &j.V {
			t.Fatalf("IntoPtr returned alias of internal value; want copy")
		}
	}
}

func TestJSON_IsZero(t *testing.T) {
	if !(JSON[testObj]{V: testObj{}}).IsZero() {
		t.Fatalf("IsZero(zero struct) = false, want true")
	}
	if (JSON[testObj]{V: testObj{Foo: "x"}}).IsZero() {
		t.Fatalf("IsZero(non-zero struct) = true, want false")
	}
	if !(JSON[[]string]{V: nil}).IsZero() {
		t.Fatalf("IsZero(nil slice) = false, want true")
	}
	if (JSON[[]string]{V: []string{}}).IsZero() {
		t.Fatalf("IsZero(empty slice non-nil) = true, want false (empty slice != zero if your IsZero treats non-nil empty as non-zero)")
	}
}

// Tambahan: Value() harus kompatibel dengan driver.Value
func TestJSON_Value_DriverCompatibility(t *testing.T) {
	j := JSON[testObj]{V: testObj{Foo: "x"}}
	val, err := j.Value()
	if err != nil {
		t.Fatalf("Value() error = %v", err)
	}
	switch val.(type) {
	case nil, string, []byte, bool, int64, float64, driver.Valuer:
		// ok
	default:
		t.Fatalf("Value() produced unsupported driver.Value type: %T", val)
	}
}

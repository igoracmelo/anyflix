package util

import (
	"database/sql/driver"
	"encoding/json"
)

type Opt[T any] struct {
	Val T
	Ok  bool
}

var _ json.Marshaler = Opt[string]{}
var _ json.Unmarshaler = &Opt[string]{}

// var _ sql.Scanner = &Opt[string]{}
var _ driver.Valuer = Opt[string]{}

// MarshalJSON implements json.Marshaler.
func (o Opt[T]) MarshalJSON() ([]byte, error) {
	if o.Ok {
		return json.Marshal(o.Val)
	}
	var null *T
	return json.Marshal(null)
}

// UnmarshalJSON implements json.Unmarshaler.
func (o *Opt[T]) UnmarshalJSON(b []byte) error {
	if string(b) == "null" {
		o.Ok = false
		var v T
		o.Val = v
	}

	var v T
	err := json.Unmarshal(b, &v)
	if err != nil {
		return err
	}

	o.Val = v
	o.Ok = any(v) != nil

	return nil
}

// Value implements driver.Valuer.
func (o Opt[T]) Value() (driver.Value, error) {
	if o.Ok {
		return o.Val, nil
	}
	return nil, nil
}

// Scan implements sql.Scanner.
func (o *Opt[T]) Scan(src any) error {
	panic("todo")
}

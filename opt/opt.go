package opt

import (
	"fmt"
	"strconv"
)

type Opt[T any] struct {
	Value T
	Ok    bool
}

func (o Opt[T]) Or(v T) T {
	if o.Ok {
		return o.Value
	}
	return v
}

func (o Opt[T]) String() string {
	if o.Ok {
		return fmt.Sprint(o.Value)
	}
	return ""
}

func ParseInt(s string) Opt[int] {
	i, err := strconv.Atoi(s)
	return Opt[int]{
		Value: i,
		Ok:    err == nil,
	}
}

func ParseFloat(s string) Opt[float64] {
	f, err := strconv.ParseFloat(s, 64)
	return Opt[float64]{
		Value: f,
		Ok:    err == nil,
	}
}

func String(s string) Opt[string] {
	return Opt[string]{
		Value: s,
		Ok:    s != "",
	}
}

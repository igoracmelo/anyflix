package th

import (
	"fmt"
	"testing"
)

var Assert = assert{
	failNow: true,
}

var Test = assert{
	failNow: false,
}

type assert struct {
	failNow bool
}

func (assert assert) True(t *testing.T, cond bool, msg string) {
	t.Helper()

	if !cond {
		if assert.failNow {
			t.Fatal(msg)
		} else {
			t.Error(msg)
		}
	}
}

func (assert assert) Equal(t *testing.T, a, b any) {
	assert.True(t, a == b, fmt.Sprintf("not equal:\na: %v\nb: %v", a, b))
}

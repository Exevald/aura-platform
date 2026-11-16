package integrationtest

import (
	"errors"
	"reflect"
	"testing"
)

func TestCloseRunsEveryCloserInReverseOrder(t *testing.T) {
	env := New()
	var order []int
	first, second := errors.New("first"), errors.New("second")
	env.AddCloser(func() error { order = append(order, 1); return first })
	env.AddCloser(func() error { order = append(order, 2); return second })
	err := env.Close()
	if !reflect.DeepEqual(order, []int{2, 1}) {
		t.Fatalf("close order = %v", order)
	}
	if !errors.Is(err, first) || !errors.Is(err, second) {
		t.Fatalf("close error = %v", err)
	}
}

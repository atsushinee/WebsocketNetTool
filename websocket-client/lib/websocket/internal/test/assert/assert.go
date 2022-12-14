package assert

import (
	"fmt"
	"reflect"
	"strings"
	"testing"

	"github.com/golang/protobuf/proto"
	"github.com/google/go-cmp/cmp"
	"github.com/google/go-cmp/cmp/cmpopts"
)


func Diff(v1, v2 interface{}) string {
	return cmp.Diff(v1, v2, cmpopts.EquateErrors(), cmp.Exporter(func(r reflect.Type) bool {
		return true
	}), cmp.Comparer(proto.Equal))
}


func Equal(t testing.TB, name string, exp, act interface{}) {
	t.Helper()

	if diff := Diff(exp, act); diff != "" {
		t.Fatalf("unexpected %v: %v", name, diff)
	}
}


func Success(t testing.TB, err error) {
	t.Helper()

	if err != nil {
		t.Fatal(err)
	}
}


func Error(t testing.TB, err error) {
	t.Helper()

	if err == nil {
		t.Fatal("expected error")
	}
}


func Contains(t testing.TB, v interface{}, sub string) {
	t.Helper()

	s := fmt.Sprint(v)
	if !strings.Contains(s, sub) {
		t.Fatalf("expected %q to contain %q", s, sub)
	}
}

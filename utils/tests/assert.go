package test_utils

import (
	"encoding/json"
	"fmt"
	"slices"
	"testing"
)

func AssertError(t testing.TB, got, want error) {
	t.Helper()
	if want == nil {
		return
	}
	if got == nil {
		t.Errorf("got nil, want %q", want.Error())
		return
	}
	wantStr := want.Error()
	gotStr := got.Error()

	if wantStr != gotStr {
		t.Errorf("got %q, want %q", gotStr, wantStr)
	}
}

func AssertSliceEqual[E comparable](t testing.TB, got, want []E) {
	t.Helper()

	if !slices.Equal(got, want) {
		gotStr, err := json.MarshalIndent(got, "", "  ")
		if err != nil {
			gotStr = []byte(fmt.Sprintf("%v", got))
		}
		wantStr, err := json.MarshalIndent(want, "", "  ")
		if err != nil {
			wantStr = []byte(fmt.Sprintf("%v", want))
		}
		t.Errorf("\ngot %s,\nwant %s", gotStr, wantStr)
	}
}

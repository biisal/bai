package test_utils

import "testing"

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

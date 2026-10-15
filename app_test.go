package main

import (
	"errors"
	"testing"

	"github.com/wailsapp/wails/v2/pkg/runtime"
)

func TestNativeCloseDialog(t *testing.T) {
	options := closeDialogOptions()
	if options.Type != runtime.QuestionDialog || options.DefaultButton != "No" || options.CancelButton != "No" || len(options.Buttons) != 2 || options.Buttons[0] != "Yes" || options.Buttons[1] != "No" {
		t.Fatalf("close dialog must use native Yes/No with safe defaults: %+v", options)
	}
	for _, tc := range []struct {
		choice  string
		err     error
		prevent bool
	}{
		{"Yes", nil, false},
		{"No", nil, true},
		{"", nil, true},
		{"Cancel", nil, true},
		{"Close Querylight", nil, true},
		{"Yes", errors.New("dialog failed"), true},
	} {
		if got := preventClose(tc.choice, tc.err); got != tc.prevent {
			t.Errorf("choice %q error %v: prevent close = %v, want %v", tc.choice, tc.err, got, tc.prevent)
		}
	}
}

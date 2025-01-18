package handlers_test

import (
	"jtracker-backend/handlers"
	"net/http"
	"testing"
)

func TestCallbackHandler(t *testing.T) {
	tests := []struct {
		name string // description of this test case
		// Named input parameters for target function.
		w http.ResponseWriter
		r *http.Request
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			handlers.CallbackHandler(tt.w, tt.r)
		})
	}
}

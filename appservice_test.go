//go:build !server

package main

import (
	"encoding/json"
	"testing"
)

func TestCommandErrorSerializesErrorKey(t *testing.T) {
	err := newCommandError(errInvalidPort, map[string]any{"min": 1})
	var body struct {
		Error  string         `json:"error"`
		Params map[string]any `json:"params"`
	}
	if decodeErr := json.Unmarshal([]byte(err.Error()), &body); decodeErr != nil {
		t.Fatal(decodeErr)
	}
	if body.Error != "invalid_port" {
		t.Fatalf("error key = %q", body.Error)
	}
	if body.Params["min"].(float64) != 1 {
		t.Fatalf("params = %#v", body.Params)
	}
}

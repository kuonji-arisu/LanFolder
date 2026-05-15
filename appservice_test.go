package main

import (
	"encoding/json"
	"errors"
	"testing"

	"github.com/wailsapp/wails/v3/pkg/application"
)

func TestCommandErrorMessageIsStableCode(t *testing.T) {
	err := newCommandError(errInvalidPort, map[string]any{"min": 1})
	if err.Error() != "invalid_port" {
		t.Fatalf("error string = %q", err.Error())
	}
}

func TestMarshalCommandErrorSerializesErrorKey(t *testing.T) {
	err := newCommandError(errInvalidPort, map[string]any{"min": 1})
	var body struct {
		Error  string         `json:"error"`
		Params map[string]any `json:"params"`
	}
	if decodeErr := json.Unmarshal(marshalCommandError(err), &body); decodeErr != nil {
		t.Fatal(decodeErr)
	}
	if body.Error != "invalid_port" {
		t.Fatalf("error key = %q", body.Error)
	}
	if body.Params["min"].(float64) != 1 {
		t.Fatalf("params = %#v", body.Params)
	}
}

func TestMarshalCommandErrorFallsBackForUnknownErrors(t *testing.T) {
	if data := marshalCommandError(errors.New("boom")); data != nil {
		t.Fatalf("unknown error marshal = %s, want nil", data)
	}
}

func TestSingleInstanceOptions(t *testing.T) {
	options := singleInstanceOptions(&AppService{})
	if options == nil {
		t.Fatal("single instance options should not be nil")
	}
	if options.UniqueID != appSingleInstanceID {
		t.Fatalf("unique ID = %q, want %q", options.UniqueID, appSingleInstanceID)
	}
	if options.OnSecondInstanceLaunch == nil {
		t.Fatal("second instance callback should be set")
	}
	options.OnSecondInstanceLaunch(application.SecondInstanceData{})
}

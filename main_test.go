package main

import (
	"testing"

	"github.com/wailsapp/wails/v3/pkg/application"

	"lanfolder/internal/appservice"
)

func TestSingleInstanceOptions(t *testing.T) {
	options := singleInstanceOptions(&appservice.AppService{})
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

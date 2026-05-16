package main

import (
	"encoding/json"
	"errors"
	"os"
	"path/filepath"
	"testing"

	"github.com/wailsapp/wails/v3/pkg/application"

	"lanfolder/internal/config"
	"lanfolder/internal/desktop"
	"lanfolder/internal/server"
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

func TestCommandErrorPayloadReusesDesktopErrorPayload(t *testing.T) {
	payload := commandErrorPayload(newCommandError(errInvalidPort, map[string]any{"min": 1}))
	if payload == nil {
		t.Fatal("expected command error payload")
	}
	if payload.Error != "invalid_port" {
		t.Fatalf("error = %q", payload.Error)
	}
	if payload.Params["min"].(int) != 1 {
		t.Fatalf("params = %#v", payload.Params)
	}
}

func TestMarshalCommandErrorFallsBackForUnknownErrors(t *testing.T) {
	if data := marshalCommandError(errors.New("boom")); data != nil {
		t.Fatalf("unknown error marshal = %s, want nil", data)
	}
}

func TestDrainNoticesReturnsAndClearsPendingNotices(t *testing.T) {
	service := &AppService{}
	service.addNotice(desktop.NoticeError, desktop.NoticeSourceStartup, newCommandError(errSharedDirRequired, nil), "")

	notices := service.DrainNotices()
	if len(notices) != 1 {
		t.Fatalf("notices = %d, want 1", len(notices))
	}
	notice := notices[0]
	if notice.ID == "" {
		t.Fatal("notice ID should not be empty")
	}
	if notice.Level != desktop.NoticeError || notice.Source != desktop.NoticeSourceStartup {
		t.Fatalf("notice = %#v", notice)
	}
	if notice.Error == nil || notice.Error.Error != "shared_dir_required" {
		t.Fatalf("notice error = %#v", notice.Error)
	}
	if again := service.DrainNotices(); len(again) != 0 {
		t.Fatalf("drained notices remained: %#v", again)
	}
}

func TestNoticeQueueKeepsMostRecentNotices(t *testing.T) {
	service := &AppService{}
	for range 55 {
		service.addNotice(desktop.NoticeError, desktop.NoticeSourceStartup, errors.New("boom"), "")
	}

	notices := service.DrainNotices()
	if len(notices) != 50 {
		t.Fatalf("notices = %d, want 50", len(notices))
	}
	if notices[0].ID != "6" || notices[len(notices)-1].ID != "55" {
		t.Fatalf("notice IDs = %q..%q, want 6..55", notices[0].ID, notices[len(notices)-1].ID)
	}
}

func TestAddNoticeDoesNotRequeueAfterDrain(t *testing.T) {
	service := &AppService{}
	service.addNotice(desktop.NoticeError, desktop.NoticeSourceStartup, errors.New("first"), "")
	if notices := service.DrainNotices(); len(notices) != 1 {
		t.Fatalf("notices = %d, want 1", len(notices))
	}

	service.addNotice(desktop.NoticeError, desktop.NoticeSourceStartup, errors.New("second"), "")
	if notices := service.DrainNotices(); len(notices) != 0 {
		t.Fatalf("post-drain notices = %#v, want none", notices)
	}
}

func TestAutoStartSharingReportsMissingSharedDir(t *testing.T) {
	service := &AppService{config: config.Config{AutoShare: true, AccessApproval: true}}
	service.autoStartSharing()

	notices := service.DrainNotices()
	if len(notices) != 1 {
		t.Fatalf("notices = %d, want 1", len(notices))
	}
	if notices[0].Level != desktop.NoticeError || notices[0].Error == nil || notices[0].Error.Error != "shared_dir_required" {
		t.Fatalf("notice = %#v", notices[0])
	}
}

func TestAutoStartSharingReportsStartFailure(t *testing.T) {
	root := t.TempDir()
	missing := filepath.Join(root, "missing")
	service := &AppService{
		server: server.New(os.DirFS(root)),
		config: config.Config{AutoShare: true, AccessApproval: true, SharedDir: missing, Port: 8899},
	}
	service.autoStartSharing()

	notices := service.DrainNotices()
	if len(notices) != 1 {
		t.Fatalf("notices = %d, want 1", len(notices))
	}
	if notices[0].Level != desktop.NoticeError || notices[0].Error != nil || notices[0].Message != "" {
		t.Fatalf("notice = %#v", notices[0])
	}
}

func TestAutoStartSharingRequiresAccessApproval(t *testing.T) {
	service := &AppService{config: config.Config{AutoShare: true, SharedDir: "C:/Share"}}
	service.autoStartSharing()

	notices := service.DrainNotices()
	if len(notices) != 1 {
		t.Fatalf("notices = %d, want 1", len(notices))
	}
	if notices[0].Level != desktop.NoticeError || notices[0].Error == nil || notices[0].Error.Error != "access_approval_required" {
		t.Fatalf("notice = %#v", notices[0])
	}
}

func TestSaveSettingsRejectsAutoShareWithoutAccessApproval(t *testing.T) {
	service := &AppService{
		server: server.New(os.DirFS(t.TempDir())),
		config: config.Default(),
	}

	_, err := service.SaveSettings(config.Config{Port: 8899, AutoShare: true})
	if err == nil {
		t.Fatal("expected access approval error")
	}
	if err.Error() != "access_approval_required" {
		t.Fatalf("error = %v", err)
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

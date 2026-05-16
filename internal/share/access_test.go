package share

import (
	"testing"
	"time"
)

func TestAccessPendingRequestExpiresBeforeApprove(t *testing.T) {
	manager := NewAccessManager()
	now := time.Date(2026, 5, 16, 12, 0, 0, 0, time.UTC)
	manager.now = func() time.Time { return now }

	req, err := manager.CreateRequest("192.168.1.20", "test")
	if err != nil {
		t.Fatal(err)
	}
	now = now.Add(AccessRequestTTL + time.Second)

	if err := manager.Approve(req.ID); err != ErrAccessRequestNotFound {
		t.Fatalf("approve expired request error = %v, want %v", err, ErrAccessRequestNotFound)
	}
	result, token := manager.Poll(req.ID)
	if result.State != AccessPollExpired || token != "" {
		t.Fatalf("poll expired = %#v token %q", result, token)
	}
}

func TestAccessApproveCreatesValidSessionToken(t *testing.T) {
	manager := NewAccessManager()
	req, err := manager.CreateRequest("192.168.1.20", "test")
	if err != nil {
		t.Fatal(err)
	}
	if err := manager.Approve(req.ID); err != nil {
		t.Fatal(err)
	}
	result, token := manager.Poll(req.ID)
	if result.State != AccessPollApproved || token == "" {
		t.Fatalf("poll approved = %#v token %q", result, token)
	}
	if !manager.Validate(token) {
		t.Fatal("approved token should validate")
	}
	manager.Clear()
	if manager.Validate(token) {
		t.Fatal("cleared token should not validate")
	}
}

package share

import (
	"strconv"
	"testing"
	"time"
)

func TestAccessPendingRequestExpiresBeforeApprove(t *testing.T) {
	manager := NewAccessManager()
	now := time.Date(2026, 5, 16, 12, 0, 0, 0, time.UTC)
	manager.now = func() time.Time { return now }
	token := "request-token"

	req, _, err := manager.CreateRequest(token, "192.168.1.20", "test")
	if err != nil {
		t.Fatal(err)
	}
	now = now.Add(AccessRequestTTL + time.Second)

	if _, err := manager.Approve(req.ID); err != ErrAccessRequestNotFound {
		t.Fatalf("approve expired request error = %v, want %v", err, ErrAccessRequestNotFound)
	}
	result, sessionToken, err := manager.Poll(token)
	if err != nil {
		t.Fatal(err)
	}
	if result.State != AccessPollExpired || sessionToken != "" {
		t.Fatalf("poll expired = %#v token %q", result, sessionToken)
	}
}

func TestAccessApproveCreatesValidSessionToken(t *testing.T) {
	manager := NewAccessManager()
	requestToken := "request-token"
	req, _, err := manager.CreateRequest(requestToken, "192.168.1.20", "test")
	if err != nil {
		t.Fatal(err)
	}
	if _, err := manager.Approve(req.ID); err != nil {
		t.Fatal(err)
	}
	if sessions := manager.Sessions(); len(sessions) != 0 {
		t.Fatalf("sessions before poll = %#v, want none", sessions)
	}
	result, token, err := manager.Poll(requestToken)
	if err != nil {
		t.Fatal(err)
	}
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

func TestAccessCreateRequestDedupesPendingRequestToken(t *testing.T) {
	manager := NewAccessManager()
	token := "request-token"
	first, created, err := manager.CreateRequest(token, "192.168.1.20", "test")
	if err != nil {
		t.Fatal(err)
	}
	if !created {
		t.Fatal("first request should be created")
	}
	second, created, err := manager.CreateRequest(token, "192.168.1.20", "changed")
	if err != nil {
		t.Fatal(err)
	}
	if created {
		t.Fatal("duplicate request should reuse existing pending request")
	}
	if second.ID != first.ID {
		t.Fatalf("duplicate request = %#v, want %#v", second, first)
	}
	if second.RequestCount != 2 || second.UserAgent != "changed" {
		t.Fatalf("duplicate metadata = %#v, want count 2 with latest user agent", second)
	}
	if pending := manager.Pending(); len(pending) != 1 {
		t.Fatalf("pending = %#v, want one request", pending)
	}
}

func TestAccessCreateRequestAllowsSameIPWithDifferentRequestTokens(t *testing.T) {
	manager := NewAccessManager()
	first, created, err := manager.CreateRequest("request-token-1", "192.168.1.20", "first")
	if err != nil {
		t.Fatal(err)
	}
	if !created {
		t.Fatal("first request should be created")
	}
	second, created, err := manager.CreateRequest("request-token-2", "192.168.1.20", "second")
	if err != nil {
		t.Fatal(err)
	}
	if !created {
		t.Fatal("second request should be created")
	}
	if second.ID == first.ID {
		t.Fatalf("same IP requests should have different IDs: %q", second.ID)
	}
	if pending := manager.Pending(); len(pending) != 2 {
		t.Fatalf("pending = %#v, want two requests", pending)
	}
}

func TestAccessCreateRequestCooldownAfterDeny(t *testing.T) {
	manager := NewAccessManager()
	now := time.Date(2026, 5, 16, 12, 0, 0, 0, time.UTC)
	manager.now = func() time.Time { return now }

	req, _, err := manager.CreateRequest("request-token-1", "192.168.1.20", "test")
	if err != nil {
		t.Fatal(err)
	}
	if _, err := manager.Deny(req.ID); err != nil {
		t.Fatal(err)
	}
	if _, _, err := manager.CreateRequest("request-token-2", "192.168.1.20", "test"); err != ErrAccessRequestLimited {
		t.Fatalf("cooldown err = %v, want %v", err, ErrAccessRequestLimited)
	}

	now = now.Add(AccessRequestCooldown)
	if _, created, err := manager.CreateRequest("request-token-2", "192.168.1.20", "test"); err != nil || !created {
		t.Fatalf("post-cooldown created=%v err=%v", created, err)
	}
}

func TestAccessCreateRequestLimitsPendingCount(t *testing.T) {
	manager := NewAccessManager()
	for i := 0; i < MaxAccessPendingCount; i++ {
		if _, created, err := manager.CreateRequest("request-token-"+strconv.Itoa(i+1), "192.168.1."+strconv.Itoa(i+1), "test"); err != nil || !created {
			t.Fatalf("request %d created=%v err=%v", i, created, err)
		}
	}
	if _, _, err := manager.CreateRequest("request-token-overflow", "192.168.2.1", "overflow"); err != ErrAccessRequestLimited {
		t.Fatalf("overflow err = %v, want %v", err, ErrAccessRequestLimited)
	}
}

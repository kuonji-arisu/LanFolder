package share

import (
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"errors"
	"sort"
	"sync"
	"time"
)

const (
	AccessRequestTTL      = 2 * time.Minute
	AccessRequestCooldown = 5 * time.Second
	MaxAccessPendingCount = 32
	MaxAccessPendingPerIP = 4
)

type AccessSessionLifetime string

const (
	AccessSession10Minutes AccessSessionLifetime = "10m"
	AccessSession30Minutes AccessSessionLifetime = "30m"
	AccessSession1Hour     AccessSessionLifetime = "1h"
	AccessSession1Day      AccessSessionLifetime = "24h"
	AccessSessionNever     AccessSessionLifetime = "never"
)

func (l AccessSessionLifetime) Valid() bool {
	switch l {
	case AccessSession10Minutes, AccessSession30Minutes, AccessSession1Hour, AccessSession1Day, AccessSessionNever:
		return true
	default:
		return false
	}
}

func (l AccessSessionLifetime) Duration() (time.Duration, bool) {
	switch l {
	case AccessSession10Minutes:
		return 10 * time.Minute, true
	case AccessSession30Minutes:
		return 30 * time.Minute, true
	case AccessSession1Hour:
		return time.Hour, true
	case AccessSession1Day:
		return 24 * time.Hour, true
	default:
		return 0, false
	}
}

func NormalizeAccessSessionLifetime(lifetime AccessSessionLifetime) AccessSessionLifetime {
	if lifetime.Valid() {
		return lifetime
	}
	return AccessSessionNever
}

var (
	ErrAccessRequestNotFound = errors.New("access_request_not_found")
	ErrAccessRequestLimited  = errors.New("access_request_limited")
)

type AccessRequest struct {
	// ID is an internal desktop approval handle. LAN browsers identify the
	// request with the lf_request cookie instead of receiving this value.
	ID           string    `json:"id"`
	IP           string    `json:"ip"`
	UserAgent    string    `json:"userAgent"`
	RequestCount int       `json:"requestCount"`
	CreatedAt    time.Time `json:"createdAt"`
	LastSeenAt   time.Time `json:"lastSeenAt"`
	ExpiresAt    time.Time `json:"expiresAt"`
}

type AccessSession struct {
	// IP and UserAgent are display-only metadata for the desktop approval UI.
	// Validate treats the session cookie as a bearer token and does not bind it
	// to either value.
	ID        string     `json:"id"`
	IP        string     `json:"ip"`
	UserAgent string     `json:"userAgent"`
	CreatedAt time.Time  `json:"createdAt"`
	ExpiresAt *time.Time `json:"expiresAt,omitempty"`
}

type AccessPollState string

const (
	AccessPollPending  AccessPollState = "pending"
	AccessPollApproved AccessPollState = "approved"
	AccessPollDenied   AccessPollState = "denied"
	AccessPollExpired  AccessPollState = "expired"
)

type AccessPollResult struct {
	State AccessPollState `json:"state"`
}

type completedAccessRequest struct {
	state     AccessPollState
	request   AccessRequest
	expiresAt time.Time
}

type AccessManager struct {
	mu            sync.Mutex
	now           func() time.Time
	pending       map[[32]byte]AccessRequest
	pendingID     map[string][32]byte
	completed     map[[32]byte]completedAccessRequest
	sessions      map[[32]byte]AccessSession
	sessionID     map[string][32]byte
	lastRequestAt map[string]time.Time
	sessionTTL    AccessSessionLifetime
}

func NewAccessManager() *AccessManager {
	return &AccessManager{
		now:           time.Now,
		pending:       map[[32]byte]AccessRequest{},
		pendingID:     map[string][32]byte{},
		completed:     map[[32]byte]completedAccessRequest{},
		sessions:      map[[32]byte]AccessSession{},
		sessionID:     map[string][32]byte{},
		lastRequestAt: map[string]time.Time{},
		sessionTTL:    AccessSessionNever,
	}
}

func (m *AccessManager) SetSessionLifetime(lifetime AccessSessionLifetime) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.sessionTTL = NormalizeAccessSessionLifetime(lifetime)
}

func (m *AccessManager) CreateRequest(requestToken, ip, userAgent string) (AccessRequest, bool, error) {
	now := m.now()
	hash := sha256.Sum256([]byte(requestToken))

	m.mu.Lock()
	defer m.mu.Unlock()
	m.pruneExpiredLocked(now)
	if req, ok := m.pending[hash]; ok {
		req.IP = ip
		req.UserAgent = userAgent
		req.RequestCount++
		req.LastSeenAt = now
		m.pending[hash] = req
		return req, false, nil
	}
	if completed, ok := m.completed[hash]; ok {
		if completed.state != AccessPollExpired && completed.request.ID != "" {
			return completed.request, false, nil
		}
		delete(m.completed, hash)
	}
	if last, ok := m.lastRequestAt[ip]; ok && now.Sub(last) < AccessRequestCooldown {
		return AccessRequest{}, false, ErrAccessRequestLimited
	}
	if m.pendingCountForIPLocked(ip) >= MaxAccessPendingPerIP {
		return AccessRequest{}, false, ErrAccessRequestLimited
	}
	if len(m.pending) >= MaxAccessPendingCount {
		return AccessRequest{}, false, ErrAccessRequestLimited
	}
	id, err := accessRandomToken(32)
	if err != nil {
		return AccessRequest{}, false, err
	}
	req := AccessRequest{
		ID:           id,
		IP:           ip,
		UserAgent:    userAgent,
		RequestCount: 1,
		CreatedAt:    now,
		LastSeenAt:   now,
		ExpiresAt:    now.Add(AccessRequestTTL),
	}
	m.pending[hash] = req
	m.pendingID[req.ID] = hash
	return req, true, nil
}

func (m *AccessManager) Pending() []AccessRequest {
	now := m.now()
	m.mu.Lock()
	defer m.mu.Unlock()
	m.pruneExpiredLocked(now)
	out := make([]AccessRequest, 0, len(m.pending))
	for _, req := range m.pending {
		out = append(out, req)
	}
	sort.Slice(out, func(i, j int) bool {
		return out[i].CreatedAt.Before(out[j].CreatedAt)
	})
	return out
}

func (m *AccessManager) Approve(id string) (AccessRequest, error) {
	now := m.now()
	m.mu.Lock()
	defer m.mu.Unlock()
	m.pruneExpiredLocked(now)
	hash, ok := m.pendingID[id]
	if !ok {
		return AccessRequest{}, ErrAccessRequestNotFound
	}
	req, ok := m.pending[hash]
	if !ok {
		delete(m.pendingID, id)
		return AccessRequest{}, ErrAccessRequestNotFound
	}
	delete(m.pending, hash)
	delete(m.pendingID, id)
	m.completed[hash] = completedAccessRequest{
		state:     AccessPollApproved,
		request:   req,
		expiresAt: now.Add(AccessRequestTTL),
	}
	return req, nil
}

func (m *AccessManager) Deny(id string) (AccessRequest, error) {
	now := m.now()
	m.mu.Lock()
	defer m.mu.Unlock()
	m.pruneExpiredLocked(now)
	hash, ok := m.pendingID[id]
	if !ok {
		return AccessRequest{}, ErrAccessRequestNotFound
	}
	req, ok := m.pending[hash]
	if !ok {
		delete(m.pendingID, id)
		return AccessRequest{}, ErrAccessRequestNotFound
	}
	delete(m.pending, hash)
	delete(m.pendingID, id)
	m.lastRequestAt[req.IP] = now
	m.completed[hash] = completedAccessRequest{
		state:     AccessPollDenied,
		request:   req,
		expiresAt: now.Add(AccessRequestTTL),
	}
	return req, nil
}

func (m *AccessManager) Poll(requestToken string) (AccessPollResult, string, AccessSession, error) {
	now := m.now()
	hash := sha256.Sum256([]byte(requestToken))

	m.mu.Lock()
	defer m.mu.Unlock()
	m.pruneExpiredLocked(now)
	if completed, ok := m.completed[hash]; ok {
		if completed.state == AccessPollApproved {
			token, session, err := m.createSessionLocked(completed.request, now)
			if err != nil {
				return AccessPollResult{}, "", AccessSession{}, err
			}
			delete(m.completed, hash)
			return AccessPollResult{State: completed.state}, token, session, nil
		}
		delete(m.completed, hash)
		return AccessPollResult{State: completed.state}, "", AccessSession{}, nil
	}
	if _, ok := m.pending[hash]; ok {
		return AccessPollResult{State: AccessPollPending}, "", AccessSession{}, nil
	}
	return AccessPollResult{State: AccessPollExpired}, "", AccessSession{}, nil
}

func (m *AccessManager) createSessionLocked(req AccessRequest, now time.Time) (string, AccessSession, error) {
	token, err := accessRandomToken(32)
	if err != nil {
		return "", AccessSession{}, err
	}
	sessionID, err := accessRandomToken(16)
	if err != nil {
		return "", AccessSession{}, err
	}
	hash := sha256.Sum256([]byte(token))
	var expiresAt *time.Time
	if ttl, ok := m.sessionTTL.Duration(); ok {
		expires := now.Add(ttl)
		expiresAt = &expires
	}
	session := AccessSession{
		ID:        sessionID,
		IP:        req.IP,
		UserAgent: req.UserAgent,
		CreatedAt: now,
		ExpiresAt: expiresAt,
	}
	m.sessions[hash] = session
	m.sessionID[sessionID] = hash
	return token, session, nil
}

func (m *AccessManager) Validate(token string) bool {
	if token == "" {
		return false
	}
	hash := sha256.Sum256([]byte(token))
	m.mu.Lock()
	defer m.mu.Unlock()
	m.pruneExpiredLocked(m.now())
	session, ok := m.sessions[hash]
	return ok && session.ID != ""
}

func (m *AccessManager) Revoke(token string) {
	if token == "" {
		return
	}
	hash := sha256.Sum256([]byte(token))
	m.mu.Lock()
	defer m.mu.Unlock()
	m.deleteSessionLocked(hash)
}

func (m *AccessManager) RevokeSession(id string) (AccessSession, bool) {
	m.mu.Lock()
	defer m.mu.Unlock()
	hash, ok := m.sessionID[id]
	if !ok {
		return AccessSession{}, false
	}
	session := m.sessions[hash]
	m.deleteSessionLocked(hash)
	return session, true
}

func (m *AccessManager) Sessions() []AccessSession {
	now := m.now()
	m.mu.Lock()
	defer m.mu.Unlock()
	m.pruneExpiredLocked(now)
	out := make([]AccessSession, 0, len(m.sessions))
	for _, session := range m.sessions {
		out = append(out, session)
	}
	sort.Slice(out, func(i, j int) bool {
		return out[i].CreatedAt.Before(out[j].CreatedAt)
	})
	return out
}

func (m *AccessManager) Clear() {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.pending = map[[32]byte]AccessRequest{}
	m.pendingID = map[string][32]byte{}
	m.completed = map[[32]byte]completedAccessRequest{}
	m.sessions = map[[32]byte]AccessSession{}
	m.sessionID = map[string][32]byte{}
	m.lastRequestAt = map[string]time.Time{}
}

func (m *AccessManager) pruneExpiredLocked(now time.Time) {
	for hash, req := range m.pending {
		if !now.Before(req.ExpiresAt) {
			delete(m.pending, hash)
			delete(m.pendingID, req.ID)
			m.completed[hash] = completedAccessRequest{state: AccessPollExpired, expiresAt: now.Add(AccessRequestTTL)}
		}
	}
	for hash, completed := range m.completed {
		if !now.Before(completed.expiresAt) {
			delete(m.completed, hash)
		}
	}
	for hash, session := range m.sessions {
		if session.ExpiresAt != nil && !now.Before(*session.ExpiresAt) {
			m.deleteSessionLocked(hash)
		}
	}
	for ip, last := range m.lastRequestAt {
		if now.Sub(last) >= AccessRequestCooldown {
			delete(m.lastRequestAt, ip)
		}
	}
}

func (m *AccessManager) deleteSessionLocked(hash [32]byte) {
	if session, ok := m.sessions[hash]; ok {
		delete(m.sessionID, session.ID)
	}
	delete(m.sessions, hash)
}

func (m *AccessManager) pendingCountForIPLocked(ip string) int {
	count := 0
	for _, req := range m.pending {
		if req.IP == ip {
			count++
		}
	}
	return count
}

func NewAccessRequestToken() (string, error) {
	return accessRandomToken(32)
}

func accessRandomToken(bytes int) (string, error) {
	buf := make([]byte, bytes)
	if _, err := rand.Read(buf); err != nil {
		return "", err
	}
	return base64.RawURLEncoding.EncodeToString(buf), nil
}

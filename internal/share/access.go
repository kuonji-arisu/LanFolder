package share

import (
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"errors"
	"fmt"
	"math/big"
	"sort"
	"sync"
	"time"
)

const (
	AccessRequestTTL      = 2 * time.Minute
	AccessRequestCooldown = 5 * time.Second
	MaxAccessPendingCount = 32
)

var (
	ErrAccessRequestNotFound = errors.New("access_request_not_found")
	ErrAccessRequestLimited  = errors.New("access_request_limited")
)

type AccessRequest struct {
	ID           string    `json:"id"`
	Code         string    `json:"code"`
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
	ID        string    `json:"id"`
	IP        string    `json:"ip"`
	UserAgent string    `json:"userAgent"`
	CreatedAt time.Time `json:"createdAt"`
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
	pending       map[string]AccessRequest
	pendingByIP   map[string]string
	completed     map[string]completedAccessRequest
	sessions      map[[32]byte]AccessSession
	sessionID     map[string][32]byte
	lastRequestAt map[string]time.Time
}

func NewAccessManager() *AccessManager {
	return &AccessManager{
		now:           time.Now,
		pending:       map[string]AccessRequest{},
		pendingByIP:   map[string]string{},
		completed:     map[string]completedAccessRequest{},
		sessions:      map[[32]byte]AccessSession{},
		sessionID:     map[string][32]byte{},
		lastRequestAt: map[string]time.Time{},
	}
}

func (m *AccessManager) CreateRequest(ip, userAgent string) (AccessRequest, bool, error) {
	now := m.now()

	m.mu.Lock()
	defer m.mu.Unlock()
	m.pruneExpiredLocked(now)
	if id, ok := m.pendingByIP[ip]; ok {
		if req, ok := m.pending[id]; ok {
			req.UserAgent = userAgent
			req.RequestCount++
			req.LastSeenAt = now
			m.pending[id] = req
			return req, false, nil
		}
		delete(m.pendingByIP, ip)
	}
	if last, ok := m.lastRequestAt[ip]; ok && now.Sub(last) < AccessRequestCooldown {
		return AccessRequest{}, false, ErrAccessRequestLimited
	}
	if len(m.pending) >= MaxAccessPendingCount {
		return AccessRequest{}, false, ErrAccessRequestLimited
	}
	id, err := accessRandomToken(32)
	if err != nil {
		return AccessRequest{}, false, err
	}
	code, err := accessDisplayCode()
	if err != nil {
		return AccessRequest{}, false, err
	}
	req := AccessRequest{
		ID:           id,
		Code:         code,
		IP:           ip,
		UserAgent:    userAgent,
		RequestCount: 1,
		CreatedAt:    now,
		LastSeenAt:   now,
		ExpiresAt:    now.Add(AccessRequestTTL),
	}
	m.pending[req.ID] = req
	m.pendingByIP[req.IP] = req.ID
	m.lastRequestAt[req.IP] = now
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
	req, ok := m.pending[id]
	if !ok {
		return AccessRequest{}, ErrAccessRequestNotFound
	}
	delete(m.pending, id)
	delete(m.pendingByIP, req.IP)
	m.completed[id] = completedAccessRequest{
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
	req, ok := m.pending[id]
	if !ok {
		return AccessRequest{}, ErrAccessRequestNotFound
	}
	delete(m.pending, id)
	delete(m.pendingByIP, req.IP)
	m.lastRequestAt[req.IP] = now
	m.completed[id] = completedAccessRequest{
		state:     AccessPollDenied,
		request:   req,
		expiresAt: now.Add(AccessRequestTTL),
	}
	return req, nil
}

func (m *AccessManager) Poll(id string) (AccessPollResult, string, error) {
	now := m.now()
	m.mu.Lock()
	defer m.mu.Unlock()
	m.pruneExpiredLocked(now)
	if completed, ok := m.completed[id]; ok {
		if completed.state == AccessPollApproved {
			token, err := m.createSessionLocked(completed.request, now)
			if err != nil {
				return AccessPollResult{}, "", err
			}
			delete(m.completed, id)
			return AccessPollResult{State: completed.state}, token, nil
		}
		delete(m.completed, id)
		return AccessPollResult{State: completed.state}, "", nil
	}
	if _, ok := m.pending[id]; ok {
		return AccessPollResult{State: AccessPollPending}, "", nil
	}
	return AccessPollResult{State: AccessPollExpired}, "", nil
}

func (m *AccessManager) createSessionLocked(req AccessRequest, now time.Time) (string, error) {
	token, err := accessRandomToken(32)
	if err != nil {
		return "", err
	}
	sessionID, err := accessRandomToken(16)
	if err != nil {
		return "", err
	}
	hash := sha256.Sum256([]byte(token))
	m.sessions[hash] = AccessSession{
		ID:        sessionID,
		IP:        req.IP,
		UserAgent: req.UserAgent,
		CreatedAt: now,
	}
	m.sessionID[sessionID] = hash
	return token, nil
}

func (m *AccessManager) Validate(token string) bool {
	if token == "" {
		return false
	}
	hash := sha256.Sum256([]byte(token))
	m.mu.Lock()
	defer m.mu.Unlock()
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
	if session, ok := m.sessions[hash]; ok {
		delete(m.sessionID, session.ID)
	}
	delete(m.sessions, hash)
}

func (m *AccessManager) RevokeSession(id string) (AccessSession, bool) {
	m.mu.Lock()
	defer m.mu.Unlock()
	hash, ok := m.sessionID[id]
	if !ok {
		return AccessSession{}, false
	}
	session := m.sessions[hash]
	delete(m.sessionID, id)
	delete(m.sessions, hash)
	return session, true
}

func (m *AccessManager) Sessions() []AccessSession {
	m.mu.Lock()
	defer m.mu.Unlock()
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
	m.pending = map[string]AccessRequest{}
	m.pendingByIP = map[string]string{}
	m.completed = map[string]completedAccessRequest{}
	m.sessions = map[[32]byte]AccessSession{}
	m.sessionID = map[string][32]byte{}
	m.lastRequestAt = map[string]time.Time{}
}

func (m *AccessManager) pruneExpiredLocked(now time.Time) {
	for id, req := range m.pending {
		if !now.Before(req.ExpiresAt) {
			delete(m.pending, id)
			delete(m.pendingByIP, req.IP)
			m.completed[id] = completedAccessRequest{state: AccessPollExpired, expiresAt: now.Add(AccessRequestTTL)}
		}
	}
	for id, completed := range m.completed {
		if !now.Before(completed.expiresAt) {
			delete(m.completed, id)
		}
	}
	for ip, last := range m.lastRequestAt {
		if now.Sub(last) >= AccessRequestCooldown {
			delete(m.lastRequestAt, ip)
		}
	}
}

func accessRandomToken(bytes int) (string, error) {
	buf := make([]byte, bytes)
	if _, err := rand.Read(buf); err != nil {
		return "", err
	}
	return base64.RawURLEncoding.EncodeToString(buf), nil
}

func accessDisplayCode() (string, error) {
	n, err := rand.Int(rand.Reader, big.NewInt(1000000))
	if err != nil {
		return "", err
	}
	return fmt.Sprintf("%06d", n.Int64()), nil
}

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

const AccessRequestTTL = 2 * time.Minute

var (
	ErrAccessRequestNotFound = errors.New("access_request_not_found")
	ErrAccessRequestExpired  = errors.New("access_request_expired")
)

type AccessRequest struct {
	ID        string    `json:"id"`
	Code      string    `json:"code"`
	IP        string    `json:"ip"`
	UserAgent string    `json:"userAgent"`
	CreatedAt time.Time `json:"createdAt"`
	ExpiresAt time.Time `json:"expiresAt"`
}

type AccessSession struct {
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
	token     string
	expiresAt time.Time
}

type AccessManager struct {
	mu        sync.Mutex
	now       func() time.Time
	pending   map[string]AccessRequest
	completed map[string]completedAccessRequest
	sessions  map[[32]byte]AccessSession
	sessionID map[string][32]byte
}

func NewAccessManager() *AccessManager {
	return &AccessManager{
		now:       time.Now,
		pending:   map[string]AccessRequest{},
		completed: map[string]completedAccessRequest{},
		sessions:  map[[32]byte]AccessSession{},
		sessionID: map[string][32]byte{},
	}
}

func (m *AccessManager) CreateRequest(ip, userAgent string) (AccessRequest, error) {
	id, err := accessRandomToken(32)
	if err != nil {
		return AccessRequest{}, err
	}
	code, err := accessDisplayCode()
	if err != nil {
		return AccessRequest{}, err
	}
	now := m.now()
	req := AccessRequest{
		ID:        id,
		Code:      code,
		IP:        ip,
		UserAgent: userAgent,
		CreatedAt: now,
		ExpiresAt: now.Add(AccessRequestTTL),
	}

	m.mu.Lock()
	defer m.mu.Unlock()
	m.pruneExpiredLocked(now)
	m.pending[req.ID] = req
	return req, nil
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

func (m *AccessManager) Approve(id string) error {
	now := m.now()
	m.mu.Lock()
	defer m.mu.Unlock()
	m.pruneExpiredLocked(now)
	req, ok := m.pending[id]
	if !ok {
		return ErrAccessRequestNotFound
	}
	if !now.Before(req.ExpiresAt) {
		delete(m.pending, id)
		m.completed[id] = completedAccessRequest{state: AccessPollExpired, expiresAt: now.Add(AccessRequestTTL)}
		return ErrAccessRequestExpired
	}
	token, err := accessRandomToken(32)
	if err != nil {
		return err
	}
	sessionID, err := accessRandomToken(16)
	if err != nil {
		return err
	}
	hash := sha256.Sum256([]byte(token))
	m.sessions[hash] = AccessSession{
		ID:        sessionID,
		IP:        req.IP,
		UserAgent: req.UserAgent,
		CreatedAt: now,
	}
	m.sessionID[sessionID] = hash
	delete(m.pending, id)
	m.completed[id] = completedAccessRequest{state: AccessPollApproved, token: token, expiresAt: now.Add(AccessRequestTTL)}
	return nil
}

func (m *AccessManager) Deny(id string) error {
	now := m.now()
	m.mu.Lock()
	defer m.mu.Unlock()
	m.pruneExpiredLocked(now)
	if _, ok := m.pending[id]; !ok {
		return ErrAccessRequestNotFound
	}
	delete(m.pending, id)
	m.completed[id] = completedAccessRequest{state: AccessPollDenied, expiresAt: now.Add(AccessRequestTTL)}
	return nil
}

func (m *AccessManager) Poll(id string) (AccessPollResult, string) {
	now := m.now()
	m.mu.Lock()
	defer m.mu.Unlock()
	m.pruneExpiredLocked(now)
	if completed, ok := m.completed[id]; ok {
		delete(m.completed, id)
		return AccessPollResult{State: completed.state}, completed.token
	}
	if _, ok := m.pending[id]; ok {
		return AccessPollResult{State: AccessPollPending}, ""
	}
	return AccessPollResult{State: AccessPollExpired}, ""
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

func (m *AccessManager) RevokeSession(id string) bool {
	m.mu.Lock()
	defer m.mu.Unlock()
	hash, ok := m.sessionID[id]
	if !ok {
		return false
	}
	delete(m.sessionID, id)
	delete(m.sessions, hash)
	return true
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
	m.completed = map[string]completedAccessRequest{}
	m.sessions = map[[32]byte]AccessSession{}
	m.sessionID = map[string][32]byte{}
}

func (m *AccessManager) pruneExpiredLocked(now time.Time) {
	for id, req := range m.pending {
		if !now.Before(req.ExpiresAt) {
			delete(m.pending, id)
			m.completed[id] = completedAccessRequest{state: AccessPollExpired, expiresAt: now.Add(AccessRequestTTL)}
		}
	}
	for id, completed := range m.completed {
		if !now.Before(completed.expiresAt) {
			delete(m.completed, id)
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

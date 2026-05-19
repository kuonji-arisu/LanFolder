package share

import (
	"bufio"
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"
)

const (
	messagesFileName    = "messages.jsonl"
	HostClientID        = "host"
	MaxMessageTextChars = 2000
	maxMessages         = 120
	maxClientIDChars    = 128
)

type MessageStore struct {
	mu sync.Mutex
}

func NewMessageStore() *MessageStore {
	return &MessageStore{}
}

func (s *MessageStore) List(root string) ([]Message, error) {
	root, err := cleanMessageRoot(root)
	if err != nil {
		return nil, err
	}

	s.mu.Lock()
	defer s.mu.Unlock()

	messages, err := readMessagesFile(root)
	if err != nil {
		return nil, err
	}
	return messages, nil
}

func (s *MessageStore) Send(root, clientID, text string) (Message, error) {
	root, err := cleanMessageRoot(root)
	if err != nil {
		return Message{}, err
	}
	clientID = strings.TrimSpace(clientID)
	text = strings.TrimSpace(text)
	if clientID == "" || len([]rune(clientID)) > maxClientIDChars || text == "" || len([]rune(text)) > MaxMessageTextChars {
		return Message{}, ErrInvalidMessage
	}

	id, err := newMessageID()
	if err != nil {
		return Message{}, err
	}
	message := Message{
		ID:        id,
		CreatedAt: time.Now().UTC(),
		ClientID:  clientID,
		Text:      text,
	}

	s.mu.Lock()
	defer s.mu.Unlock()

	dir := filepath.Join(root, managedDirName)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return Message{}, err
	}
	messages, err := readMessagesFile(root)
	if err != nil {
		return Message{}, err
	}
	messages = append(messages, message)
	if len(messages) > maxMessages {
		messages = messages[len(messages)-maxMessages:]
	}
	if err := writeMessagesFile(root, messages); err != nil {
		return Message{}, err
	}
	return message, nil
}

func readMessagesFile(root string) ([]Message, error) {
	file, err := os.Open(messagesPath(root))
	if errors.Is(err, os.ErrNotExist) {
		return []Message{}, nil
	}
	if err != nil {
		return nil, err
	}
	defer file.Close()

	messages := []Message{}
	scanner := bufio.NewScanner(file)
	scanner.Buffer(make([]byte, 0, 64*1024), 1024*1024)
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line == "" {
			continue
		}
		var message Message
		if err := json.Unmarshal([]byte(line), &message); err != nil {
			continue
		}
		messages = append(messages, message)
	}
	if err := scanner.Err(); err != nil {
		return nil, err
	}
	if len(messages) > maxMessages {
		messages = messages[len(messages)-maxMessages:]
	}
	return messages, nil
}

func writeMessagesFile(root string, messages []Message) error {
	var builder strings.Builder
	encoder := json.NewEncoder(&builder)
	for _, message := range messages {
		if err := encoder.Encode(message); err != nil {
			return err
		}
	}
	return writeMessageFileAtomic(messagesPath(root), []byte(builder.String()), 0644)
}

func writeMessageFileAtomic(path string, data []byte, perm os.FileMode) error {
	dir := filepath.Dir(path)
	file, err := os.CreateTemp(dir, "messages-*.tmp")
	if err != nil {
		return err
	}
	tmp := file.Name()
	cleanup := true
	closed := false
	closeFile := func() error {
		if closed {
			return nil
		}
		closed = true
		return file.Close()
	}
	defer func() {
		if cleanup {
			_ = closeFile()
			_ = os.Remove(tmp)
		}
	}()

	if err := file.Chmod(perm); err != nil {
		return err
	}
	if _, err := file.Write(data); err != nil {
		return err
	}
	if err := file.Sync(); err != nil {
		return err
	}
	if err := closeFile(); err != nil {
		return err
	}
	if err := os.Rename(tmp, path); err != nil {
		return err
	}
	cleanup = false
	return nil
}

func (s *MessageStore) Clear(root string) error {
	root, err := cleanMessageRoot(root)
	if err != nil {
		return err
	}

	s.mu.Lock()
	defer s.mu.Unlock()

	if err := os.Remove(messagesPath(root)); err != nil && !errors.Is(err, os.ErrNotExist) {
		return err
	}
	return nil
}

func (m *Manager) ListMessages() ([]Message, error) {
	root, err := m.messageRoot()
	if err != nil {
		return nil, err
	}
	return m.messages.List(root)
}

func (m *Manager) SendMessage(clientID, text string) (Message, error) {
	root, permission, err := m.messageRootAndPermission()
	if err != nil {
		return Message{}, err
	}
	if !permission.CanUpload() {
		return Message{}, ErrPermissionDenied
	}
	if strings.TrimSpace(clientID) == HostClientID {
		return Message{}, ErrInvalidMessage
	}
	return m.messages.Send(root, clientID, text)
}

func (m *Manager) SendHostMessage(text string) (Message, error) {
	root, err := m.messageRoot()
	if err != nil {
		return Message{}, err
	}
	return m.messages.Send(root, HostClientID, text)
}

func (m *Manager) ClearMessages() error {
	root, permission, err := m.messageRootAndPermission()
	if err != nil {
		return err
	}
	if !permission.CanDelete() {
		return ErrPermissionDenied
	}
	return m.messages.Clear(root)
}

func (m *Manager) ClearHostMessages() error {
	root, err := m.messageRoot()
	if err != nil {
		return err
	}
	return m.messages.Clear(root)
}

func (m *Manager) ResetMessages() error {
	root, err := m.messageRoot()
	if err != nil {
		return err
	}
	return m.messages.Clear(root)
}

func (m *Manager) messageRoot() (string, error) {
	root, _, err := m.messageRootAndPermission()
	return root, err
}

func (m *Manager) messageRootAndPermission() (string, Permission, error) {
	m.mu.RLock()
	root := m.root
	permission := m.permission
	m.mu.RUnlock()
	root, err := cleanMessageRoot(root)
	return root, permission, err
}

func cleanMessageRoot(root string) (string, error) {
	if root == "" {
		return "", ErrInvalidRoot
	}
	abs, err := filepath.Abs(root)
	if err != nil {
		return "", err
	}
	info, err := os.Stat(abs)
	if err != nil {
		return "", err
	}
	if !info.IsDir() {
		return "", ErrInvalidRoot
	}
	return abs, nil
}

func messagesPath(root string) string {
	return filepath.Join(root, managedDirName, messagesFileName)
}

func newMessageID() (string, error) {
	var b [16]byte
	if _, err := rand.Read(b[:]); err != nil {
		return "", err
	}
	return fmt.Sprintf("%d-%s", time.Now().UTC().UnixNano(), hex.EncodeToString(b[:])), nil
}

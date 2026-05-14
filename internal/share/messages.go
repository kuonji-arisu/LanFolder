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
	MaxMessageTextChars = 2000
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
	file, err := os.OpenFile(messagesPath(root), os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0644)
	if err != nil {
		return Message{}, err
	}
	defer file.Close()

	data, err := json.Marshal(message)
	if err != nil {
		return Message{}, err
	}
	if _, err := file.Write(append(data, '\n')); err != nil {
		return Message{}, err
	}
	return message, nil
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
	root, err := m.messageRoot()
	if err != nil {
		return Message{}, err
	}
	return m.messages.Send(root, clientID, text)
}

func (m *Manager) ClearMessages() error {
	root, err := m.messageRoot()
	if err != nil {
		return err
	}
	return m.messages.Clear(root)
}

func (m *Manager) messageRoot() (string, error) {
	m.mu.RLock()
	root := m.root
	m.mu.RUnlock()
	return cleanMessageRoot(root)
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

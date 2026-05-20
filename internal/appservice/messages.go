package appservice

import "lanfolder/internal/share"

func (s *AppService) Messages() ([]share.Message, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	if !s.server.Running() {
		return nil, newCommandError(errShareNotRunning, nil)
	}
	return s.server.HostMessages()
}

func (s *AppService) SendMessage(text string) (share.Message, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	if !s.server.Running() {
		return share.Message{}, newCommandError(errShareNotRunning, nil)
	}
	return s.server.SendHostMessage(text)
}

func (s *AppService) ClearMessages() error {
	s.mu.Lock()
	defer s.mu.Unlock()

	if !s.server.Running() {
		return newCommandError(errShareNotRunning, nil)
	}
	return s.server.ClearHostMessages()
}

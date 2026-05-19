package appservice

import (
	"strconv"
	"sync"
	"time"

	"github.com/wailsapp/wails/v3/pkg/application"

	"lanfolder/internal/desktop"
	"lanfolder/internal/notice"
)

const maxQueuedNotices = 50

type NoticeCenter struct {
	mu       sync.Mutex
	app      *application.App
	notifier *notice.RuntimeService
	notices  []desktop.Notice
	seq      uint64
	drained  bool
}

func NewNoticeCenter(notifier *notice.RuntimeService) *NoticeCenter {
	return &NoticeCenter{notifier: notifier}
}

func (c *NoticeCenter) SetApp(app *application.App) {
	c.mu.Lock()
	c.app = app
	c.mu.Unlock()
}

func (c *NoticeCenter) Add(level desktop.NoticeLevel, source desktop.NoticeSource, err error, message string) {
	noticeItem := desktop.Notice{
		Level:     level,
		Source:    source,
		Message:   message,
		CreatedAt: time.Now(),
	}
	if payload := commandErrorPayload(err); payload != nil {
		noticeItem.Error = payload
	}

	c.mu.Lock()
	c.seq++
	noticeItem.ID = strconv.FormatUint(c.seq, 10)
	if !c.drained {
		c.notices = append(c.notices, noticeItem)
		if len(c.notices) > maxQueuedNotices {
			c.notices = c.notices[len(c.notices)-maxQueuedNotices:]
		}
	}
	app := c.app
	c.mu.Unlock()

	if app != nil {
		app.Event.Emit("app:notice", noticeItem)
	}
}

func (c *NoticeCenter) Drain() []desktop.Notice {
	c.mu.Lock()
	defer c.mu.Unlock()
	out := make([]desktop.Notice, len(c.notices))
	copy(out, c.notices)
	c.notices = nil
	c.drained = true
	return out
}

func (c *NoticeCenter) Present(window notice.Window, noticeItem desktop.Notice, message string, language string) string {
	c.mu.Lock()
	notifier := c.notifier
	c.mu.Unlock()
	return notice.Present(window, notifier, noticeItem, message, language)
}

func (s *AppService) DrainNotices() []desktop.Notice {
	return s.noticeCenter().Drain()
}

func (s *AppService) PresentNotice(noticeItem desktop.Notice, message string) string {
	s.mu.Lock()
	window := s.window
	language := s.config.Language
	notices := s.noticeCenterLocked()
	s.mu.Unlock()
	return notices.Present(window, noticeItem, message, language)
}

func (s *AppService) addNotice(level desktop.NoticeLevel, source desktop.NoticeSource, err error, message string) {
	s.noticeCenter().Add(level, source, err, message)
}

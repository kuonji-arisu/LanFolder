package server

import (
	"context"
	"errors"
	"fmt"
	"io/fs"
	"log"
	"net"
	"net/http"
	"strconv"
	"sync"
	"time"

	"lanfolder/internal/i18n"
	"lanfolder/internal/share"
)

var errServerAlreadyRunning = errors.New("server is already running")

type Server struct {
	mu              sync.RWMutex
	manager         *share.Manager
	access          *share.AccessManager
	invalidPolls    *fixedWindowLimiter
	staticFS        fs.FS
	httpSrv         *http.Server
	config          Config
	logs            []LogEntry
	maxLogs         int
	onAccessRequest func()
}

func New(staticFS fs.FS) *Server {
	return &Server{
		manager:      share.NewManager(),
		access:       share.NewAccessManager(),
		invalidPolls: newFixedWindowLimiter(time.Minute, 30),
		staticFS:     staticFS,
		config: Config{
			Host:                  "0.0.0.0",
			Port:                  8899,
			Permission:            share.PermissionReadOnly,
			AccessSessionLifetime: share.AccessSessionNever,
			Language:              i18n.Chinese,
		},
		maxLogs: 120,
	}
}

func (s *Server) Start(cfg Config) error {
	if cfg.Host == "" {
		cfg.Host = "0.0.0.0"
	}
	if cfg.Port <= 0 || cfg.Port > 65535 {
		return fmt.Errorf("invalid port: %d", cfg.Port)
	}
	if !cfg.Permission.Valid() {
		cfg.Permission = share.PermissionReadOnly
	}
	cfg.AccessSessionLifetime = share.NormalizeAccessSessionLifetime(cfg.AccessSessionLifetime)
	cfg.Language = i18n.NormalizeLanguage(cfg.Language)

	s.mu.Lock()
	if s.httpSrv != nil {
		s.mu.Unlock()
		return errServerAlreadyRunning
	}
	if err := s.manager.Configure(cfg.Root, cfg.Permission, cfg.ShowHidden); err != nil {
		s.mu.Unlock()
		return err
	}
	s.access.SetSessionLifetime(cfg.AccessSessionLifetime)
	mux := http.NewServeMux()
	s.routes(mux)
	addr := net.JoinHostPort(cfg.Host, strconv.Itoa(cfg.Port))
	httpSrv := &http.Server{
		Addr:              addr,
		Handler:           s.logMiddleware(s.secureMiddleware(s.accessMiddleware(mux))),
		ReadHeaderTimeout: 10 * time.Second,
		ReadTimeout:       15 * time.Minute,
		WriteTimeout:      15 * time.Minute,
		IdleTimeout:       60 * time.Second,
	}
	ln, err := net.Listen("tcp", addr)
	if err != nil {
		s.mu.Unlock()
		return err
	}
	if err := s.manager.ClearMessages(); err != nil {
		_ = ln.Close()
		s.mu.Unlock()
		return err
	}
	s.access.Clear()
	s.invalidPolls.Clear()
	s.httpSrv = httpSrv
	s.config = cfg
	s.mu.Unlock()

	go func() {
		if err := httpSrv.Serve(ln); err != nil && !errors.Is(err, http.ErrServerClosed) {
			log.Printf("lan server stopped: %v", err)
		}
		s.mu.Lock()
		if s.httpSrv == httpSrv {
			s.httpSrv = nil
		}
		s.mu.Unlock()
	}()
	return nil
}

func (s *Server) Stop(ctx context.Context) error {
	s.mu.Lock()
	srv := s.httpSrv
	s.httpSrv = nil
	s.mu.Unlock()
	if srv == nil {
		return nil
	}
	s.access.Clear()
	s.invalidPolls.Clear()
	return errors.Join(srv.Shutdown(ctx), s.manager.ClearMessages())
}

func (s *Server) Status() RuntimeStatus {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return RuntimeStatus{
		Running:               s.httpSrv != nil,
		Host:                  s.config.Host,
		Port:                  s.config.Port,
		Root:                  s.config.Root,
		Permission:            s.config.Permission,
		AccessApproval:        s.config.AccessApproval,
		AccessSessionLifetime: s.config.AccessSessionLifetime,
		Language:              s.config.Language,
	}
}

func (s *Server) Logs() []LogEntry {
	s.mu.RLock()
	defer s.mu.RUnlock()
	out := make([]LogEntry, len(s.logs))
	copy(out, s.logs)
	return out
}

func (s *Server) SetAccessRequestCallback(fn func()) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.onAccessRequest = fn
}

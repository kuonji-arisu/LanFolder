package appservice

import (
	"fmt"
	"slices"
	"strings"
	"sync"
	"time"

	"lanfolder/internal/config"
	"lanfolder/internal/platform"
)

const addressWatchInterval = 10 * time.Second

type AddressService struct {
	mu      sync.Mutex
	lanIPs  func() []string
	started bool
}

func NewAddressService(lanIPs func() []string) *AddressService {
	return &AddressService{lanIPs: lanIPs}
}

func (a *AddressService) Addresses(cfg config.Config) []string {
	ips := a.currentLANIPs()
	if len(ips) == 0 {
		ips = []string{"127.0.0.1"}
	}
	addrs := make([]string, 0, len(ips))
	for _, ip := range ips {
		addrs = append(addrs, fmt.Sprintf("http://%s:%d", ip, cfg.Port))
	}
	return addrs
}

func (a *AddressService) StartWatcher(getConfig func() config.Config, emitStateChanged func(string)) {
	a.mu.Lock()
	if a.started {
		a.mu.Unlock()
		return
	}
	a.started = true
	a.mu.Unlock()

	last := addressListKey(a.Addresses(getConfig()))
	go func() {
		ticker := time.NewTicker(addressWatchInterval)
		defer ticker.Stop()
		for range ticker.C {
			next := addressListKey(a.Addresses(getConfig()))
			if next == last {
				continue
			}
			last = next
			emitStateChanged("addresses")
		}
	}()
}

func (a *AddressService) currentLANIPs() []string {
	a.mu.Lock()
	lanIPs := a.lanIPs
	a.mu.Unlock()
	if lanIPs != nil {
		return lanIPs()
	}
	return platform.LANIPs()
}

func (s *AppService) addressesForConfig(cfg config.Config) []string {
	if s.addressSvc != nil {
		return s.addressSvc.Addresses(cfg)
	}
	return NewAddressService(s.lanIPs).Addresses(cfg)
}

func (s *AppService) addresses(cfg config.Config) []string {
	return s.addressesForConfig(cfg)
}

func (s *AppService) currentLANIPs() []string {
	if s.addressSvc != nil {
		return s.addressSvc.currentLANIPs()
	}
	return NewAddressService(s.lanIPs).currentLANIPs()
}

func (s *AppService) startAddressWatcher() {
	addresses := s.addressService()
	addresses.StartWatcher(
		func() config.Config {
			s.mu.Lock()
			defer s.mu.Unlock()
			return s.config
		},
		func(reason string) {
			s.emitStateChanged(reason)
		},
	)
}

func addressListKey(addresses []string) string {
	sorted := slices.Clone(addresses)
	slices.Sort(sorted)
	return strings.Join(sorted, "\n")
}

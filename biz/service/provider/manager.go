package provider

import (
	"fmt"
	"sync"

	"github.com/yi-nology/git-manage-service/biz/dal/db"
	"github.com/yi-nology/git-manage-service/biz/model/po"
)

var (
	instance *ProviderManager
	once     sync.Once
)

type ProviderManager struct {
	mu    sync.RWMutex
	cache map[uint]Provider
}

func GetManager() *ProviderManager {
	once.Do(func() {
		instance = &ProviderManager{
			cache: make(map[uint]Provider),
		}
	})
	return instance
}

func (m *ProviderManager) GetProvider(configID uint) (Provider, error) {
	m.mu.RLock()
	if p, ok := m.cache[configID]; ok {
		m.mu.RUnlock()
		return p, nil
	}
	m.mu.RUnlock()

	m.mu.Lock()
	defer m.mu.Unlock()

	dao := db.NewProviderConfigDAO()
	cfg, err := dao.FindByID(configID)
	if err != nil {
		return nil, fmt.Errorf("provider config not found: %w", err)
	}

	cred, err := resolveCredential(cfg.CredentialID)
	if err != nil {
		return nil, fmt.Errorf("credential not found: %w", err)
	}

	p, err := newProvider(cfg, cred)
	if err != nil {
		return nil, err
	}
	m.cache[configID] = p
	return p, nil
}

func (m *ProviderManager) Invalidate(configID uint) {
	m.mu.Lock()
	defer m.mu.Unlock()
	delete(m.cache, configID)
}

func (m *ProviderManager) DetectAndCreate(remoteURL string, credentialID uint) (Provider, *DetectResult, error) {
	result, err := DetectPlatform(remoteURL)
	if err != nil {
		return nil, nil, err
	}

	cred, err := resolveCredential(credentialID)
	if err != nil {
		return nil, nil, err
	}

	cfg := &po.ProviderConfig{
		Platform:     string(result.Platform),
		BaseURL:      result.BaseURL,
		CredentialID: credentialID,
	}

	p, err := newProvider(cfg, cred)
	if err != nil {
		return nil, nil, err
	}
	return p, result, nil
}

func resolveCredential(credentialID uint) (*po.Credential, error) {
	if credentialID == 0 {
		return nil, fmt.Errorf("credential ID is 0")
	}
	dao := db.NewCredentialDAO()
	cred, err := dao.FindByID(credentialID)
	if err != nil {
		return nil, fmt.Errorf("credential %d not found: %w", credentialID, err)
	}
	return cred, nil
}

func newProvider(cfg *po.ProviderConfig, cred *po.Credential) (Provider, error) {
	token := cred.Secret
	switch Platform(cfg.Platform) {
	case PlatformGitLab:
		return NewGitLabProvider(cfg.BaseURL, token), nil
	case PlatformGitHub:
		return NewGitHubProvider(cfg.BaseURL, token), nil
	case PlatformGitea:
		return NewGiteaProvider(cfg.BaseURL, token), nil
	default:
		return nil, fmt.Errorf("unsupported platform: %s", cfg.Platform)
	}
}

package service

import (
	"context"
	"sync"

	"github.com/kavos113/quickctf/ctf-manager/domain"
)

type mockInstanceRepository struct {
	instances map[string]*domain.Instance
	mu        sync.RWMutex
}

func newMockInstanceRepository() *mockInstanceRepository {
	return &mockInstanceRepository{
		instances: make(map[string]*domain.Instance),
	}
}

func (m *mockInstanceRepository) Create(ctx context.Context, instance *domain.Instance) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	if _, exists := m.instances[instance.InstanceID]; exists {
		return domain.ErrInstanceAlreadyExists
	}

	m.instances[instance.InstanceID] = instance
	return nil
}

func (m *mockInstanceRepository) FindByID(ctx context.Context, instanceID string) (*domain.Instance, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	instance, exists := m.instances[instanceID]
	if !exists {
		return nil, domain.ErrInstanceNotFound
	}

	return instance, nil
}

func (m *mockInstanceRepository) Update(ctx context.Context, instance *domain.Instance) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	if _, exists := m.instances[instance.InstanceID]; !exists {
		return domain.ErrInstanceNotFound
	}

	m.instances[instance.InstanceID] = instance
	return nil
}

func (m *mockInstanceRepository) Delete(ctx context.Context, instanceID string) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	if _, exists := m.instances[instanceID]; !exists {
		return domain.ErrInstanceNotFound
	}

	delete(m.instances, instanceID)
	return nil
}

func (m *mockInstanceRepository) FindAll(ctx context.Context) ([]*domain.Instance, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	instances := make([]*domain.Instance, 0, len(m.instances))
	for _, instance := range m.instances {
		instances = append(instances, instance)
	}

	return instances, nil
}

func (m *mockInstanceRepository) FindByRunnerURL(ctx context.Context, runnerURL string) ([]*domain.Instance, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	instances := make([]*domain.Instance, 0)
	for _, instance := range m.instances {
		if instance.RunnerURL == runnerURL {
			instances = append(instances, instance)
		}
	}

	return instances, nil
}

func (m *mockInstanceRepository) FindExpired(ctx context.Context) ([]*domain.Instance, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	instances := make([]*domain.Instance, 0)
	for _, instance := range m.instances {
		if instance.IsExpired() {
			instances = append(instances, instance)
		}
	}

	return instances, nil
}

package registry

import (
	"reflect"
	"sync"
)

type ServiceRegistry struct {
	mu       sync.RWMutex
	services map[reflect.Type]any
}

func NewServiceRegistry() *ServiceRegistry {
	return &ServiceRegistry{
		services: make(map[reflect.Type]any),
	}
}

func (sr *ServiceRegistry) Add(dep any) {
	sr.mu.Lock()
	defer sr.mu.Unlock()

	rtype := reflect.TypeOf(dep)
	sr.services[rtype] = dep
}

func (sr *ServiceRegistry) Get(t reflect.Type) any {
	sr.mu.RLock()
	defer sr.mu.RUnlock()
	return sr.services[t]

}

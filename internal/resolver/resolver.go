package resolver

import (
	"sync"

	"gitlab.com/ludovic-alarcon/azabox/internal/dto"
)

var (
	once     sync.Once
	registry *RegistryResolver
)

type Resolver interface {
	Resolve(dto.BinaryInfo) (string, error)
}

type RegistryResolver struct {
	mutex     sync.RWMutex
	resolvers []Resolver
}

func GetRegistryResolver() *RegistryResolver {
	once.Do(func() {
		registry = newRegistryResolver()
	})
	return registry
}

func newRegistryResolver() *RegistryResolver {
	return &RegistryResolver{
		resolvers: make([]Resolver, 0, 1),
	}
}

func (r *RegistryResolver) WithDefaultResolvers() *RegistryResolver {
	r.Register(&GithubResolver{})
	return r
}

func (r *RegistryResolver) Register(resolver Resolver) {
	r.mutex.Lock()
	defer r.mutex.Unlock()
	r.resolvers = append(r.resolvers, resolver)
}

func (r *RegistryResolver) GetResolvers() []Resolver {
	r.mutex.RLock()
	defer r.mutex.RUnlock()
	return append([]Resolver{}, r.resolvers...)
}

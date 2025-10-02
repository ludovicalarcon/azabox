package resolver

import (
	"sync"

	"gitlab.com/ludovic-alarcon/azabox/internal/dto"
	"gitlab.com/ludovic-alarcon/azabox/internal/types"
)

var (
	once     sync.Once
	registry *RegistryResolver
)

type Resolver interface {
	Resolve(*dto.BinaryInfo) (string, error)
}

type RegistryResolver struct {
	mutex     sync.RWMutex
	resolvers types.Set[Resolver]
}

func GetRegistryResolver() *RegistryResolver {
	once.Do(func() {
		registry = newRegistryResolver()
	})
	return registry
}

func newRegistryResolver() *RegistryResolver {
	return &RegistryResolver{
		resolvers: types.NewSet[Resolver](1),
	}
}

func (r *RegistryResolver) WithDefaultResolvers() *RegistryResolver {
	r.Register(NewGithubResolver(GHBaseAPIUrl))
	return r
}

func (r *RegistryResolver) Register(resolver Resolver) {
	r.mutex.Lock()
	defer r.mutex.Unlock()
	r.resolvers.Add(resolver)
}

func (r *RegistryResolver) Unregister(resolver Resolver) {
	r.mutex.Lock()
	defer r.mutex.Unlock()
	r.resolvers.Remove(resolver)
}

func (r *RegistryResolver) GetResolvers() types.Set[Resolver] {
	r.mutex.RLock()
	defer r.mutex.RUnlock()
	return r.resolvers
}

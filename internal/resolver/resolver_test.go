package resolver

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gitlab.com/ludovic-alarcon/azabox/internal/dto"
)

type DummyResolver struct{}

func (d *DummyResolver) Resolve(dto.BinaryInfo) (string, error) {
	return "", nil
}

func TestRegistry(t *testing.T) {
	t.Run("should create a registry", func(t *testing.T) {
		registry := newRegistryResolver()
		require.NotNil(t, registry)
		assert.Len(t, registry.resolvers, 0)
	})

	t.Run("should register and return resolver list", func(t *testing.T) {
		registry := newRegistryResolver()
		dummyResolver := &DummyResolver{}

		registry.Register(dummyResolver)

		resolvers := registry.GetResolvers()
		require.Len(t, resolvers, 1)
		assert.Equal(t, dummyResolver, resolvers.ToSlice()[0])
	})

	t.Run("should not register duplicate", func(t *testing.T) {
		registry := newRegistryResolver()
		dummyResolver := &DummyResolver{}

		registry.Register(dummyResolver)
		registry.Register(dummyResolver)
		registry.Register(dummyResolver)

		resolvers := registry.GetResolvers()
		require.Len(t, resolvers, 1)
		assert.Equal(t, dummyResolver, resolvers.ToSlice()[0])
	})
}

func TestGetRegistry(t *testing.T) {
	t.Run("should return a single instance", func(t *testing.T) {
		registry := GetRegistryResolver()

		require.NotNil(t, registry)
		for i := 0; i < 100; i++ {
			registryNew := GetRegistryResolver()
			// Should have the same addr
			assert.True(t, registry == registryNew)
		}

		registry.Register(&DummyResolver{})
		resolvers := GetRegistryResolver().GetResolvers()

		require.Len(t, resolvers, 1)
		assert.IsType(t, &DummyResolver{}, resolvers.ToSlice()[0])
	})
}

func TestWithDefaultResolvers(t *testing.T) {
	t.Run("should handle default resolvers", func(t *testing.T) {
		registry := newRegistryResolver().WithDefaultResolvers()
		resolvers := registry.GetResolvers()

		require.Len(t, resolvers, 1)
		assert.IsType(t, &GithubResolver{}, resolvers.ToSlice()[0])
	})
}

package vulkan

import (
	"fmt"
	"os"
	"path/filepath"
	"sync"

	"github.com/bnema/purego-vulkan/internal/capi"
	internalloader "github.com/bnema/purego-vulkan/internal/loader"
)

type config struct {
	LibraryPath string
}

// Option configures process-level Vulkan initialization.
type Option func(*config)

// WithLibraryPath makes Init try path before the default Vulkan loader names.
func WithLibraryPath(path string) Option {
	return func(c *config) { c.LibraryPath = path }
}

type runtimeHooks struct {
	open     func([]string, internalloader.OpenFunc) (internalloader.SharedLibrary, error)
	lookup   func(internalloader.SharedLibrary, string, internalloader.LookupFunc) (uintptr, error)
	register func(any, uintptr)
}

var defaultRuntimeHooks = runtimeHooks{
	open:     internalloader.Open,
	lookup:   internalloader.Lookup,
	register: capi.RegisterFunc,
}

var (
	initOnce sync.Once
	initErr  error

	initializedLibrary internalloader.SharedLibrary
	hooks              = defaultRuntimeHooks
)

// Init opens the Vulkan loader once and resolves vkGetInstanceProcAddr.
func Init(opts ...Option) error {
	cfg := config{}
	for _, opt := range opts {
		opt(&cfg)
	}

	initOnce.Do(func() {
		initErr = initRuntime(cfg)
	})
	return initErr
}

func initRuntime(cfg config) error {
	h := hooks
	if h.open == nil {
		h.open = internalloader.Open
	}
	if h.lookup == nil {
		h.lookup = internalloader.Lookup
	}
	if h.register == nil {
		h.register = capi.RegisterFunc
	}

	lib, err := h.open(vulkanLoaderCandidates(cfg.LibraryPath), internalloader.PuregoOpen)
	if err != nil {
		return err
	}

	addr, err := h.lookup(lib, "vkGetInstanceProcAddr", internalloader.PuregoLookup)
	if err != nil {
		return fmt.Errorf("vulkan: resolve vkGetInstanceProcAddr: %w", err)
	}
	h.register(&vkGetInstanceProcAddr, addr)
	initializedLibrary = lib
	return nil
}

func vulkanLoaderCandidates(explicit string) []string {
	var candidates []string
	if explicit != "" {
		candidates = append(candidates, explicit)
	}
	if sdk := os.Getenv("VULKAN_SDK"); sdk != "" {
		candidates = append(candidates,
			filepath.Join(sdk, "lib", "libvulkan.so.1"),
			filepath.Join(sdk, "lib", "libvulkan.so"),
		)
	}
	candidates = append(candidates, "libvulkan.so.1", "libvulkan.so")
	return candidates
}

func setRuntimeHooksForTest(h runtimeHooks) {
	hooks = h
}

func resetInitForTest() {
	initOnce = sync.Once{}
	initErr = nil
	initializedLibrary = internalloader.SharedLibrary{}
	vkGetInstanceProcAddr = nil
	hooks = defaultRuntimeHooks
}

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
	close    func(internalloader.SharedLibrary, internalloader.CloseFunc) error
	register func(any, uintptr)
}

var defaultRuntimeHooks = runtimeHooks{
	open:     internalloader.Open,
	lookup:   internalloader.Lookup,
	close:    internalloader.Close,
	register: capi.RegisterFunc,
}

var (
	initMu   sync.Mutex
	initDone bool
	initErr  error

	hooks = defaultRuntimeHooks
)

// Init opens the Vulkan loader and resolves vkGetInstanceProcAddr.
//
// A successful initialization is process-global and runs once. Failed attempts
// are not cached so callers can retry with different options, such as an
// explicit loader path discovered after startup.
func Init(opts ...Option) error {
	cfg := config{}
	for _, opt := range opts {
		opt(&cfg)
	}

	initMu.Lock()
	defer initMu.Unlock()

	if initDone {
		return nil
	}
	initErr = initRuntime(cfg)
	if initErr != nil {
		return initErr
	}
	initDone = true
	return nil
}

func initRuntime(cfg config) error {
	h := hooks
	if h.open == nil {
		h.open = internalloader.Open
	}
	if h.lookup == nil {
		h.lookup = internalloader.Lookup
	}
	if h.close == nil {
		h.close = internalloader.Close
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
		_ = h.close(lib, internalloader.PuregoClose)
		return fmt.Errorf("vulkan: resolve vkGetInstanceProcAddr: %w", err)
	}
	h.register(&vkGetInstanceProcAddr, addr)
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
	initMu = sync.Mutex{}
	initDone = false
	initErr = nil
	vkGetInstanceProcAddr = nil
	hooks = defaultRuntimeHooks
}

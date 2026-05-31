package loader

import (
	"errors"
	"fmt"

	"github.com/bnema/purego"
)

// SharedLibrary is a dlopen handle paired with the path that succeeded.
type SharedLibrary struct {
	Handle uintptr
	Path   string
}

// OpenFunc opens a shared library path with a dlopen mode.
type OpenFunc func(path string, mode int) (uintptr, error)

// LookupFunc resolves a symbol address from a shared library handle.
type LookupFunc func(handle uintptr, symbol string) (uintptr, error)

// CloseFunc closes a shared library handle.
type CloseFunc func(handle uintptr) error

// PuregoOpen opens a shared library through github.com/bnema/purego.
func PuregoOpen(path string, mode int) (uintptr, error) {
	return purego.Dlopen(path, mode)
}

// PuregoLookup resolves a shared library symbol through github.com/bnema/purego.
func PuregoLookup(handle uintptr, symbol string) (uintptr, error) {
	return purego.Dlsym(handle, symbol)
}

// PuregoClose closes a shared library through github.com/bnema/purego.
func PuregoClose(handle uintptr) error {
	return purego.Dlclose(handle)
}

// Open tries each candidate in order and returns the first successfully opened
// shared library.
func Open(candidates []string, open OpenFunc) (SharedLibrary, error) {
	if open == nil {
		open = PuregoOpen
	}
	if len(candidates) == 0 {
		return SharedLibrary{}, errors.New("loader: no library candidates")
	}

	var errs []error
	for _, candidate := range candidates {
		if candidate == "" {
			continue
		}
		handle, err := open(candidate, purego.RTLD_NOW|purego.RTLD_GLOBAL)
		if err == nil {
			return SharedLibrary{Handle: handle, Path: candidate}, nil
		}
		errs = append(errs, fmt.Errorf("%s: %w", candidate, err))
	}
	if len(errs) == 0 {
		return SharedLibrary{}, errors.New("loader: no non-empty library candidates")
	}
	return SharedLibrary{}, fmt.Errorf("loader: open shared library: %w", errors.Join(errs...))
}

// Lookup resolves name in lib and wraps missing-symbol errors with useful
// context.
func Lookup(lib SharedLibrary, name string, lookup LookupFunc) (uintptr, error) {
	if lookup == nil {
		lookup = PuregoLookup
	}
	addr, err := lookup(lib.Handle, name)
	if err != nil {
		return 0, fmt.Errorf("loader: resolve %s from %s: %w", name, lib.Path, err)
	}
	if addr == 0 {
		return 0, fmt.Errorf("loader: resolve %s from %s: zero address", name, lib.Path)
	}
	return addr, nil
}

// Close closes lib if it has a non-zero handle.
func Close(lib SharedLibrary, close CloseFunc) error {
	if lib.Handle == 0 {
		return nil
	}
	if close == nil {
		close = PuregoClose
	}
	if err := close(lib.Handle); err != nil {
		return fmt.Errorf("loader: close %s: %w", lib.Path, err)
	}
	return nil
}

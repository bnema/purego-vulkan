package vulkan

import (
	"errors"
	"strings"
	"testing"

	internalloader "github.com/bnema/purego-vulkan/internal/loader"
)

func TestInitRunsOnlyOnce(t *testing.T) {
	resetInitForTest()
	defer resetInitForTest()

	var opens int
	setRuntimeHooksForTest(runtimeHooks{
		open: func(candidates []string, open internalloader.OpenFunc) (internalloader.SharedLibrary, error) {
			opens++
			return internalloader.SharedLibrary{Handle: 7, Path: candidates[0]}, nil
		},
		lookup: func(lib internalloader.SharedLibrary, name string, lookup internalloader.LookupFunc) (uintptr, error) {
			return 0x1000, nil
		},
		register: func(fptr any, addr uintptr) {},
	})

	if err := Init(); err != nil {
		t.Fatalf("first Init() error = %v", err)
	}
	if err := Init(); err != nil {
		t.Fatalf("second Init() error = %v", err)
	}
	if opens != 1 {
		t.Fatalf("open count = %d, want 1", opens)
	}
}

func TestInitReturnsLookupErrorWhenVkGetInstanceProcAddrMissing(t *testing.T) {
	resetInitForTest()
	defer resetInitForTest()

	var closes int
	setRuntimeHooksForTest(runtimeHooks{
		open: func(candidates []string, open internalloader.OpenFunc) (internalloader.SharedLibrary, error) {
			return internalloader.SharedLibrary{Handle: 7, Path: "libvulkan.so.1"}, nil
		},
		lookup: func(lib internalloader.SharedLibrary, name string, lookup internalloader.LookupFunc) (uintptr, error) {
			return 0, errors.New("missing")
		},
		close: func(lib internalloader.SharedLibrary, close internalloader.CloseFunc) error {
			closes++
			return nil
		},
		register: func(fptr any, addr uintptr) {},
	})

	err := Init()
	if err == nil {
		t.Fatal("Init() error = nil")
	}
	if !strings.Contains(err.Error(), "vkGetInstanceProcAddr") {
		t.Fatalf("Init() error = %q", err.Error())
	}
	if closes != 1 {
		t.Fatalf("close count = %d, want 1", closes)
	}
}

func TestInitUsesExplicitLibraryPath(t *testing.T) {
	resetInitForTest()
	defer resetInitForTest()

	var candidates []string
	setRuntimeHooksForTest(runtimeHooks{
		open: func(got []string, open internalloader.OpenFunc) (internalloader.SharedLibrary, error) {
			candidates = append(candidates, got...)
			return internalloader.SharedLibrary{Handle: 7, Path: got[0]}, nil
		},
		lookup: func(lib internalloader.SharedLibrary, name string, lookup internalloader.LookupFunc) (uintptr, error) {
			return 0x1000, nil
		},
		register: func(fptr any, addr uintptr) {},
	})

	if err := Init(WithLibraryPath("/custom/libvulkan.so.1")); err != nil {
		t.Fatalf("Init() error = %v", err)
	}
	if len(candidates) == 0 || candidates[0] != "/custom/libvulkan.so.1" {
		t.Fatalf("candidates = %#v", candidates)
	}
}

func TestInitReturnsLookupAndCloseErrors(t *testing.T) {
	resetInitForTest()
	defer resetInitForTest()

	setRuntimeHooksForTest(runtimeHooks{
		open: func(candidates []string, open internalloader.OpenFunc) (internalloader.SharedLibrary, error) {
			return internalloader.SharedLibrary{Handle: 7, Path: "libvulkan.so.1"}, nil
		},
		lookup: func(lib internalloader.SharedLibrary, name string, lookup internalloader.LookupFunc) (uintptr, error) {
			return 0, errors.New("missing symbol")
		},
		close: func(lib internalloader.SharedLibrary, close internalloader.CloseFunc) error {
			return errors.New("close failed")
		},
		register: func(fptr any, addr uintptr) {},
	})

	err := Init()
	if err == nil {
		t.Fatal("Init() error = nil")
	}
	if !strings.Contains(err.Error(), "missing symbol") || !strings.Contains(err.Error(), "close failed") {
		t.Fatalf("Init() error = %q", err.Error())
	}
}

func TestInitCanRetryAfterFailureWithExplicitLibraryPath(t *testing.T) {
	resetInitForTest()
	defer resetInitForTest()

	setRuntimeHooksForTest(runtimeHooks{
		open: func(candidates []string, open internalloader.OpenFunc) (internalloader.SharedLibrary, error) {
			if candidates[0] != "/custom/libvulkan.so.1" {
				return internalloader.SharedLibrary{}, errors.New("default path unavailable")
			}
			return internalloader.SharedLibrary{Handle: 7, Path: candidates[0]}, nil
		},
		lookup: func(lib internalloader.SharedLibrary, name string, lookup internalloader.LookupFunc) (uintptr, error) {
			return 0x1000, nil
		},
		register: func(fptr any, addr uintptr) {},
	})

	if err := Init(); err == nil {
		t.Fatal("first Init() error = nil")
	}
	if err := Init(WithLibraryPath("/custom/libvulkan.so.1")); err != nil {
		t.Fatalf("retry Init() error = %v", err)
	}
}

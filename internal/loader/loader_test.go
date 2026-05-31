package loader

import (
	"errors"
	"strings"
	"testing"
)

func TestOpenPrefersExplicitLibraryPath(t *testing.T) {
	var opened []string
	lib, err := Open([]string{"/opt/vulkan/libvulkan.so.1", "libvulkan.so.1"}, func(path string, mode int) (uintptr, error) {
		opened = append(opened, path)
		return 99, nil
	})
	if err != nil {
		t.Fatalf("Open() error = %v", err)
	}
	if lib.Handle != 99 || lib.Path != "/opt/vulkan/libvulkan.so.1" {
		t.Fatalf("Open() = %+v", lib)
	}
	if len(opened) != 1 || opened[0] != "/opt/vulkan/libvulkan.so.1" {
		t.Fatalf("opened candidates = %#v", opened)
	}
}

func TestOpenFallsBackToVersionedThenUnversionedSoname(t *testing.T) {
	var opened []string
	lib, err := Open([]string{"libvulkan.so.1", "libvulkan.so"}, func(path string, mode int) (uintptr, error) {
		opened = append(opened, path)
		if path == "libvulkan.so.1" {
			return 0, errors.New("missing versioned")
		}
		return 123, nil
	})
	if err != nil {
		t.Fatalf("Open() error = %v", err)
	}
	if lib.Handle != 123 || lib.Path != "libvulkan.so" {
		t.Fatalf("Open() = %+v", lib)
	}
	if got := strings.Join(opened, ","); got != "libvulkan.so.1,libvulkan.so" {
		t.Fatalf("opened candidates = %s", got)
	}
}

func TestOpenReturnsJoinedCandidateErrors(t *testing.T) {
	_, err := Open([]string{"libvulkan.so.1", "libvulkan.so"}, func(path string, mode int) (uintptr, error) {
		return 0, errors.New("nope")
	})
	if err == nil {
		t.Fatal("Open() error = nil")
	}
	msg := err.Error()
	if !strings.Contains(msg, "libvulkan.so.1") || !strings.Contains(msg, "libvulkan.so") {
		t.Fatalf("error %q does not mention all candidates", msg)
	}
}

func TestLookupWrapsMissingSymbol(t *testing.T) {
	_, err := Lookup(SharedLibrary{Handle: 1, Path: "libvulkan.so.1"}, "vkGetInstanceProcAddr", func(handle uintptr, symbol string) (uintptr, error) {
		return 0, errors.New("not found")
	})
	if err == nil {
		t.Fatal("Lookup() error = nil")
	}
	if !strings.Contains(err.Error(), "vkGetInstanceProcAddr") {
		t.Fatalf("Lookup() error = %q", err.Error())
	}
}

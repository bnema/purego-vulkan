package integration_test

import (
	"testing"

	"bnema.dev/purego"
	"bnema.dev/purego-vulkan/vulkan"
)

func TestInitWithSystemLoader(t *testing.T) {
	h, err := purego.Dlopen("libvulkan.so.1", purego.RTLD_NOW|purego.RTLD_GLOBAL)
	if err != nil {
		t.Skipf("libvulkan.so.1 not available: %v", err)
	}
	if err := purego.Dlclose(h); err != nil {
		t.Fatalf("close probe handle: %v", err)
	}

	if err := vulkan.Init(); err != nil {
		t.Fatalf("Init() error = %v", err)
	}
}

package integration_test

import (
	"testing"

	"github.com/bnema/purego"
	"github.com/bnema/purego-vulkan/vulkan"
)

func TestInitWithSystemLoader(t *testing.T) {
	h, err := purego.Dlopen("libvulkan.so.1", purego.RTLD_NOW|purego.RTLD_GLOBAL)
	if err != nil {
		t.Skipf("libvulkan.so.1 not available: %v", err)
	}
	_ = purego.Dlclose(h)

	if err := vulkan.Init(); err != nil {
		t.Fatalf("Init() error = %v", err)
	}
}

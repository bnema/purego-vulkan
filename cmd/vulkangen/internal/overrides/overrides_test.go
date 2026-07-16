package overrides_test

import (
	"path/filepath"
	"testing"

	"github.com/bnema/purego-vulkan/cmd/vulkangen/internal/model"
	"github.com/bnema/purego-vulkan/cmd/vulkangen/internal/overrides"
	"github.com/bnema/purego-vulkan/cmd/vulkangen/internal/parser"
)

func TestWSIProfileSelectsLinuxPresentationAndReadbackSurface(t *testing.T) {
	reg := parsePinnedRegistry(t)
	cfg, err := overrides.SelectionForProfile(reg, overrides.ProfileWSI)
	if err != nil {
		t.Fatalf("SelectionForProfile(wsi) error = %v", err)
	}
	sel, err := model.Select(reg, cfg)
	if err != nil {
		t.Fatalf("Select(wsi) error = %v", err)
	}

	for _, tt := range []struct {
		name     string
		optional bool
	}{
		{"vkCreateWaylandSurfaceKHR", true},
		{"vkCreateXcbSurfaceKHR", true},
		{"vkCreateXlibSurfaceKHR", true},
		{"vkCreateSwapchainKHR", true},
		{"vkGetSwapchainImagesKHR", true},
		{"vkAcquireNextImageKHR", true},
		{"vkQueuePresentKHR", true},
		{"vkCmdCopyImageToBuffer", false},
		{"vkCmdClearColorImage", false},
		{"vkGetImageSubresourceLayout", false},
		{"vkGetSemaphoreCounterValue", true},
		{"vkGetSemaphoreCounterValueKHR", true},
		{"vkWaitSemaphores", true},
		{"vkWaitSemaphoresKHR", true},
		{"vkSignalSemaphore", true},
		{"vkSignalSemaphoreKHR", true},
	} {
		cmd := sel.CommandByName(tt.name)
		if cmd == nil {
			t.Fatalf("WSI profile missing command %s", tt.name)
		}
		if cmd.Optional != tt.optional {
			t.Fatalf("WSI command %s optional = %t, want %t", tt.name, cmd.Optional, tt.optional)
		}
	}

	for _, name := range []string{
		"VK_KHR_surface",
		"VK_KHR_swapchain",
		"VK_KHR_wayland_surface",
		"VK_KHR_xcb_surface",
		"VK_KHR_xlib_surface",
		"VK_KHR_timeline_semaphore",
	} {
		if sel.ExtensionByName(name) == nil {
			t.Fatalf("WSI profile missing extension %s", name)
		}
	}

	for _, name := range []string{"wl_display", "wl_surface", "xcb_connection_t", "xcb_window_t", "Display", "Window"} {
		if sel.TypeByName(name) == nil {
			t.Fatalf("WSI profile missing native platform type %s", name)
		}
	}
}

func TestCompleteProfileSelectsPinnedRegistryCommandCoverageAsOptionalExpansion(t *testing.T) {
	reg := parsePinnedRegistry(t)
	cfg, err := overrides.SelectionForProfile(reg, overrides.ProfileComplete)
	if err != nil {
		t.Fatalf("SelectionForProfile(complete) error = %v", err)
	}
	sel, err := model.Select(reg, cfg)
	if err != nil {
		t.Fatalf("Select(complete) error = %v", err)
	}

	if got, wantAtLeast := len(sel.Commands), 850; got < wantAtLeast {
		t.Fatalf("complete profile selected %d commands, want at least %d", got, wantAtLeast)
	}
	for _, name := range []string{
		"vkCreateSwapchainKHR",
		"vkQueuePresentKHR",
		"vkCmdCopyImageToBuffer",
		"vkCreateRenderPass",
		"vkCreateFramebuffer",
		"vkCreatePipelineCache",
		"vkCreateComputePipelines",
		"vkCmdSetViewport",
		"vkCreateDebugUtilsMessengerEXT",
		"vkCreateQueryPool",
	} {
		cmd := sel.CommandByName(name)
		if cmd == nil {
			t.Fatalf("complete profile missing command %s", name)
		}
		if !cmd.Optional {
			t.Fatalf("complete expansion command %s optional = false", name)
		}
	}

	for _, name := range []string{"vkCreateInstance", "vkDestroyInstance", "vkCreateDevice", "vkDestroyDevice", "vkCmdCopyBufferToImage"} {
		cmd := sel.CommandByName(name)
		if cmd == nil {
			t.Fatalf("complete profile missing required renderer command %s", name)
		}
		if cmd.Optional {
			t.Fatalf("required renderer command %s optional = true", name)
		}
	}
}

func parsePinnedRegistry(t *testing.T) *model.Registry {
	t.Helper()
	reg, err := parser.ParseFile(filepath.Join("..", "..", "..", "..", "registry", "vk.xml"))
	if err != nil {
		t.Fatalf("ParseFile() error = %v", err)
	}
	return reg
}

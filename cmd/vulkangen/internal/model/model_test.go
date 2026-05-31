package model_test

import (
	"testing"

	"github.com/bnema/purego-vulkan/cmd/vulkangen/internal/model"
	"github.com/bnema/purego-vulkan/cmd/vulkangen/internal/overrides"
)

func TestNormalizeClassifiesDispatchableAndNonDispatchableHandles(t *testing.T) {
	reg := testRegistry()
	sel, err := model.Select(reg, model.SelectionConfig{RootTypes: []string{"VkInstance", "VkImage", "VkBool32"}})
	if err != nil {
		t.Fatalf("Select() error = %v", err)
	}

	instance := sel.TypeByName("VkInstance")
	if instance == nil {
		t.Fatal("VkInstance missing")
	}
	if !instance.Dispatchable || instance.GoType != "uintptr" {
		t.Fatalf("VkInstance selected as %+v", *instance)
	}

	image := sel.TypeByName("VkImage")
	if image == nil {
		t.Fatal("VkImage missing")
	}
	if image.Dispatchable || image.GoType != "uint64" {
		t.Fatalf("VkImage selected as %+v", *image)
	}

	bool32 := sel.TypeByName("VkBool32")
	if bool32 == nil || bool32.GoType != "uint32" {
		t.Fatalf("VkBool32 selected as %+v", bool32)
	}
}

func TestSelectCommandsBuildsDependencyClosure(t *testing.T) {
	reg := testRegistry()
	sel, err := model.Select(reg, model.SelectionConfig{Commands: []string{"vkCreateInstance"}})
	if err != nil {
		t.Fatalf("Select() error = %v", err)
	}

	for _, name := range []string{"VkResult", "VkInstanceCreateInfo", "VkStructureType", "VkInstance"} {
		if sel.TypeByName(name) == nil {
			t.Fatalf("dependent type %s missing from selection", name)
		}
	}
	cmd := sel.CommandByName("vkCreateInstance")
	if cmd == nil {
		t.Fatal("vkCreateInstance missing")
	}
	if cmd.Dispatch != model.DispatchGlobal {
		t.Fatalf("vkCreateInstance dispatch = %q, want global", cmd.Dispatch)
	}
}

func TestOverridesMarkOptionalExtensionCommands(t *testing.T) {
	reg := testRegistry()
	cfg := model.SelectionConfig{
		Extensions: []string{"VK_KHR_external_memory_fd"},
		CommandOverrides: map[string]model.CommandOverride{
			"vkGetMemoryFdKHR": {Optional: true},
		},
	}
	sel, err := model.Select(reg, cfg)
	if err != nil {
		t.Fatalf("Select() error = %v", err)
	}
	cmd := sel.CommandByName("vkGetMemoryFdKHR")
	if cmd == nil {
		t.Fatal("vkGetMemoryFdKHR missing")
	}
	if !cmd.Optional {
		t.Fatalf("vkGetMemoryFdKHR optional = false")
	}
	if cmd.Dispatch != model.DispatchDevice {
		t.Fatalf("vkGetMemoryFdKHR dispatch = %q, want device", cmd.Dispatch)
	}
}

func TestSelectionFiltersToLinuxCompositorSubset(t *testing.T) {
	reg := testRegistry()
	cfg := overrides.DefaultSelection()
	cfg.Commands = []string{"vkCreateInstance"}
	sel, err := model.Select(reg, cfg)
	if err != nil {
		t.Fatalf("Select() error = %v", err)
	}

	if sel.ExtensionByName("VK_KHR_external_memory_fd") == nil {
		t.Fatal("expected Linux fd extension to be selected")
	}
	if sel.ExtensionByName("VK_KHR_win32_keyed_mutex") != nil {
		t.Fatal("selected win32-only extension")
	}
	if sel.CommandByName("vkGetMemoryWin32HandleKHR") != nil {
		t.Fatal("selected win32-only command")
	}
}

func testRegistry() *model.Registry {
	return &model.Registry{
		Types: []model.TypeDecl{
			{Name: "VkInstance", Category: "handle", Type: "VK_DEFINE_HANDLE"},
			{Name: "VkPhysicalDevice", Category: "handle", Type: "VK_DEFINE_HANDLE", Parent: "VkInstance"},
			{Name: "VkDevice", Category: "handle", Type: "VK_DEFINE_HANDLE", Parent: "VkPhysicalDevice"},
			{Name: "VkDeviceMemory", Category: "handle", Type: "VK_DEFINE_NON_DISPATCHABLE_HANDLE", Parent: "VkDevice"},
			{Name: "VkImage", Category: "handle", Type: "VK_DEFINE_NON_DISPATCHABLE_HANDLE", Parent: "VkDevice"},
			{Name: "VkBool32", Category: "basetype", Type: "uint32_t"},
			{Name: "VkResult", Category: "basetype", Type: "int32_t"},
			{Name: "VkStructureType", Category: "enum"},
			{Name: "VkInstanceCreateInfo", Category: "struct", Members: []model.MemberDecl{
				{Name: "sType", Type: "VkStructureType"},
				{Name: "pNext", Type: "void", Const: true, PointerDepth: 1},
				{Name: "enabledExtensionCount", Type: "uint32_t"},
				{Name: "ppEnabledExtensionNames", Type: "char", Const: true, PointerDepth: 2},
			}},
			{Name: "VkMemoryGetFdInfoKHR", Category: "struct", Members: []model.MemberDecl{
				{Name: "sType", Type: "VkStructureType"},
				{Name: "pNext", Type: "void", Const: true, PointerDepth: 1},
				{Name: "memory", Type: "VkDeviceMemory"},
			}},
			{Name: "VkAllocationCallbacks", Category: "struct"},
		},
		Commands: []model.CommandDecl{
			{Name: "vkCreateInstance", Return: "VkResult", Params: []model.ParamDecl{
				{Name: "pCreateInfo", Type: "VkInstanceCreateInfo", Const: true, PointerDepth: 1},
				{Name: "pAllocator", Type: "VkAllocationCallbacks", Const: true, PointerDepth: 1, Optional: "true"},
				{Name: "pInstance", Type: "VkInstance", PointerDepth: 1},
			}},
			{Name: "vkGetMemoryFdKHR", Return: "VkResult", Params: []model.ParamDecl{
				{Name: "device", Type: "VkDevice"},
				{Name: "pGetFdInfo", Type: "VkMemoryGetFdInfoKHR", Const: true, PointerDepth: 1},
			}},
			{Name: "vkGetMemoryWin32HandleKHR", Return: "VkResult", Params: []model.ParamDecl{
				{Name: "device", Type: "VkDevice"},
			}},
		},
		Extensions: []model.ExtensionDecl{
			{Name: "VK_KHR_external_memory_fd", Type: "device", Supported: "vulkan", Requires: []model.RequireDecl{{
				Types:    []string{"VkMemoryGetFdInfoKHR"},
				Commands: []string{"vkGetMemoryFdKHR"},
				Enums:    []model.EnumDecl{{Name: "VK_KHR_EXTERNAL_MEMORY_FD_EXTENSION_NAME", Value: "\"VK_KHR_external_memory_fd\""}},
			}}},
			{Name: "VK_KHR_win32_keyed_mutex", Type: "device", Platform: "win32", Supported: "vulkan", Requires: []model.RequireDecl{{
				Commands: []string{"vkGetMemoryWin32HandleKHR"},
			}}},
		},
	}
}

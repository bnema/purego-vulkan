package model_test

import (
	"path/filepath"
	"testing"

	"github.com/bnema/purego-vulkan/cmd/vulkangen/internal/model"
	"github.com/bnema/purego-vulkan/cmd/vulkangen/internal/overrides"
	"github.com/bnema/purego-vulkan/cmd/vulkangen/internal/parser"
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
	if !containsConstant(sel, "VK_SUCCESS") || !containsConstant(sel, "VK_ERROR_OUT_OF_HOST_MEMORY") {
		t.Fatalf("expected VkResult constants in dependency closure, got %+v", sel.Constants)
	}
}

func TestSelectReportsMissingConfiguredCommandsAndDependencies(t *testing.T) {
	reg := testRegistry()
	_, err := model.Select(reg, model.SelectionConfig{Commands: []string{"vkTypoCommand"}})
	if err == nil {
		t.Fatal("Select() error = nil for missing configured command")
	}

	reg = testRegistry()
	reg.Types = append(reg.Types, model.TypeDecl{Name: "VkBrokenInfo", Category: "struct", Members: []model.MemberDecl{{Name: "missing", Type: "VkMissingType"}}})
	reg.Commands = append(reg.Commands, model.CommandDecl{Name: "vkBroken", Return: "void", Params: []model.ParamDecl{{Name: "info", Type: "VkBrokenInfo", PointerDepth: 1}}})
	_, err = model.Select(reg, model.SelectionConfig{Commands: []string{"vkBroken"}})
	if err == nil {
		t.Fatal("Select() error = nil for missing dependent type")
	}
}

func TestSelectIncludesFeatureConstantsForSelectedTypes(t *testing.T) {
	reg := testRegistry()
	sel, err := model.Select(reg, model.SelectionConfig{Commands: []string{"vkGetPhysicalDeviceProperties2"}})
	if err != nil {
		t.Fatalf("Select() error = %v", err)
	}
	if !containsConstant(sel, "VK_STRUCTURE_TYPE_PHYSICAL_DEVICE_PROPERTIES_2") {
		t.Fatalf("expected feature constant for selected sType, got %+v", sel.Constants)
	}
}

func TestSelectIncludesVulkanTypedefDependencies(t *testing.T) {
	reg := testRegistry()
	sel, err := model.Select(reg, model.SelectionConfig{Commands: []string{"vkAllocateMemory"}, RootTypes: []string{"VkAccessFlags2"}})
	if err != nil {
		t.Fatalf("Select() error = %v", err)
	}
	for _, name := range []string{"VkDeviceSize", "VkFlags64"} {
		if sel.TypeByName(name) == nil {
			t.Fatalf("%s typedef missing from selected types: %+v", name, sel.Types)
		}
	}
}

func TestSelectResolvesAliasCommandsAndTypes(t *testing.T) {
	reg := testRegistry()
	sel, err := model.Select(reg, model.SelectionConfig{Extensions: []string{"VK_KHR_get_physical_device_properties2"}})
	if err != nil {
		t.Fatalf("Select() error = %v", err)
	}

	aliasCmd := sel.CommandByName("vkGetPhysicalDeviceProperties2KHR")
	if aliasCmd == nil {
		t.Fatal("vkGetPhysicalDeviceProperties2KHR missing")
	}
	if aliasCmd.Return != "void" || len(aliasCmd.Params) != 2 {
		t.Fatalf("alias command signature = return %q params %+v", aliasCmd.Return, aliasCmd.Params)
	}
	if aliasCmd.Dispatch != model.DispatchInstance {
		t.Fatalf("alias command dispatch = %q, want instance", aliasCmd.Dispatch)
	}

	aliasType := sel.TypeByName("VkAccessFlags2KHR")
	if aliasType == nil {
		t.Fatal("VkAccessFlags2KHR missing")
	}
	if aliasType.GoType != "uint64" {
		t.Fatalf("VkAccessFlags2KHR GoType = %q, want uint64", aliasType.GoType)
	}
}

func TestSelectIncludesExtensionDependsClosure(t *testing.T) {
	reg := testRegistry()
	sel, err := model.Select(reg, model.SelectionConfig{Extensions: []string{"VK_EXT_external_memory_dma_buf"}})
	if err != nil {
		t.Fatalf("Select() error = %v", err)
	}
	if sel.ExtensionByName("VK_EXT_external_memory_dma_buf") == nil {
		t.Fatal("VK_EXT_external_memory_dma_buf missing")
	}
	if sel.ExtensionByName("VK_KHR_external_memory_fd") == nil {
		t.Fatal("dependent VK_KHR_external_memory_fd extension missing")
	}
	if sel.CommandByName("vkGetMemoryFdKHR") == nil {
		t.Fatal("dependent extension command vkGetMemoryFdKHR missing")
	}
}

func TestSelectFiltersDuplicateCommandVariants(t *testing.T) {
	reg := testRegistry()
	sel, err := model.Select(reg, model.SelectionConfig{Commands: []string{"vkCreateDevice"}})
	if err != nil {
		t.Fatalf("Select() error = %v", err)
	}
	var count int
	for _, cmd := range sel.Commands {
		if cmd.Name == "vkCreateDevice" {
			count++
		}
	}
	if count != 1 {
		t.Fatalf("vkCreateDevice selected %d times, want 1", count)
	}
	cmd := sel.CommandByName("vkCreateDevice")
	if cmd == nil || cmd.Dispatch != model.DispatchInstance || len(cmd.Params) != 3 {
		t.Fatalf("vkCreateDevice selected as %+v", cmd)
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

func TestDefaultSelectionOnPinnedRegistryHasNoDuplicateOrBrokenAliasCommands(t *testing.T) {
	reg, err := parser.ParseFile(filepath.Join("..", "..", "..", "..", "registry", "vk.xml"))
	if err != nil {
		t.Fatalf("ParseFile() error = %v", err)
	}
	sel, err := model.Select(reg, overrides.DefaultSelection())
	if err != nil {
		t.Fatalf("Select() error = %v", err)
	}

	seen := make(map[string]bool)
	for _, cmd := range sel.Commands {
		if seen[cmd.Name] {
			t.Fatalf("duplicate selected command %s", cmd.Name)
		}
		seen[cmd.Name] = true
		if cmd.Return == "" && len(cmd.Params) == 0 {
			t.Fatalf("selected command %s has empty alias signature", cmd.Name)
		}
	}

	for _, name := range []string{"VkDeviceSize", "VkFlags", "VkFlags64"} {
		if sel.TypeByName(name) == nil {
			t.Fatalf("%s missing from default selected type closure", name)
		}
	}
	for _, name := range []string{"VK_STRUCTURE_TYPE_PHYSICAL_DEVICE_PROPERTIES_2", "VK_STRUCTURE_TYPE_MEMORY_ALLOCATE_INFO"} {
		if !containsConstant(sel, name) {
			t.Fatalf("feature constant %s missing from default selected constants", name)
		}
	}

	for _, name := range []string{"vkGetPhysicalDeviceProperties2KHR", "vkGetPhysicalDeviceFeatures2KHR", "vkGetPhysicalDeviceQueueFamilyProperties2KHR"} {
		cmd := sel.CommandByName(name)
		if cmd == nil {
			t.Fatalf("%s missing", name)
		}
		if cmd.Dispatch != model.DispatchInstance {
			t.Fatalf("%s dispatch = %q, want instance", name, cmd.Dispatch)
		}
		if len(cmd.Params) == 0 {
			t.Fatalf("%s has no params", name)
		}
	}
}

func TestSelectionFiltersToLinuxCompositorSubset(t *testing.T) {
	reg := testRegistry()
	cfg := overrides.DefaultSelection()
	cfg.Commands = []string{"vkCreateInstance"}
	cfg.Extensions = []string{"VK_KHR_external_memory_fd", "VK_KHR_win32_keyed_mutex"}
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

func containsConstant(sel *model.SelectedRegistry, name string) bool {
	for _, c := range sel.Constants {
		if c.Name == name {
			return true
		}
	}
	return false
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
			{Name: "VkFlags", Category: "basetype", Type: "uint32_t"},
			{Name: "VkFlags64", Category: "basetype", Type: "uint64_t"},
			{Name: "VkDeviceSize", Category: "basetype", Type: "uint64_t"},
			{Name: "VkAccessFlags2", Category: "bitmask", Type: "VkFlags64"},
			{Name: "VkAccessFlags2KHR", Category: "bitmask", Alias: "VkAccessFlags2"},
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
			{Name: "VkMemoryAllocateInfo", Category: "struct", Members: []model.MemberDecl{
				{Name: "sType", Type: "VkStructureType", Values: "VK_STRUCTURE_TYPE_MEMORY_ALLOCATE_INFO"},
				{Name: "pNext", Type: "void", Const: true, PointerDepth: 1},
				{Name: "allocationSize", Type: "VkDeviceSize"},
			}},
			{Name: "VkAllocationCallbacks", Category: "struct"},
			{Name: "VkPhysicalDeviceProperties2", Category: "struct", Members: []model.MemberDecl{
				{Name: "sType", Type: "VkStructureType", Values: "VK_STRUCTURE_TYPE_PHYSICAL_DEVICE_PROPERTIES_2"},
				{Name: "pNext", Type: "void", PointerDepth: 1},
			}},
			{Name: "VkDeviceCreateInfo", Category: "struct", Members: []model.MemberDecl{
				{Name: "sType", Type: "VkStructureType"},
				{Name: "pNext", Type: "void", Const: true, PointerDepth: 1},
			}},
		},
		EnumGroups: []model.EnumGroup{
			{Name: "VkResult", Type: "enum", Enums: []model.EnumDecl{
				{Name: "VK_SUCCESS", Value: "0"},
				{Name: "VK_ERROR_OUT_OF_HOST_MEMORY", Value: "-1"},
			}},
			{Name: "VkStructureType", Type: "enum", Enums: []model.EnumDecl{
				{Name: "VK_STRUCTURE_TYPE_INSTANCE_CREATE_INFO", Value: "1"},
			}},
		},
		Features: []model.FeatureDecl{
			{Name: "VK_VERSION_1_0", Number: "1.0", API: "vulkan", Requires: []model.RequireDecl{{
				Enums: []model.EnumDecl{{Name: "VK_STRUCTURE_TYPE_MEMORY_ALLOCATE_INFO", Value: "5", Extends: "VkStructureType"}},
			}}},
			{Name: "VK_VERSION_1_1", Number: "1.1", API: "vulkan", Requires: []model.RequireDecl{{
				Enums: []model.EnumDecl{{Name: "VK_STRUCTURE_TYPE_PHYSICAL_DEVICE_PROPERTIES_2", Extends: "VkStructureType", Offset: "1"}},
			}}},
		},
		Commands: []model.CommandDecl{
			{Name: "vkCreateInstance", Return: "VkResult", Params: []model.ParamDecl{
				{Name: "pCreateInfo", Type: "VkInstanceCreateInfo", Const: true, PointerDepth: 1},
				{Name: "pAllocator", Type: "VkAllocationCallbacks", Const: true, PointerDepth: 1, Optional: "true"},
				{Name: "pInstance", Type: "VkInstance", PointerDepth: 1},
			}},
			{Name: "vkGetPhysicalDeviceProperties2", Return: "void", API: "vulkan", Params: []model.ParamDecl{
				{Name: "physicalDevice", Type: "VkPhysicalDevice"},
				{Name: "pProperties", Type: "VkPhysicalDeviceProperties2", PointerDepth: 1},
			}},
			{Name: "vkGetPhysicalDeviceProperties2KHR", Alias: "vkGetPhysicalDeviceProperties2"},
			{Name: "vkCreateDevice", Return: "VkResult", API: "vulkan", Params: []model.ParamDecl{
				{Name: "physicalDevice", Type: "VkPhysicalDevice"},
				{Name: "pCreateInfo", Type: "VkDeviceCreateInfo", Const: true, PointerDepth: 1},
				{Name: "pAllocator", Type: "VkAllocationCallbacks", Const: true, PointerDepth: 1},
			}},
			{Name: "vkCreateDevice", Return: "VkResult", API: "vulkansc", Params: []model.ParamDecl{
				{Name: "physicalDevice", Type: "VkPhysicalDevice"},
			}},
			{Name: "vkAllocateMemory", Return: "VkResult", Params: []model.ParamDecl{
				{Name: "device", Type: "VkDevice"},
				{Name: "pAllocateInfo", Type: "VkMemoryAllocateInfo", Const: true, PointerDepth: 1},
				{Name: "pAllocator", Type: "VkAllocationCallbacks", Const: true, PointerDepth: 1},
				{Name: "pMemory", Type: "VkDeviceMemory", PointerDepth: 1},
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
			{Name: "VK_KHR_get_physical_device_properties2", Type: "instance", Supported: "vulkan", Requires: []model.RequireDecl{{
				Types:    []string{"VkPhysicalDeviceProperties2", "VkAccessFlags2KHR"},
				Commands: []string{"vkGetPhysicalDeviceProperties2KHR"},
			}}},
			{Name: "VK_EXT_external_memory_dma_buf", Type: "device", Depends: "VK_KHR_external_memory_fd", Supported: "vulkan", Requires: []model.RequireDecl{{
				Enums: []model.EnumDecl{{Name: "VK_EXTERNAL_MEMORY_HANDLE_TYPE_DMA_BUF_BIT_EXT", Bitpos: "9", Extends: "VkExternalMemoryHandleTypeFlagBits"}},
			}}},
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

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

func TestSelectResolvesTransitiveAliasSTypeConstants(t *testing.T) {
	reg := testRegistry()
	sel, err := model.Select(reg, model.SelectionConfig{RootTypes: []string{"VkPhysicalDeviceVariablePointerFeaturesKHR"}})
	if err != nil {
		t.Fatalf("Select() error = %v", err)
	}
	for _, name := range []string{
		"VK_STRUCTURE_TYPE_PHYSICAL_DEVICE_VARIABLE_POINTERS_FEATURES",
		"VK_STRUCTURE_TYPE_PHYSICAL_DEVICE_VARIABLE_POINTERS_FEATURES_KHR",
		"VK_STRUCTURE_TYPE_PHYSICAL_DEVICE_VARIABLE_POINTER_FEATURES_KHR",
	} {
		constant := constantByName(sel, name)
		if constant == nil {
			t.Fatalf("%s missing from constants: %+v", name, sel.Constants)
		}
		if constant.Value == "" {
			t.Fatalf("%s unresolved: %+v", name, *constant)
		}
	}
}

func TestSelectResolvesAliasSTypeConstants(t *testing.T) {
	reg := testRegistry()
	sel, err := model.Select(reg, model.SelectionConfig{RootTypes: []string{"VkPhysicalDeviceProperties2KHR"}})
	if err != nil {
		t.Fatalf("Select() error = %v", err)
	}
	alias := constantByName(sel, "VK_STRUCTURE_TYPE_PHYSICAL_DEVICE_PROPERTIES_2_KHR")
	if alias == nil {
		t.Fatalf("alias sType constant missing from selection: %+v", sel.Constants)
	}
	if alias.Value == "" {
		t.Fatalf("alias sType constant was not resolved: %+v", *alias)
	}
	base := constantByName(sel, "VK_STRUCTURE_TYPE_PHYSICAL_DEVICE_PROPERTIES_2")
	if base == nil {
		t.Fatalf("base sType constant missing from selection: %+v", sel.Constants)
	}
	if alias.Value != base.Value {
		t.Fatalf("alias value = %q, base value = %q", alias.Value, base.Value)
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

func TestSelectRepresentsOpaqueNativeBasetypes(t *testing.T) {
	reg := &model.Registry{Types: []model.TypeDecl{
		{Name: "ANativeWindow", Category: "basetype", RawText: "struct ANativeWindow ;"},
		{Name: "VkRemoteAddressNV", Category: "basetype", Type: "void"},
	}}
	sel, err := model.Select(reg, model.SelectionConfig{RootTypes: []string{"ANativeWindow", "VkRemoteAddressNV"}})
	if err != nil {
		t.Fatalf("Select() error = %v", err)
	}
	if native := sel.TypeByName("ANativeWindow"); native == nil || native.GoType != "uintptr" {
		t.Fatalf("ANativeWindow selected as %+v, want uintptr opaque handle", native)
	}
	if remote := sel.TypeByName("VkRemoteAddressNV"); remote == nil || remote.GoType != "unsafe.Pointer" {
		t.Fatalf("VkRemoteAddressNV selected as %+v, want unsafe.Pointer", remote)
	}
}

func TestSelectIncludesFunctionPointerReturnTypes(t *testing.T) {
	reg := testRegistry()
	reg.Types = append(reg.Types, model.TypeDecl{Name: "PFN_vkVoidFunction", Category: "funcpointer"})
	reg.Commands = append(reg.Commands, model.CommandDecl{Name: "vkGetInstanceProcAddr", Return: "PFN_vkVoidFunction", Params: []model.ParamDecl{
		{Name: "instance", Type: "VkInstance"},
		{Name: "pName", Type: "char", PointerDepth: 1},
	}})

	sel, err := model.Select(reg, model.SelectionConfig{Commands: []string{"vkGetInstanceProcAddr"}})
	if err != nil {
		t.Fatalf("Select() error = %v", err)
	}
	fn := sel.TypeByName("PFN_vkVoidFunction")
	if fn == nil || fn.GoType != "uintptr" {
		t.Fatalf("PFN_vkVoidFunction selected as %+v", fn)
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

func TestSelectResolvesExtensionAliasConstants(t *testing.T) {
	reg := testRegistry()
	sel, err := model.Select(reg, model.SelectionConfig{Extensions: []string{"VK_KHR_get_physical_device_properties2"}})
	if err != nil {
		t.Fatalf("Select() error = %v", err)
	}
	alias := constantByName(sel, "VK_STRUCTURE_TYPE_PHYSICAL_DEVICE_PROPERTIES_2_KHR")
	if alias == nil {
		t.Fatalf("extension alias constant missing: %+v", sel.Constants)
	}
	if alias.Value == "" {
		t.Fatalf("extension alias constant unresolved: %+v", *alias)
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

func TestSelectUsesPreferredDuplicateTypeVariant(t *testing.T) {
	reg := &model.Registry{
		Types: []model.TypeDecl{
			{Name: "VkDevice", Category: "handle", Type: "VK_DEFINE_HANDLE"},
			{Name: "VkPreferredInfo", Category: "struct", API: "vulkansc", Members: []model.MemberDecl{{Name: "scOnly", Type: "uint32_t"}}},
			{Name: "VkPreferredInfo", Category: "struct", API: "vulkan", Members: []model.MemberDecl{{Name: "value", Type: "VkDeviceSize"}}},
			{Name: "VkDeviceSize", Category: "basetype", Type: "uint64_t"},
		},
		Commands: []model.CommandDecl{{Name: "vkUsePreferredInfo", Return: "void", Params: []model.ParamDecl{
			{Name: "device", Type: "VkDevice"},
			{Name: "pInfo", Type: "VkPreferredInfo", PointerDepth: 1},
		}}},
	}

	sel, err := model.Select(reg, model.SelectionConfig{Commands: []string{"vkUsePreferredInfo"}})
	if err != nil {
		t.Fatalf("Select() error = %v", err)
	}
	info := sel.TypeByName("VkPreferredInfo")
	if info == nil {
		t.Fatal("VkPreferredInfo missing")
	}
	if len(info.Members) != 1 || info.Members[0].Name != "value" {
		t.Fatalf("selected type members = %+v, want Vulkan variant", info.Members)
	}
}

func TestSelectScoresTypeAPITokensOrderIndependently(t *testing.T) {
	reg := &model.Registry{
		Types: []model.TypeDecl{
			{Name: "VkDevice", Category: "handle", Type: "VK_DEFINE_HANDLE"},
			{Name: "VkPreferredInfo", Category: "struct", API: "vulkansc", Members: []model.MemberDecl{{Name: "scOnly", Type: "uint32_t"}}},
			{Name: "VkPreferredInfo", Category: "struct", API: "vulkansc,vulkan", Members: []model.MemberDecl{{Name: "value", Type: "VkDeviceSize"}}},
			{Name: "VkDeviceSize", Category: "basetype", Type: "uint64_t"},
		},
		Commands: []model.CommandDecl{{Name: "vkUsePreferredInfo", Return: "void", Params: []model.ParamDecl{
			{Name: "device", Type: "VkDevice"},
			{Name: "pInfo", Type: "VkPreferredInfo", PointerDepth: 1},
		}}},
	}

	sel, err := model.Select(reg, model.SelectionConfig{Commands: []string{"vkUsePreferredInfo"}})
	if err != nil {
		t.Fatalf("Select() error = %v", err)
	}
	info := sel.TypeByName("VkPreferredInfo")
	if info == nil {
		t.Fatal("VkPreferredInfo missing")
	}
	if len(info.Members) != 1 || info.Members[0].Name != "value" {
		t.Fatalf("selected type members = %+v, want combined Vulkan token variant", info.Members)
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

func TestSelectScoresCommandAPITokensOrderIndependently(t *testing.T) {
	reg := &model.Registry{
		Types: []model.TypeDecl{
			{Name: "VkDevice", Category: "handle", Type: "VK_DEFINE_HANDLE"},
		},
		Commands: []model.CommandDecl{
			{Name: "vkUsePreferredCommand", Return: "void", API: "vulkansc", Params: []model.ParamDecl{{Name: "device", Type: "VkDevice"}}},
			{Name: "vkUsePreferredCommand", Return: "void", API: "vulkansc,vulkan", Params: []model.ParamDecl{{Name: "device", Type: "VkDevice"}, {Name: "value", Type: "uint32_t"}}},
		},
	}

	sel, err := model.Select(reg, model.SelectionConfig{Commands: []string{"vkUsePreferredCommand"}})
	if err != nil {
		t.Fatalf("Select() error = %v", err)
	}
	cmd := sel.CommandByName("vkUsePreferredCommand")
	if cmd == nil {
		t.Fatal("vkUsePreferredCommand missing")
	}
	if len(cmd.Params) != 2 || cmd.Params[1].Name != "value" {
		t.Fatalf("selected command params = %+v, want combined Vulkan token variant", cmd.Params)
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

func TestDefaultSelectionDoesNotPullUnrelatedDependencyAlternatives(t *testing.T) {
	reg, err := parser.ParseFile(filepath.Join("..", "..", "..", "..", "registry", "vk.xml"))
	if err != nil {
		t.Fatalf("ParseFile() error = %v", err)
	}
	sel, err := model.Select(reg, overrides.DefaultSelection())
	if err != nil {
		t.Fatalf("Select() error = %v", err)
	}
	for _, name := range []string{"vkCreateRayTracingPipelinesKHR", "vkCreateRayTracingPipelinesNV", "VK_KHR_ray_tracing_pipeline", "VK_NV_ray_tracing"} {
		if sel.CommandByName(name) != nil || sel.ExtensionByName(name) != nil {
			t.Fatalf("default selection pulled unrelated dependency alternative %s", name)
		}
	}
	if len(sel.Commands) > len(overrides.InitialCommands)+40 {
		t.Fatalf("default selection selected %d commands for %d configured commands", len(sel.Commands), len(overrides.InitialCommands))
	}
}

func TestSelectSkipsUnsatisfiedGuardedRequireBlocks(t *testing.T) {
	reg := testRegistry()
	reg.Commands = append(reg.Commands,
		model.CommandDecl{Name: "vkBaseExtensionCommand", Return: "void", Params: []model.ParamDecl{{Name: "device", Type: "VkDevice"}}},
		model.CommandDecl{Name: "vkGuardedExtensionCommand", Return: "void", Params: []model.ParamDecl{{Name: "device", Type: "VkDevice"}}},
	)
	reg.Extensions = append(reg.Extensions,
		model.ExtensionDecl{Name: "VK_EXT_guard_dependency", Type: "device", Supported: "vulkan"},
		model.ExtensionDecl{Name: "VK_EXT_guarded", Type: "device", Supported: "vulkan", Requires: []model.RequireDecl{
			{Commands: []string{"vkBaseExtensionCommand"}},
			{Depends: "VK_EXT_guard_dependency", Commands: []string{"vkGuardedExtensionCommand"}},
		}},
	)

	sel, err := model.Select(reg, model.SelectionConfig{Extensions: []string{"VK_EXT_guarded"}})
	if err != nil {
		t.Fatalf("Select() error = %v", err)
	}
	if sel.CommandByName("vkBaseExtensionCommand") == nil {
		t.Fatal("base extension command missing")
	}
	if sel.CommandByName("vkGuardedExtensionCommand") != nil {
		t.Fatal("guarded command selected without guard dependency")
	}

	sel, err = model.Select(reg, model.SelectionConfig{Extensions: []string{"VK_EXT_guarded", "VK_EXT_guard_dependency"}})
	if err != nil {
		t.Fatalf("Select() with guard dependency error = %v", err)
	}
	if sel.CommandByName("vkGuardedExtensionCommand") == nil {
		t.Fatal("guarded command missing when guard dependency selected")
	}
}

func TestDefaultSelectionOnPinnedRegistryHasRendererReadySurface(t *testing.T) {
	reg, err := parser.ParseFile(filepath.Join("..", "..", "..", "..", "registry", "vk.xml"))
	if err != nil {
		t.Fatalf("ParseFile() error = %v", err)
	}
	sel, err := model.Select(reg, overrides.DefaultSelection())
	if err != nil {
		t.Fatalf("Select() error = %v", err)
	}

	for _, name := range []string{
		"vkGetPhysicalDeviceMemoryProperties", "vkGetPhysicalDeviceMemoryProperties2", "vkGetPhysicalDeviceMemoryProperties2KHR",
		"vkCreateBuffer", "vkDestroyBuffer", "vkGetBufferMemoryRequirements", "vkBindBufferMemory",
		"vkMapMemory", "vkUnmapMemory", "vkFlushMappedMemoryRanges", "vkInvalidateMappedMemoryRanges",
		"vkCmdCopyBufferToImage", "vkCmdPipelineBarrier",
		"vkCreateGraphicsPipelines", "vkDestroyPipeline",
		"vkCmdBindPipeline", "vkCmdBindDescriptorSets", "vkCmdBindVertexBuffers", "vkCmdBindIndexBuffer", "vkCmdDraw", "vkCmdDrawIndexed",
		"vkCmdBeginRendering", "vkCmdEndRendering", "vkCmdBeginRenderingKHR", "vkCmdEndRenderingKHR",
	} {
		cmd := sel.CommandByName(name)
		if cmd == nil {
			t.Fatalf("renderer command %s missing", name)
		}
		wantDispatch := model.DispatchDevice
		if name == "vkGetPhysicalDeviceMemoryProperties" || name == "vkGetPhysicalDeviceMemoryProperties2" || name == "vkGetPhysicalDeviceMemoryProperties2KHR" {
			wantDispatch = model.DispatchInstance
		}
		if cmd.Dispatch != wantDispatch {
			t.Fatalf("%s dispatch = %q, want %q", name, cmd.Dispatch, wantDispatch)
		}
		wantOptional := name == "vkCmdBeginRenderingKHR" || name == "vkCmdEndRenderingKHR" || name == "vkGetPhysicalDeviceMemoryProperties2KHR"
		if cmd.Optional != wantOptional {
			t.Fatalf("%s optional = %t, want %t", name, cmd.Optional, wantOptional)
		}
	}

	for _, name := range []string{
		"VkPhysicalDeviceMemoryProperties", "VkMemoryType", "VkMemoryHeap",
		"VkBuffer", "VkBufferCreateInfo", "VkBufferUsageFlags", "VkBufferUsageFlagBits", "VkMappedMemoryRange",
		"VkBufferImageCopy", "VkImageSubresourceLayers", "VkMemoryBarrier", "VkBufferMemoryBarrier", "VkImageMemoryBarrier",
		"VkPipeline", "VkPipelineCache", "VkGraphicsPipelineCreateInfo", "VkPipelineShaderStageCreateInfo", "VkPipelineVertexInputStateCreateInfo", "VkPipelineInputAssemblyStateCreateInfo", "VkPipelineViewportStateCreateInfo", "VkPipelineRasterizationStateCreateInfo", "VkPipelineMultisampleStateCreateInfo", "VkPipelineColorBlendStateCreateInfo", "VkPipelineDynamicStateCreateInfo",
		"VkRenderingInfo", "VkRenderingInfoKHR", "VkRenderingAttachmentInfo", "VkRenderingAttachmentInfoKHR", "VkPipelineRenderingCreateInfo", "VkPipelineRenderingCreateInfoKHR", "VkPhysicalDeviceDynamicRenderingFeatures", "VkPhysicalDeviceDynamicRenderingFeaturesKHR",
	} {
		if sel.TypeByName(name) == nil {
			t.Fatalf("renderer type %s missing", name)
		}
	}

	for _, name := range []string{
		"VK_BUFFER_USAGE_TRANSFER_SRC_BIT", "VK_BUFFER_USAGE_TRANSFER_DST_BIT", "VK_BUFFER_USAGE_VERTEX_BUFFER_BIT", "VK_BUFFER_USAGE_INDEX_BUFFER_BIT",
		"VK_KHR_DYNAMIC_RENDERING_EXTENSION_NAME",
		"VK_STRUCTURE_TYPE_RENDERING_INFO", "VK_STRUCTURE_TYPE_RENDERING_INFO_KHR", "VK_STRUCTURE_TYPE_RENDERING_ATTACHMENT_INFO", "VK_STRUCTURE_TYPE_RENDERING_ATTACHMENT_INFO_KHR", "VK_STRUCTURE_TYPE_PIPELINE_RENDERING_CREATE_INFO", "VK_STRUCTURE_TYPE_PIPELINE_RENDERING_CREATE_INFO_KHR", "VK_STRUCTURE_TYPE_PHYSICAL_DEVICE_DYNAMIC_RENDERING_FEATURES", "VK_STRUCTURE_TYPE_PHYSICAL_DEVICE_DYNAMIC_RENDERING_FEATURES_KHR",
	} {
		if constantByName(sel, name) == nil {
			t.Fatalf("renderer constant %s missing", name)
		}
	}
	for _, name := range []string{
		"VK_KHR_DYNAMIC_RENDERING_EXTENSION_NAME",
		"VK_STRUCTURE_TYPE_RENDERING_INFO", "VK_STRUCTURE_TYPE_RENDERING_INFO_KHR", "VK_STRUCTURE_TYPE_RENDERING_ATTACHMENT_INFO", "VK_STRUCTURE_TYPE_RENDERING_ATTACHMENT_INFO_KHR", "VK_STRUCTURE_TYPE_PIPELINE_RENDERING_CREATE_INFO", "VK_STRUCTURE_TYPE_PIPELINE_RENDERING_CREATE_INFO_KHR", "VK_STRUCTURE_TYPE_PHYSICAL_DEVICE_DYNAMIC_RENDERING_FEATURES", "VK_STRUCTURE_TYPE_PHYSICAL_DEVICE_DYNAMIC_RENDERING_FEATURES_KHR",
	} {
		constant := constantByName(sel, name)
		if constant.Value == "" {
			t.Fatalf("renderer constant %s has empty value: %+v", name, *constant)
		}
	}

	for _, name := range []string{"vkCreateRenderPass", "vkDestroyRenderPass", "vkCreateFramebuffer", "vkDestroyFramebuffer", "vkCmdBeginRenderPass", "vkCmdEndRenderPass"} {
		if sel.CommandByName(name) != nil {
			t.Fatalf("dynamic rendering selection pulled classic render-pass command %s", name)
		}
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
	for _, name := range []string{"VK_STRUCTURE_TYPE_PHYSICAL_DEVICE_PROPERTIES_2", "VK_STRUCTURE_TYPE_PHYSICAL_DEVICE_PROPERTIES_2_KHR", "VK_STRUCTURE_TYPE_MEMORY_ALLOCATE_INFO", "VK_MAX_PHYSICAL_DEVICE_NAME_SIZE", "VK_UUID_SIZE"} {
		constant := constantByName(sel, name)
		if constant == nil {
			t.Fatalf("feature constant %s missing from default selected constants", name)
		}
		if constant.Value == "" {
			t.Fatalf("feature constant %s has empty value: %+v", name, *constant)
		}
	}
	for _, name := range []string{"VK_STRUCTURE_TYPE_IMPORT_MEMORY_FD_INFO_KHR", "VK_STRUCTURE_TYPE_MEMORY_FD_PROPERTIES_KHR", "VK_STRUCTURE_TYPE_DRM_FORMAT_MODIFIER_PROPERTIES_LIST_EXT", "VK_QUEUE_FAMILY_EXTERNAL_KHR", "VK_LUID_SIZE_KHR", "VK_QUEUE_FAMILY_FOREIGN_EXT"} {
		constant := constantByName(sel, name)
		if constant == nil {
			t.Fatalf("extension constant %s missing", name)
		}
		if constant.Value == "" {
			t.Fatalf("extension constant %s has empty value: %+v", name, *constant)
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
	for _, name := range []string{"vkBindImageMemory2KHR", "vkGetImageMemoryRequirements2KHR"} {
		cmd := sel.CommandByName(name)
		if cmd == nil {
			t.Fatalf("%s missing", name)
		}
		if !cmd.Optional {
			t.Fatalf("%s optional = false, want true for promoted KHR fallback", name)
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
	return constantByName(sel, name) != nil
}

func constantByName(sel *model.SelectedRegistry, name string) *model.SelectedConstant {
	for i := range sel.Constants {
		if sel.Constants[i].Name == name {
			return &sel.Constants[i]
		}
	}
	return nil
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
			{Name: "VkPhysicalDeviceProperties2KHR", Category: "struct", Alias: "VkPhysicalDeviceProperties2"},
			{Name: "VkDeviceCreateInfo", Category: "struct", Members: []model.MemberDecl{
				{Name: "sType", Type: "VkStructureType"},
				{Name: "pNext", Type: "void", Const: true, PointerDepth: 1},
			}},
			{Name: "VkPhysicalDeviceVariablePointersFeatures", Category: "struct", Members: []model.MemberDecl{
				{Name: "sType", Type: "VkStructureType", Values: "VK_STRUCTURE_TYPE_PHYSICAL_DEVICE_VARIABLE_POINTERS_FEATURES"},
				{Name: "pNext", Type: "void", PointerDepth: 1},
				{Name: "variablePointers", Type: "VkBool32"},
			}},
			{Name: "VkPhysicalDeviceVariablePointersFeaturesKHR", Category: "struct", Alias: "VkPhysicalDeviceVariablePointersFeatures"},
			{Name: "VkPhysicalDeviceVariablePointerFeaturesKHR", Category: "struct", Alias: "VkPhysicalDeviceVariablePointersFeaturesKHR"},
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
				Enums: []model.EnumDecl{
					{Name: "VK_STRUCTURE_TYPE_PHYSICAL_DEVICE_PROPERTIES_2", Extends: "VkStructureType", Offset: "1", Value: "1000059001"},
					{Name: "VK_STRUCTURE_TYPE_PHYSICAL_DEVICE_PROPERTIES_2_KHR", Extends: "VkStructureType", Alias: "VK_STRUCTURE_TYPE_PHYSICAL_DEVICE_PROPERTIES_2"},
					{Name: "VK_STRUCTURE_TYPE_PHYSICAL_DEVICE_VARIABLE_POINTERS_FEATURES", Extends: "VkStructureType", Value: "1000120000"},
					{Name: "VK_STRUCTURE_TYPE_PHYSICAL_DEVICE_VARIABLE_POINTERS_FEATURES_KHR", Extends: "VkStructureType", Alias: "VK_STRUCTURE_TYPE_PHYSICAL_DEVICE_VARIABLE_POINTERS_FEATURES"},
					{Name: "VK_STRUCTURE_TYPE_PHYSICAL_DEVICE_VARIABLE_POINTER_FEATURES_KHR", Extends: "VkStructureType", Alias: "VK_STRUCTURE_TYPE_PHYSICAL_DEVICE_VARIABLE_POINTERS_FEATURES_KHR"},
				},
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

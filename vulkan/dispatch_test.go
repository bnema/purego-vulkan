package vulkan

import (
	"reflect"
	"strings"
	"testing"
	"unsafe"
)

func TestCheckWrapsVkResultNames(t *testing.T) {
	if err := Check(Success); err != nil {
		t.Fatalf("Check(Success) error = %v", err)
	}

	err := Check(ErrorOutOfHostMemory)
	if err == nil {
		t.Fatal("Check(ErrorOutOfHostMemory) error = nil")
	}
	if !strings.Contains(err.Error(), "VK_ERROR_OUT_OF_HOST_MEMORY") {
		t.Fatalf("Check(ErrorOutOfHostMemory) error = %q", err.Error())
	}

	err = Check(Result(-999999))
	if err == nil {
		t.Fatal("Check(unknown) error = nil")
	}
	if !strings.Contains(err.Error(), "VK_UNKNOWN_RESULT") {
		t.Fatalf("Check(unknown) error = %q", err.Error())
	}
}

func TestFormatVersion(t *testing.T) {
	version := MakeVersion(1, 3, 275)
	if got := FormatVersion(version); got != "1.3.275" {
		t.Fatalf("FormatVersion() = %q, want 1.3.275", got)
	}

	variantVersion := MakeAPIVersion(2, 1, 3, 275)
	if got := VersionVariant(variantVersion); got != 2 {
		t.Fatalf("VersionVariant() = %d, want 2", got)
	}
	if got := FormatVersion(variantVersion); got != "1.3.275" {
		t.Fatalf("FormatVersion(variant) = %q, want 1.3.275", got)
	}
}

func TestLoadInstanceDispatchRequiresCoreCommands(t *testing.T) {
	resetDispatchForTest()
	defer resetDispatchForTest()

	VkGetInstanceProcAddr = fakeGetInstanceProcAddr(requiredInstanceSymbolsExcept("vkDestroyInstance"))
	_, err := LoadInstanceDispatch(Instance(1))
	if err == nil {
		t.Fatal("LoadInstanceDispatch() error = nil")
	}
	if !strings.Contains(err.Error(), "vkDestroyInstance") {
		t.Fatalf("LoadInstanceDispatch() error = %q", err.Error())
	}

	resetDispatchForTest()
	VkGetInstanceProcAddr = fakeGetInstanceProcAddr(requiredInstanceSymbols())
	dispatch, err := LoadInstanceDispatch(Instance(1))
	if err != nil {
		t.Fatalf("LoadInstanceDispatch() error = %v", err)
	}
	if dispatch.Instance != Instance(1) {
		t.Fatalf("InstanceDispatch.Instance = %v, want 1", dispatch.Instance)
	}
	if dispatch.DestroyInstance == nil {
		t.Fatal("DestroyInstance was not loaded")
	}
	if dispatch.GetPhysicalDeviceMemoryProperties == nil {
		t.Fatal("GetPhysicalDeviceMemoryProperties was not loaded")
	}
	if dispatch.GetPhysicalDeviceMemoryProperties2 == nil || dispatch.GetPhysicalDeviceMemoryProperties2KHR == nil {
		t.Fatal("memory properties2 core/KHR alias group was not populated from one available symbol")
	}
	if dispatch.GetPhysicalDeviceProperties2 == nil || dispatch.GetPhysicalDeviceProperties2KHR == nil {
		t.Fatal("core/KHR alias group was not populated from one available symbol")
	}
	if dispatch.GetPhysicalDeviceFormatProperties2 == nil || dispatch.GetPhysicalDeviceFormatProperties2KHR == nil {
		t.Fatal("format properties2 core/KHR alias group was not populated from one available symbol")
	}

	resetDispatchForTest()
	VkGetInstanceProcAddr = fakeGetInstanceProcAddr(requiredInstanceSymbolsExcept("vkGetPhysicalDeviceMemoryProperties2"))
	dispatch, err = LoadInstanceDispatch(Instance(2))
	if err != nil {
		t.Fatalf("LoadInstanceDispatch() with KHR memory properties2 error = %v", err)
	}
	if dispatch.GetPhysicalDeviceMemoryProperties2 == nil || dispatch.GetPhysicalDeviceMemoryProperties2KHR == nil {
		t.Fatal("memory properties2 core/KHR alias group was not populated from KHR symbol")
	}
}

func TestWSIProfileHasTimelineSemaphoreCommands(t *testing.T) {
	dispatchType := reflect.TypeFor[DeviceDispatch]()
	for _, name := range []string{
		"WaitSemaphores",
		"WaitSemaphoresKHR",
		"SignalSemaphore",
		"SignalSemaphoreKHR",
		"GetSemaphoreCounterValue",
		"GetSemaphoreCounterValueKHR",
		"CmdClearColorImage",
		"GetImageSubresourceLayout",
	} {
		if _, ok := dispatchType.FieldByName(name); !ok {
			t.Errorf("DeviceDispatch is missing %s", name)
		}
	}
}

func TestLoadDeviceDispatchLeavesOptionalExtensionCommandsNil(t *testing.T) {
	resetDispatchForTest()
	defer resetDispatchForTest()

	withOptional := requiredDeviceSymbols()
	withOptional["vkGetMemoryFdKHR"] = 0x9000
	instanceDispatch := &InstanceDispatch{GetDeviceProcAddr: fakeGetDeviceProcAddr(withOptional)}
	dispatch, err := LoadDeviceDispatch(instanceDispatch, Device(2))
	if err != nil {
		t.Fatalf("first LoadDeviceDispatch() error = %v", err)
	}
	if dispatch.GetMemoryFdKHR == nil {
		t.Fatal("optional GetMemoryFdKHR was not loaded when symbol was present")
	}

	instanceDispatch = &InstanceDispatch{GetDeviceProcAddr: fakeGetDeviceProcAddr(requiredDeviceSymbols())}
	dispatch, err = LoadDeviceDispatch(instanceDispatch, Device(3))
	if err != nil {
		t.Fatalf("second LoadDeviceDispatch() error = %v", err)
	}
	if dispatch.Device != Device(3) {
		t.Fatalf("DeviceDispatch.Device = %v, want 3", dispatch.Device)
	}
	if dispatch.DestroyDevice == nil {
		t.Fatal("DestroyDevice was not loaded")
	}
	if dispatch.BindImageMemory2 == nil || dispatch.BindImageMemory2KHR == nil {
		t.Fatal("core/KHR device alias group was not populated from one available symbol")
	}
	assertRendererDeviceDispatch(t, dispatch)
	if dispatch.GetMemoryFdKHR != nil {
		t.Fatal("optional GetMemoryFdKHR loaded despite missing symbol")
	}
	if dispatch.GetImageDrmFormatModifierPropertiesEXT != nil {
		t.Fatal("optional GetImageDrmFormatModifierPropertiesEXT loaded despite missing symbol")
	}
}

func TestLoadDeviceDispatchFallsBackToTimelineSemaphoreKHRAliases(t *testing.T) {
	resetDispatchForTest()
	defer resetDispatchForTest()

	symbols := requiredDeviceSymbols()
	symbols["vkGetSemaphoreCounterValueKHR"] = 0xa001
	symbols["vkWaitSemaphoresKHR"] = 0xa002
	symbols["vkSignalSemaphoreKHR"] = 0xa003
	instanceDispatch := &InstanceDispatch{GetDeviceProcAddr: fakeGetDeviceProcAddr(symbols)}
	dispatch, err := LoadDeviceDispatch(instanceDispatch, Device(4))
	if err != nil {
		t.Fatalf("LoadDeviceDispatch() error = %v", err)
	}
	for _, check := range []struct {
		name string
		fn   any
	}{
		{"GetSemaphoreCounterValue", dispatch.GetSemaphoreCounterValue},
		{"GetSemaphoreCounterValueKHR", dispatch.GetSemaphoreCounterValueKHR},
		{"WaitSemaphores", dispatch.WaitSemaphores},
		{"WaitSemaphoresKHR", dispatch.WaitSemaphoresKHR},
		{"SignalSemaphore", dispatch.SignalSemaphore},
		{"SignalSemaphoreKHR", dispatch.SignalSemaphoreKHR},
	} {
		if check.fn == nil || reflect.ValueOf(check.fn).IsNil() {
			t.Fatalf("timeline semaphore %s alias was not loaded from the KHR symbol", check.name)
		}
	}
}

func TestLoadDeviceDispatchFallsBackToDynamicRenderingKHRAliases(t *testing.T) {
	resetDispatchForTest()
	defer resetDispatchForTest()

	symbols := requiredDeviceSymbolsExcept("vkCmdBeginRendering", "vkCmdEndRendering")
	symbols["vkCmdBeginRenderingKHR"] = 0xa001
	symbols["vkCmdEndRenderingKHR"] = 0xa002
	instanceDispatch := &InstanceDispatch{GetDeviceProcAddr: fakeGetDeviceProcAddr(symbols)}
	dispatch, err := LoadDeviceDispatch(instanceDispatch, Device(4))
	if err != nil {
		t.Fatalf("LoadDeviceDispatch() error = %v", err)
	}
	if dispatch.CmdBeginRendering == nil || dispatch.CmdBeginRenderingKHR == nil {
		t.Fatal("dynamic rendering begin core/KHR alias group was not populated from KHR symbol")
	}
	if dispatch.CmdEndRendering == nil || dispatch.CmdEndRenderingKHR == nil {
		t.Fatal("dynamic rendering end core/KHR alias group was not populated from KHR symbol")
	}
}

func assertRendererDeviceDispatch(t *testing.T, dispatch *DeviceDispatch) {
	t.Helper()
	checks := []struct {
		name string
		fn   any
	}{
		{"GetImageSubresourceLayout", dispatch.GetImageSubresourceLayout},
		{"CreateBuffer", dispatch.CreateBuffer},
		{"DestroyBuffer", dispatch.DestroyBuffer},
		{"GetBufferMemoryRequirements", dispatch.GetBufferMemoryRequirements},
		{"BindBufferMemory", dispatch.BindBufferMemory},
		{"MapMemory", dispatch.MapMemory},
		{"UnmapMemory", dispatch.UnmapMemory},
		{"FlushMappedMemoryRanges", dispatch.FlushMappedMemoryRanges},
		{"InvalidateMappedMemoryRanges", dispatch.InvalidateMappedMemoryRanges},
		{"CmdCopyBufferToImage", dispatch.CmdCopyBufferToImage},
		{"CmdCopyImageToBuffer", dispatch.CmdCopyImageToBuffer},
		{"CmdClearColorImage", dispatch.CmdClearColorImage},
		{"CmdPipelineBarrier", dispatch.CmdPipelineBarrier},
		{"CreateGraphicsPipelines", dispatch.CreateGraphicsPipelines},
		{"DestroyPipeline", dispatch.DestroyPipeline},
		{"CmdBindPipeline", dispatch.CmdBindPipeline},
		{"CmdBindDescriptorSets", dispatch.CmdBindDescriptorSets},
		{"CmdBindVertexBuffers", dispatch.CmdBindVertexBuffers},
		{"CmdBindIndexBuffer", dispatch.CmdBindIndexBuffer},
		{"CmdDraw", dispatch.CmdDraw},
		{"CmdDrawIndexed", dispatch.CmdDrawIndexed},
		{"CmdBeginRendering", dispatch.CmdBeginRendering},
		{"CmdBeginRenderingKHR", dispatch.CmdBeginRenderingKHR},
		{"CmdEndRendering", dispatch.CmdEndRendering},
		{"CmdEndRenderingKHR", dispatch.CmdEndRenderingKHR},
	}
	for _, check := range checks {
		if check.fn == nil || reflect.ValueOf(check.fn).IsNil() {
			t.Fatalf("%s was not loaded", check.name)
		}
	}
}

func fakeGetInstanceProcAddr(symbols map[string]uintptr) func(Instance, *byte) PFN_vkVoidFunction {
	return func(instance Instance, name *byte) PFN_vkVoidFunction {
		return PFN_vkVoidFunction(symbols[cStringFromPtr(name)])
	}
}

func fakeGetDeviceProcAddr(symbols map[string]uintptr) func(Device, *byte) PFN_vkVoidFunction {
	return func(device Device, name *byte) PFN_vkVoidFunction {
		return PFN_vkVoidFunction(symbols[cStringFromPtr(name)])
	}
}

func cStringFromPtr(p *byte) string {
	if p == nil {
		return ""
	}
	var out []byte
	for offset := uintptr(0); ; offset++ {
		b := *(*byte)(unsafe.Add(unsafe.Pointer(p), offset))
		if b == 0 {
			return string(out)
		}
		out = append(out, b)
	}
}

func requiredInstanceSymbols() map[string]uintptr {
	return symbolMap([]string{
		"vkDestroyInstance",
		"vkEnumeratePhysicalDevices",
		"vkGetDeviceProcAddr",
		"vkGetPhysicalDeviceProperties",
		"vkGetPhysicalDeviceMemoryProperties",
		"vkGetPhysicalDeviceMemoryProperties2",
		"vkGetPhysicalDeviceQueueFamilyProperties",
		"vkCreateDevice",
		"vkEnumerateDeviceExtensionProperties",
		"vkGetPhysicalDeviceFeatures2",
		"vkGetPhysicalDeviceProperties2",
		"vkGetPhysicalDeviceFormatProperties2",
		"vkGetPhysicalDeviceQueueFamilyProperties2",
	})
}

func requiredInstanceSymbolsExcept(excluded string) map[string]uintptr {
	symbols := requiredInstanceSymbols()
	delete(symbols, excluded)
	if excluded == "vkGetPhysicalDeviceMemoryProperties2" {
		symbols["vkGetPhysicalDeviceMemoryProperties2KHR"] = 0x8fff
	}
	return symbols
}

func requiredDeviceSymbolsExcept(excluded ...string) map[string]uintptr {
	symbols := requiredDeviceSymbols()
	for _, name := range excluded {
		delete(symbols, name)
	}
	return symbols
}

func requiredDeviceSymbols() map[string]uintptr {
	return symbolMap([]string{
		"vkDestroyDevice",
		"vkGetDeviceQueue",
		"vkQueueSubmit",
		"vkQueueWaitIdle",
		"vkDeviceWaitIdle",
		"vkAllocateMemory",
		"vkFreeMemory",
		"vkGetImageMemoryRequirements",
		"vkGetImageSubresourceLayout",
		"vkBindImageMemory",
		"vkCreateFence",
		"vkDestroyFence",
		"vkResetFences",
		"vkWaitForFences",
		"vkCreateSemaphore",
		"vkDestroySemaphore",
		"vkCreateImage",
		"vkDestroyImage",
		"vkCreateImageView",
		"vkDestroyImageView",
		"vkCreateShaderModule",
		"vkDestroyShaderModule",
		"vkCreatePipelineLayout",
		"vkDestroyPipelineLayout",
		"vkCreateSampler",
		"vkDestroySampler",
		"vkCreateDescriptorSetLayout",
		"vkDestroyDescriptorSetLayout",
		"vkCreateDescriptorPool",
		"vkDestroyDescriptorPool",
		"vkAllocateDescriptorSets",
		"vkUpdateDescriptorSets",
		"vkCreateCommandPool",
		"vkDestroyCommandPool",
		"vkAllocateCommandBuffers",
		"vkFreeCommandBuffers",
		"vkBeginCommandBuffer",
		"vkEndCommandBuffer",
		"vkResetCommandBuffer",
		"vkBindImageMemory2",
		"vkGetImageMemoryRequirements2",
		"vkCreateBuffer",
		"vkDestroyBuffer",
		"vkGetBufferMemoryRequirements",
		"vkBindBufferMemory",
		"vkMapMemory",
		"vkUnmapMemory",
		"vkFlushMappedMemoryRanges",
		"vkInvalidateMappedMemoryRanges",
		"vkCmdCopyBufferToImage",
		"vkCmdCopyImageToBuffer",
		"vkCmdClearColorImage",
		"vkCmdPipelineBarrier",
		"vkCreateGraphicsPipelines",
		"vkDestroyPipeline",
		"vkCmdBindPipeline",
		"vkCmdBindDescriptorSets",
		"vkCmdBindVertexBuffers",
		"vkCmdBindIndexBuffer",
		"vkCmdDraw",
		"vkCmdDrawIndexed",
		"vkCmdBeginRendering",
		"vkCmdEndRendering",
	})
}

func symbolMap(names []string) map[string]uintptr {
	symbols := make(map[string]uintptr, len(names))
	for i, name := range names {
		symbols[name] = uintptr(0x1000 + i + 1)
	}
	return symbols
}

func resetDispatchForTest() {
	resetInitForTest()
	clearCommandPointers(globalCommandPointers())
	clearCommandPointers(instanceCommandPointers())
	clearCommandPointers(deviceCommandPointers())
	globalDispatch = GlobalDispatch{}
}

package vulkan

import (
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
	if dispatch.GetPhysicalDeviceProperties2 == nil || dispatch.GetPhysicalDeviceProperties2KHR == nil {
		t.Fatal("core/KHR alias group was not populated from one available symbol")
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
	if dispatch.GetMemoryFdKHR != nil {
		t.Fatal("optional GetMemoryFdKHR loaded despite missing symbol")
	}
	if dispatch.GetImageDrmFormatModifierPropertiesEXT != nil {
		t.Fatal("optional GetImageDrmFormatModifierPropertiesEXT loaded despite missing symbol")
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
		"vkGetPhysicalDeviceQueueFamilyProperties",
		"vkCreateDevice",
		"vkEnumerateDeviceExtensionProperties",
		"vkGetPhysicalDeviceFeatures2",
		"vkGetPhysicalDeviceProperties2",
		"vkGetPhysicalDeviceQueueFamilyProperties2",
	})
}

func requiredInstanceSymbolsExcept(excluded string) map[string]uintptr {
	symbols := requiredInstanceSymbols()
	delete(symbols, excluded)
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

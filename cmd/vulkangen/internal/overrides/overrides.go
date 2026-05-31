package overrides

import "github.com/bnema/purego-vulkan/cmd/vulkangen/internal/model"

var InitialCommands = []string{
	"vkGetInstanceProcAddr",
	"vkCreateInstance",
	"vkDestroyInstance",
	"vkEnumerateInstanceVersion",
	"vkEnumerateInstanceExtensionProperties",
	"vkEnumerateInstanceLayerProperties",
	"vkEnumeratePhysicalDevices",
	"vkGetPhysicalDeviceProperties",
	"vkGetPhysicalDeviceProperties2",
	"vkGetPhysicalDeviceFeatures2",
	"vkGetPhysicalDeviceQueueFamilyProperties",
	"vkGetPhysicalDeviceQueueFamilyProperties2",
	"vkEnumerateDeviceExtensionProperties",
	"vkCreateDevice",
	"vkGetDeviceProcAddr",
	"vkDestroyDevice",
	"vkGetDeviceQueue",
	"vkDeviceWaitIdle",
	"vkQueueWaitIdle",
	"vkQueueSubmit",
	"vkCreateCommandPool",
	"vkDestroyCommandPool",
	"vkAllocateCommandBuffers",
	"vkFreeCommandBuffers",
	"vkBeginCommandBuffer",
	"vkEndCommandBuffer",
	"vkResetCommandBuffer",
	"vkCreateFence",
	"vkDestroyFence",
	"vkWaitForFences",
	"vkResetFences",
	"vkCreateSemaphore",
	"vkDestroySemaphore",
	"vkCreateImage",
	"vkDestroyImage",
	"vkGetImageMemoryRequirements",
	"vkGetImageMemoryRequirements2",
	"vkAllocateMemory",
	"vkFreeMemory",
	"vkBindImageMemory",
	"vkBindImageMemory2",
	"vkCreateImageView",
	"vkDestroyImageView",
	"vkCreateSampler",
	"vkDestroySampler",
	"vkCreateShaderModule",
	"vkDestroyShaderModule",
	"vkCreatePipelineLayout",
	"vkDestroyPipelineLayout",
	"vkCreateDescriptorSetLayout",
	"vkDestroyDescriptorSetLayout",
	"vkCreateDescriptorPool",
	"vkDestroyDescriptorPool",
	"vkAllocateDescriptorSets",
	"vkUpdateDescriptorSets",

	// Renderer-ready low-level surface: buffer upload, transfer, graphics pipeline,
	// draw, and dynamic rendering. Pipeline cache management commands are omitted;
	// VkPipelineCache is still selected through vkCreateGraphicsPipelines.
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
}

var RequiredExtensions = []string{
	"VK_KHR_get_physical_device_properties2",
	"VK_KHR_bind_memory2",
	"VK_KHR_get_memory_requirements2",
	"VK_KHR_external_memory",
	"VK_KHR_external_memory_fd",
	"VK_EXT_external_memory_dma_buf",
	"VK_EXT_image_drm_format_modifier",
	"VK_EXT_physical_device_drm",
	"VK_KHR_external_semaphore",
	"VK_KHR_external_semaphore_fd",
	"VK_KHR_synchronization2",
	"VK_EXT_queue_family_foreign",
	"VK_KHR_dynamic_rendering",
}

var CommandOverrides = map[string]model.CommandOverride{
	"vkEnumerateInstanceVersion": {Optional: true},
	"vkGetDeviceProcAddr":        {Dispatch: model.DispatchInstance},

	"vkGetMemoryFdKHR":                         {Dispatch: model.DispatchDevice, Optional: true},
	"vkGetMemoryFdPropertiesKHR":               {Dispatch: model.DispatchDevice, Optional: true},
	"vkImportSemaphoreFdKHR":                   {Dispatch: model.DispatchDevice, Optional: true},
	"vkGetSemaphoreFdKHR":                      {Dispatch: model.DispatchDevice, Optional: true},
	"vkGetImageDrmFormatModifierPropertiesEXT": {Dispatch: model.DispatchDevice, Optional: true},
}

func DefaultSelection() model.SelectionConfig {
	return model.SelectionConfig{
		Commands:         append([]string(nil), InitialCommands...),
		Extensions:       append([]string(nil), RequiredExtensions...),
		CoreVersions:     []string{"VK_VERSION_1_0", "VK_VERSION_1_1", "VK_VERSION_1_2"},
		CommandOverrides: cloneCommandOverrides(),
	}
}

func cloneCommandOverrides() map[string]model.CommandOverride {
	out := make(map[string]model.CommandOverride, len(CommandOverrides))
	for name, entry := range CommandOverrides {
		out[name] = entry
	}
	return out
}

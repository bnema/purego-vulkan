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
}

var RequiredExtensions = []string{
	"VK_KHR_get_physical_device_properties2",
	"VK_KHR_external_memory",
	"VK_KHR_external_memory_fd",
	"VK_EXT_external_memory_dma_buf",
	"VK_EXT_image_drm_format_modifier",
	"VK_EXT_physical_device_drm",
	"VK_KHR_external_semaphore",
	"VK_KHR_external_semaphore_fd",
	"VK_KHR_synchronization2",
	"VK_EXT_queue_family_foreign",
}

var CommandOverrides = map[string]model.CommandOverride{
	"vkGetDeviceProcAddr": {Dispatch: model.DispatchInstance},

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

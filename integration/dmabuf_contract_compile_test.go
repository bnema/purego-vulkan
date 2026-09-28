package integration_test

import "github.com/bnema/purego-vulkan/vulkan"

// Compile-time consumer coverage: keep the names and method sets used by a DMA-BUF presentation path
// visible to an external package without initializing a Vulkan loader.
var (
	_ = vulkan.Init
	_ = vulkan.Global
	_ = vulkan.LoadInstanceDispatch
	_ = vulkan.LoadDeviceDispatch
	_ = vulkan.Check
	_ = (*vulkan.GlobalDispatch).CreateInstance
	_ = (*vulkan.GlobalDispatch).EnumerateInstanceLayerProperties
	_ = (*vulkan.InstanceDispatch).DestroyInstance
	_ = (*vulkan.InstanceDispatch).EnumeratePhysicalDevices
	_ = (*vulkan.InstanceDispatch).EnumerateDeviceExtensionProperties
	_ = (*vulkan.InstanceDispatch).GetPhysicalDeviceProperties2
	_ = (*vulkan.InstanceDispatch).GetPhysicalDeviceFeatures2
	_ = (*vulkan.InstanceDispatch).GetPhysicalDeviceQueueFamilyProperties
	_ = (*vulkan.InstanceDispatch).GetPhysicalDeviceMemoryProperties
	_ = (*vulkan.InstanceDispatch).GetPhysicalDeviceFormatProperties2
	_ = (*vulkan.InstanceDispatch).GetPhysicalDeviceImageFormatProperties2
	_ = (*vulkan.InstanceDispatch).GetPhysicalDeviceExternalSemaphorePropertiesKHR
	_ = (*vulkan.InstanceDispatch).CreateDevice
	_ = (*vulkan.InstanceDispatch).CreateDebugUtilsMessengerEXT
	_ = (*vulkan.InstanceDispatch).DestroyDebugUtilsMessengerEXT
	_ = (*vulkan.DeviceDispatch).GetDeviceQueue
	_ = (*vulkan.DeviceDispatch).DestroyDevice
	_ = (*vulkan.DeviceDispatch).DeviceWaitIdle
	_ = (*vulkan.DeviceDispatch).QueueWaitIdle
	_ = (*vulkan.DeviceDispatch).CreateImage
	_ = (*vulkan.DeviceDispatch).DestroyImage
	_ = (*vulkan.DeviceDispatch).GetImageMemoryRequirements2
	_ = (*vulkan.DeviceDispatch).BindImageMemory2
	_ = (*vulkan.DeviceDispatch).CreateImageView
	_ = (*vulkan.DeviceDispatch).DestroyImageView
	_ = (*vulkan.DeviceDispatch).AllocateMemory
	_ = (*vulkan.DeviceDispatch).FreeMemory
	_ = (*vulkan.DeviceDispatch).GetMemoryFdKHR
	_ = (*vulkan.DeviceDispatch).GetMemoryFdPropertiesKHR
	_ = (*vulkan.DeviceDispatch).GetImageDrmFormatModifierPropertiesEXT
	_ = (*vulkan.DeviceDispatch).CreateSemaphore
	_ = (*vulkan.DeviceDispatch).DestroySemaphore
	_ = (*vulkan.DeviceDispatch).GetSemaphoreCounterValue
	_ = (*vulkan.DeviceDispatch).WaitSemaphores
	_ = (*vulkan.DeviceDispatch).SignalSemaphore
	_ = (*vulkan.DeviceDispatch).GetSemaphoreFdKHR
	_ = (*vulkan.DeviceDispatch).ImportSemaphoreFdKHR
	_ = (*vulkan.DeviceDispatch).CreateBuffer
	_ = (*vulkan.DeviceDispatch).DestroyBuffer
	_ = (*vulkan.DeviceDispatch).BindBufferMemory
	_ = (*vulkan.DeviceDispatch).MapMemory
	_ = (*vulkan.DeviceDispatch).UnmapMemory
	_ = (*vulkan.DeviceDispatch).CreateCommandPool
	_ = (*vulkan.DeviceDispatch).DestroyCommandPool
	_ = (*vulkan.DeviceDispatch).AllocateCommandBuffers
	_ = (*vulkan.DeviceDispatch).FreeCommandBuffers
	_ = (*vulkan.DeviceDispatch).BeginCommandBuffer
	_ = (*vulkan.DeviceDispatch).EndCommandBuffer
	_ = (*vulkan.DeviceDispatch).CmdPipelineBarrier2
	_ = (*vulkan.DeviceDispatch).CmdPipelineBarrier2KHR
	_ = (*vulkan.DeviceDispatch).CmdCopyImageToBuffer
	_ = (*vulkan.DeviceDispatch).CmdCopyBufferToImage
	_ = (*vulkan.DeviceDispatch).CmdBeginRendering
	_ = (*vulkan.DeviceDispatch).CmdEndRendering
	_ = (*vulkan.DeviceDispatch).CmdBindPipeline
	_ = (*vulkan.DeviceDispatch).CmdBindDescriptorSets
	_ = (*vulkan.DeviceDispatch).CmdPushConstants
	_ = (*vulkan.DeviceDispatch).CmdDraw
	_ = (*vulkan.DeviceDispatch).QueueSubmit2
	_ = (*vulkan.DeviceDispatch).QueueSubmit2KHR
	_ = (*vulkan.DeviceDispatch).CreateShaderModule
	_ = (*vulkan.DeviceDispatch).DestroyShaderModule
	_ = (*vulkan.DeviceDispatch).CreateGraphicsPipelines
	_ = (*vulkan.DeviceDispatch).DestroyPipeline
	_ = (*vulkan.DeviceDispatch).CreatePipelineLayout
	_ = (*vulkan.DeviceDispatch).DestroyPipelineLayout
	_ = (*vulkan.DeviceDispatch).CreateDescriptorSetLayout
	_ = (*vulkan.DeviceDispatch).DestroyDescriptorSetLayout
	_ = (*vulkan.DeviceDispatch).CreateDescriptorPool
	_ = (*vulkan.DeviceDispatch).DestroyDescriptorPool
	_ = (*vulkan.DeviceDispatch).AllocateDescriptorSets
	_ = (*vulkan.DeviceDispatch).UpdateDescriptorSets
	_ = (*vulkan.DeviceDispatch).CreateSampler
	_ = (*vulkan.DeviceDispatch).DestroySampler
	_ vulkan.PhysicalDeviceDrmPropertiesEXT
	_ vulkan.PhysicalDeviceProperties2
	_ vulkan.PhysicalDeviceFeatures2
	_ vulkan.PhysicalDeviceTimelineSemaphoreFeatures
	_ vulkan.PhysicalDeviceSynchronization2Features
	_ vulkan.PhysicalDeviceDynamicRenderingFeatures
	_ vulkan.FormatProperties2
	_ vulkan.DrmFormatModifierPropertiesListEXT
	_ vulkan.DrmFormatModifierPropertiesList2EXT
	_ vulkan.PhysicalDeviceImageDrmFormatModifierInfoEXT
	_ vulkan.ImageDrmFormatModifierListCreateInfoEXT
	_ vulkan.ImageDrmFormatModifierExplicitCreateInfoEXT
	_ vulkan.ImageDrmFormatModifierPropertiesEXT
	_ vulkan.ExternalImageFormatProperties
	_ vulkan.ExternalMemoryImageCreateInfo
	_ vulkan.ExportMemoryAllocateInfo
	_ vulkan.MemoryDedicatedAllocateInfo
	_ vulkan.MemoryGetFdInfoKHR
	_ vulkan.MemoryFdPropertiesKHR
	_ vulkan.ExportSemaphoreCreateInfo
	_ vulkan.SemaphoreTypeCreateInfo
	_ vulkan.SemaphoreGetFdInfoKHR
	_ vulkan.ImportSemaphoreFdInfoKHR
	_ vulkan.SemaphoreSubmitInfo
	_ vulkan.SubmitInfo2
	_ vulkan.DependencyInfo
	_ vulkan.ImageMemoryBarrier2
	_ vulkan.RenderingInfo
	_ vulkan.RenderingAttachmentInfo
	_ vulkan.DebugUtilsMessengerCreateInfoEXT
	_ vulkan.BufferImageCopy
	_ vulkan.PushConstantRange
	_ = vulkan.FormatB8g8r8a8Unorm
	_ = vulkan.ExternalMemoryHandleTypeDMABUFBitEXT
	_ = vulkan.ExternalSemaphoreHandleTypeOpaqueFDBit
	_ = vulkan.ExternalSemaphoreHandleTypeSyncFDBit
	_ = vulkan.ImageTilingDRMFormatModifierEXT
	_ = vulkan.SemaphoreTypeTimeline
)

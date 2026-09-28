package vulkan

import (
	"testing"
	"unsafe"
)

func TestGeneratedWSILayouts(t *testing.T) {
	require64BitLayout(t)
	tests := []struct {
		name      string
		size      uintptr
		alignment uintptr
		wantSize  uintptr
		wantAlign uintptr
	}{
		{
			name:      "WaylandSurfaceCreateInfoKHR",
			size:      unsafe.Sizeof(WaylandSurfaceCreateInfoKHR{}),
			alignment: unsafe.Alignof(WaylandSurfaceCreateInfoKHR{}),
			wantSize:  40,
			wantAlign: 8,
		},
		{
			name:      "XlibSurfaceCreateInfoKHR",
			size:      unsafe.Sizeof(XlibSurfaceCreateInfoKHR{}),
			alignment: unsafe.Alignof(XlibSurfaceCreateInfoKHR{}),
			wantSize:  40,
			wantAlign: 8,
		},
		{
			name:      "XcbSurfaceCreateInfoKHR",
			size:      unsafe.Sizeof(XcbSurfaceCreateInfoKHR{}),
			alignment: unsafe.Alignof(XcbSurfaceCreateInfoKHR{}),
			wantSize:  40,
			wantAlign: 8,
		},
		{
			name:      "SwapchainCreateInfoKHR",
			size:      unsafe.Sizeof(SwapchainCreateInfoKHR{}),
			alignment: unsafe.Alignof(SwapchainCreateInfoKHR{}),
			wantSize:  104,
			wantAlign: 8,
		},
		{
			name:      "PresentInfoKHR",
			size:      unsafe.Sizeof(PresentInfoKHR{}),
			alignment: unsafe.Alignof(PresentInfoKHR{}),
			wantSize:  64,
			wantAlign: 8,
		},
		{
			name:      "DeviceGroupPresentCapabilitiesKHR",
			size:      unsafe.Sizeof(DeviceGroupPresentCapabilitiesKHR{}),
			alignment: unsafe.Alignof(DeviceGroupPresentCapabilitiesKHR{}),
			wantSize:  152,
			wantAlign: 8,
		},
		{
			name:      "ImageSwapchainCreateInfoKHR",
			size:      unsafe.Sizeof(ImageSwapchainCreateInfoKHR{}),
			alignment: unsafe.Alignof(ImageSwapchainCreateInfoKHR{}),
			wantSize:  24,
			wantAlign: 8,
		},
		{
			name:      "BindImageMemorySwapchainInfoKHR",
			size:      unsafe.Sizeof(BindImageMemorySwapchainInfoKHR{}),
			alignment: unsafe.Alignof(BindImageMemorySwapchainInfoKHR{}),
			wantSize:  32,
			wantAlign: 8,
		},
		{
			name:      "AcquireNextImageInfoKHR",
			size:      unsafe.Sizeof(AcquireNextImageInfoKHR{}),
			alignment: unsafe.Alignof(AcquireNextImageInfoKHR{}),
			wantSize:  56,
			wantAlign: 8,
		},
		{
			name:      "DeviceGroupPresentInfoKHR",
			size:      unsafe.Sizeof(DeviceGroupPresentInfoKHR{}),
			alignment: unsafe.Alignof(DeviceGroupPresentInfoKHR{}),
			wantSize:  40,
			wantAlign: 8,
		},
		{
			name:      "DeviceGroupSwapchainCreateInfoKHR",
			size:      unsafe.Sizeof(DeviceGroupSwapchainCreateInfoKHR{}),
			alignment: unsafe.Alignof(DeviceGroupSwapchainCreateInfoKHR{}),
			wantSize:  24,
			wantAlign: 8,
		},
	}
	for _, tt := range tests {
		if tt.size != tt.wantSize || tt.alignment != tt.wantAlign {
			t.Fatalf("%s layout = size %d align %d, want size %d align %d", tt.name, tt.size, tt.alignment, tt.wantSize, tt.wantAlign)
		}
	}
}

// TestGeneratedDMABUFPresentationLayouts checks the 64-bit C ABI layouts of the
// DRM, format-modifier and debug-utils structures in registry/vk.xml.
func TestGeneratedDMABUFPresentationLayouts(t *testing.T) {
	require64BitLayout(t)
	for _, tc := range []struct {
		name                             string
		size, align, wantSize, wantAlign uintptr
	}{
		{"PhysicalDeviceDrmPropertiesEXT", unsafe.Sizeof(PhysicalDeviceDrmPropertiesEXT{}), unsafe.Alignof(PhysicalDeviceDrmPropertiesEXT{}), 56, 8},
		{"DrmFormatModifierPropertiesList2EXT", unsafe.Sizeof(DrmFormatModifierPropertiesList2EXT{}), unsafe.Alignof(DrmFormatModifierPropertiesList2EXT{}), 32, 8},
		{"DrmFormatModifierProperties2EXT", unsafe.Sizeof(DrmFormatModifierProperties2EXT{}), unsafe.Alignof(DrmFormatModifierProperties2EXT{}), 24, 8},
		{"DebugUtilsMessengerCreateInfoEXT", unsafe.Sizeof(DebugUtilsMessengerCreateInfoEXT{}), unsafe.Alignof(DebugUtilsMessengerCreateInfoEXT{}), 48, 8},
	} {
		if tc.size != tc.wantSize || tc.align != tc.wantAlign {
			t.Errorf("%s layout size=%d align=%d, want size=%d align=%d", tc.name, tc.size, tc.align, tc.wantSize, tc.wantAlign)
		}
	}
}

// Offsets follow the x86_64 Vulkan C ABI: pointers and 64-bit values align to 8,
// while VkBool32, enums and flag bits align to 4.
func TestGeneratedDMABUFPresentationFieldOffsets(t *testing.T) {
	require64BitLayout(t)
	for _, tc := range []struct {
		name      string
		got, want uintptr
	}{
		{"PhysicalDeviceDrmPropertiesEXT.PrimaryMajor", unsafe.Offsetof(PhysicalDeviceDrmPropertiesEXT{}.PrimaryMajor), 24},
		{"PhysicalDeviceDrmPropertiesEXT.PrimaryMinor", unsafe.Offsetof(PhysicalDeviceDrmPropertiesEXT{}.PrimaryMinor), 32},
		{"PhysicalDeviceDrmPropertiesEXT.RenderMajor", unsafe.Offsetof(PhysicalDeviceDrmPropertiesEXT{}.RenderMajor), 40},
		{"PhysicalDeviceDrmPropertiesEXT.RenderMinor", unsafe.Offsetof(PhysicalDeviceDrmPropertiesEXT{}.RenderMinor), 48},
		{"DrmFormatModifierPropertiesEXT.DrmFormatModifier", unsafe.Offsetof(DrmFormatModifierPropertiesEXT{}.DrmFormatModifier), 0},
		{"DrmFormatModifierPropertiesEXT.DrmFormatModifierPlaneCount", unsafe.Offsetof(DrmFormatModifierPropertiesEXT{}.DrmFormatModifierPlaneCount), 8},
		{"DrmFormatModifierPropertiesEXT.DrmFormatModifierTilingFeatures", unsafe.Offsetof(DrmFormatModifierPropertiesEXT{}.DrmFormatModifierTilingFeatures), 12},
		{"ImageDrmFormatModifierExplicitCreateInfoEXT.DrmFormatModifier", unsafe.Offsetof(ImageDrmFormatModifierExplicitCreateInfoEXT{}.DrmFormatModifier), 16},
		{"ImageDrmFormatModifierExplicitCreateInfoEXT.DrmFormatModifierPlaneCount", unsafe.Offsetof(ImageDrmFormatModifierExplicitCreateInfoEXT{}.DrmFormatModifierPlaneCount), 24},
		{"ImageDrmFormatModifierExplicitCreateInfoEXT.PlaneLayouts", unsafe.Offsetof(ImageDrmFormatModifierExplicitCreateInfoEXT{}.PlaneLayouts), 32},
		{"SubresourceLayout.Offset", unsafe.Offsetof(SubresourceLayout{}.Offset), 0},
		{"SubresourceLayout.Size", unsafe.Offsetof(SubresourceLayout{}.Size), 8},
		{"SubresourceLayout.RowPitch", unsafe.Offsetof(SubresourceLayout{}.RowPitch), 16},
		{"SubresourceLayout.ArrayPitch", unsafe.Offsetof(SubresourceLayout{}.ArrayPitch), 24},
		{"SubresourceLayout.DepthPitch", unsafe.Offsetof(SubresourceLayout{}.DepthPitch), 32},
		{"SemaphoreGetFdInfoKHR.Semaphore", unsafe.Offsetof(SemaphoreGetFdInfoKHR{}.Semaphore), 16},
		{"SemaphoreGetFdInfoKHR.HandleType", unsafe.Offsetof(SemaphoreGetFdInfoKHR{}.HandleType), 24},
		{"ImportSemaphoreFdInfoKHR.Semaphore", unsafe.Offsetof(ImportSemaphoreFdInfoKHR{}.Semaphore), 16},
		{"ImportSemaphoreFdInfoKHR.Flags", unsafe.Offsetof(ImportSemaphoreFdInfoKHR{}.Flags), 24},
		{"ImportSemaphoreFdInfoKHR.HandleType", unsafe.Offsetof(ImportSemaphoreFdInfoKHR{}.HandleType), 28},
		{"ImportSemaphoreFdInfoKHR.Fd", unsafe.Offsetof(ImportSemaphoreFdInfoKHR{}.Fd), 32},
		{"MemoryGetFdInfoKHR.Memory", unsafe.Offsetof(MemoryGetFdInfoKHR{}.Memory), 16},
		{"MemoryGetFdInfoKHR.HandleType", unsafe.Offsetof(MemoryGetFdInfoKHR{}.HandleType), 24},
		{"ExportMemoryAllocateInfo.HandleTypes", unsafe.Offsetof(ExportMemoryAllocateInfo{}.HandleTypes), 16},
		{"DebugUtilsMessengerCallbackDataEXT.Flags", unsafe.Offsetof(DebugUtilsMessengerCallbackDataEXT{}.Flags), 16},
		{"DebugUtilsMessengerCallbackDataEXT.MessageIdName", unsafe.Offsetof(DebugUtilsMessengerCallbackDataEXT{}.MessageIdName), 24},
		{"DebugUtilsMessengerCallbackDataEXT.MessageIdNumber", unsafe.Offsetof(DebugUtilsMessengerCallbackDataEXT{}.MessageIdNumber), 32},
		{"DebugUtilsMessengerCallbackDataEXT.Message", unsafe.Offsetof(DebugUtilsMessengerCallbackDataEXT{}.Message), 40},
		{"DebugUtilsMessengerCallbackDataEXT.QueueLabelCount", unsafe.Offsetof(DebugUtilsMessengerCallbackDataEXT{}.QueueLabelCount), 48},
		{"DebugUtilsMessengerCallbackDataEXT.QueueLabels", unsafe.Offsetof(DebugUtilsMessengerCallbackDataEXT{}.QueueLabels), 56},
		{"DebugUtilsMessengerCallbackDataEXT.CmdBufLabelCount", unsafe.Offsetof(DebugUtilsMessengerCallbackDataEXT{}.CmdBufLabelCount), 64},
		{"DebugUtilsMessengerCallbackDataEXT.CmdBufLabels", unsafe.Offsetof(DebugUtilsMessengerCallbackDataEXT{}.CmdBufLabels), 72},
		{"DebugUtilsMessengerCallbackDataEXT.ObjectCount", unsafe.Offsetof(DebugUtilsMessengerCallbackDataEXT{}.ObjectCount), 80},
		{"DebugUtilsMessengerCallbackDataEXT.Objects", unsafe.Offsetof(DebugUtilsMessengerCallbackDataEXT{}.Objects), 88},
		{"TimelineSemaphoreSubmitInfo.WaitSemaphoreValueCount", unsafe.Offsetof(TimelineSemaphoreSubmitInfo{}.WaitSemaphoreValueCount), 16},
		{"TimelineSemaphoreSubmitInfo.WaitSemaphoreValues", unsafe.Offsetof(TimelineSemaphoreSubmitInfo{}.WaitSemaphoreValues), 24},
		{"TimelineSemaphoreSubmitInfo.SignalSemaphoreValueCount", unsafe.Offsetof(TimelineSemaphoreSubmitInfo{}.SignalSemaphoreValueCount), 32},
		{"TimelineSemaphoreSubmitInfo.SignalSemaphoreValues", unsafe.Offsetof(TimelineSemaphoreSubmitInfo{}.SignalSemaphoreValues), 40},
		{"SemaphoreSubmitInfo.Semaphore", unsafe.Offsetof(SemaphoreSubmitInfo{}.Semaphore), 16},
		{"SemaphoreSubmitInfo.Value", unsafe.Offsetof(SemaphoreSubmitInfo{}.Value), 24},
		{"SemaphoreSubmitInfo.StageMask", unsafe.Offsetof(SemaphoreSubmitInfo{}.StageMask), 32},
		{"SemaphoreSubmitInfo.DeviceIndex", unsafe.Offsetof(SemaphoreSubmitInfo{}.DeviceIndex), 40},
	} {
		if tc.got != tc.want {
			t.Errorf("%s offset = %d, want %d", tc.name, tc.got, tc.want)
		}
	}
}

// TestGeneratedTimelineSemaphoreLayouts checks the 64-bit C ABI layouts declared
// for the timeline semaphore structures in registry/vk.xml.
func TestGeneratedTimelineSemaphoreLayouts(t *testing.T) {
	require64BitLayout(t)
	tests := []struct {
		name      string
		size      uintptr
		alignment uintptr
		wantSize  uintptr
		wantAlign uintptr
	}{
		{"PhysicalDeviceTimelineSemaphoreFeatures", unsafe.Sizeof(PhysicalDeviceTimelineSemaphoreFeatures{}), unsafe.Alignof(PhysicalDeviceTimelineSemaphoreFeatures{}), 24, 8},
		{"PhysicalDeviceTimelineSemaphoreProperties", unsafe.Sizeof(PhysicalDeviceTimelineSemaphoreProperties{}), unsafe.Alignof(PhysicalDeviceTimelineSemaphoreProperties{}), 24, 8},
		{"SemaphoreTypeCreateInfo", unsafe.Sizeof(SemaphoreTypeCreateInfo{}), unsafe.Alignof(SemaphoreTypeCreateInfo{}), 32, 8},
		{"TimelineSemaphoreSubmitInfo", unsafe.Sizeof(TimelineSemaphoreSubmitInfo{}), unsafe.Alignof(TimelineSemaphoreSubmitInfo{}), 48, 8},
		{"SemaphoreWaitInfo", unsafe.Sizeof(SemaphoreWaitInfo{}), unsafe.Alignof(SemaphoreWaitInfo{}), 40, 8},
		{"SemaphoreSignalInfo", unsafe.Sizeof(SemaphoreSignalInfo{}), unsafe.Alignof(SemaphoreSignalInfo{}), 32, 8},
	}
	for _, tt := range tests {
		if tt.size != tt.wantSize || tt.alignment != tt.wantAlign {
			t.Fatalf("%s layout = size %d align %d, want size %d align %d", tt.name, tt.size, tt.alignment, tt.wantSize, tt.wantAlign)
		}
	}
}

func TestGeneratedDedicatedAllocationKHRAliases(t *testing.T) {
	require64BitLayout(t)

	var requirements MemoryDedicatedRequirementsKHR
	requirements.PrefersDedicatedAllocation = 1
	requirements.RequiresDedicatedAllocation = 1
	var allocateInfo MemoryDedicatedAllocateInfoKHR
	allocateInfo.Image = Image(1)
	allocateInfo.Buffer = Buffer(1)

	// Passing KHR values to core parameters requires aliases, rather than duplicate types.
	requireCoreDedicatedRequirements(requirements)
	requireCoreDedicatedAllocateInfo(allocateInfo)

	if got, want := unsafe.Sizeof(requirements), unsafe.Sizeof(MemoryDedicatedRequirements{}); got != want {
		t.Fatalf("MemoryDedicatedRequirementsKHR size = %d, want core size %d", got, want)
	}
	if got, want := unsafe.Alignof(requirements), unsafe.Alignof(MemoryDedicatedRequirements{}); got != want {
		t.Fatalf("MemoryDedicatedRequirementsKHR alignment = %d, want core alignment %d", got, want)
	}
	if got, want := unsafe.Sizeof(allocateInfo), unsafe.Sizeof(MemoryDedicatedAllocateInfo{}); got != want {
		t.Fatalf("MemoryDedicatedAllocateInfoKHR size = %d, want core size %d", got, want)
	}
	if got, want := unsafe.Alignof(allocateInfo), unsafe.Alignof(MemoryDedicatedAllocateInfo{}); got != want {
		t.Fatalf("MemoryDedicatedAllocateInfoKHR alignment = %d, want core alignment %d", got, want)
	}
}

func requireCoreDedicatedRequirements(MemoryDedicatedRequirements) {}

func requireCoreDedicatedAllocateInfo(MemoryDedicatedAllocateInfo) {}

func TestGeneratedRendererHardeningLayouts(t *testing.T) {
	require64BitLayout(t)
	tests := []struct {
		name      string
		size      uintptr
		alignment uintptr
		wantSize  uintptr
		wantAlign uintptr
	}{
		{"PhysicalDeviceMemoryProperties", unsafe.Sizeof(PhysicalDeviceMemoryProperties{}), unsafe.Alignof(PhysicalDeviceMemoryProperties{}), 520, 8},
		{"PhysicalDeviceMemoryProperties2", unsafe.Sizeof(PhysicalDeviceMemoryProperties2{}), unsafe.Alignof(PhysicalDeviceMemoryProperties2{}), 536, 8},
		{"MappedMemoryRange", unsafe.Sizeof(MappedMemoryRange{}), unsafe.Alignof(MappedMemoryRange{}), 40, 8},
		{"ExternalMemoryImageCreateInfo", unsafe.Sizeof(ExternalMemoryImageCreateInfo{}), unsafe.Alignof(ExternalMemoryImageCreateInfo{}), 24, 8},
		{"ImportMemoryFdInfoKHR", unsafe.Sizeof(ImportMemoryFdInfoKHR{}), unsafe.Alignof(ImportMemoryFdInfoKHR{}), 24, 8},
		{"DrmFormatModifierPropertiesListEXT", unsafe.Sizeof(DrmFormatModifierPropertiesListEXT{}), unsafe.Alignof(DrmFormatModifierPropertiesListEXT{}), 32, 8},
		{"ImageDrmFormatModifierExplicitCreateInfoEXT", unsafe.Sizeof(ImageDrmFormatModifierExplicitCreateInfoEXT{}), unsafe.Alignof(ImageDrmFormatModifierExplicitCreateInfoEXT{}), 40, 8},
		{"MemoryBarrier", unsafe.Sizeof(MemoryBarrier{}), unsafe.Alignof(MemoryBarrier{}), 24, 8},
		{"BufferMemoryBarrier", unsafe.Sizeof(BufferMemoryBarrier{}), unsafe.Alignof(BufferMemoryBarrier{}), 56, 8},
		{"ImageMemoryBarrier", unsafe.Sizeof(ImageMemoryBarrier{}), unsafe.Alignof(ImageMemoryBarrier{}), 72, 8},
		{"MemoryBarrier2", unsafe.Sizeof(MemoryBarrier2{}), unsafe.Alignof(MemoryBarrier2{}), 48, 8},
		{"ImageMemoryBarrier2", unsafe.Sizeof(ImageMemoryBarrier2{}), unsafe.Alignof(ImageMemoryBarrier2{}), 96, 8},
		{"DependencyInfo", unsafe.Sizeof(DependencyInfo{}), unsafe.Alignof(DependencyInfo{}), 64, 8},
		{"PipelineRenderingCreateInfo", unsafe.Sizeof(PipelineRenderingCreateInfo{}), unsafe.Alignof(PipelineRenderingCreateInfo{}), 40, 8},
		{"RenderingInfo", unsafe.Sizeof(RenderingInfo{}), unsafe.Alignof(RenderingInfo{}), 72, 8},
		{"RenderingAttachmentInfo", unsafe.Sizeof(RenderingAttachmentInfo{}), unsafe.Alignof(RenderingAttachmentInfo{}), 72, 8},
	}
	for _, tt := range tests {
		if tt.size != tt.wantSize || tt.alignment != tt.wantAlign {
			t.Fatalf("%s layout = size %d align %d, want size %d align %d", tt.name, tt.size, tt.alignment, tt.wantSize, tt.wantAlign)
		}
	}
}

func require64BitLayout(t *testing.T) {
	t.Helper()
	if unsafe.Sizeof(uintptr(0)) != 8 {
		t.Skip("generated Vulkan ABI layout expectations are for 64-bit targets")
	}
}

func TestGeneratedUnionLayouts(t *testing.T) {
	tests := []struct {
		name      string
		size      uintptr
		alignment uintptr
		wantSize  uintptr
		wantAlign uintptr
	}{
		{
			name:      "ClearColorValue",
			size:      unsafe.Sizeof(ClearColorValue{}),
			alignment: unsafe.Alignof(ClearColorValue{}),
			wantSize:  16,
			wantAlign: 4,
		},
		{
			name:      "ClearValue",
			size:      unsafe.Sizeof(ClearValue{}),
			alignment: unsafe.Alignof(ClearValue{}),
			wantSize:  16,
			wantAlign: 4,
		},
	}
	for _, tt := range tests {
		if tt.size != tt.wantSize || tt.alignment != tt.wantAlign {
			t.Fatalf("%s layout = size %d align %d, want size %d align %d", tt.name, tt.size, tt.alignment, tt.wantSize, tt.wantAlign)
		}
	}
}

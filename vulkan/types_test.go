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

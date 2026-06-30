package vulkan

import (
	"testing"
	"unsafe"
)

func TestGeneratedWSILayouts(t *testing.T) {
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
	}
	for _, tt := range tests {
		if tt.size != tt.wantSize || tt.alignment != tt.wantAlign {
			t.Fatalf("%s layout = size %d align %d, want size %d align %d", tt.name, tt.size, tt.alignment, tt.wantSize, tt.wantAlign)
		}
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

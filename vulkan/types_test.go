package vulkan

import (
	"testing"
	"unsafe"
)

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

package integration_test

import (
	"fmt"
	"os"
	"runtime"
	"testing"
	"unsafe"

	"github.com/bnema/purego-vulkan/vulkan"
)

// Run with VEYA_VULKAN_PROBE=1. Capability rejection is a skip, never a positive pass.
func TestVeyaHardwareExport(t *testing.T) {
	if os.Getenv("VEYA_VULKAN_PROBE") != "1" {
		t.Skip("set VEYA_VULKAN_PROBE=1")
	}
	gd := requireGlobalDispatch(t)
	app := cStringBytes("Vulkan hardware probe")
	ai := vulkan.ApplicationInfo{SType: vulkan.StructureTypeApplicationInfo, ApplicationName: &app[0], ApiVersion: (1 << 22) | (3 << 12)}
	ci := vulkan.InstanceCreateInfo{SType: vulkan.StructureTypeInstanceCreateInfo, ApplicationInfo: &ai}
	var instance vulkan.Instance
	mustResult(t, "CreateInstance", gd.CreateInstance(&ci, nil, &instance))
	id, err := vulkan.LoadInstanceDispatch(instance)
	if err != nil {
		t.Fatalf("BINDING: LoadInstanceDispatch: %v", err)
	}
	defer id.DestroyInstance(instance, nil)
	devices := enumeratePhysicalDevices(t, id, instance)
	if len(devices) == 0 {
		t.Skip("UNSUPPORTED: no Vulkan devices")
	}
	passed := 0
	for _, phys := range devices {
		exts := enumerateDeviceExtensions(t, id, phys)
		if !contains(exts, "VK_EXT_physical_device_drm") {
			t.Log("UNSUPPORTED: VK_EXT_physical_device_drm")
			continue
		}
		props := vulkan.PhysicalDeviceProperties2{SType: vulkan.StructureTypePhysicalDeviceProperties2}
		drm := vulkan.PhysicalDeviceDrmPropertiesEXT{SType: vulkan.StructureTypePhysicalDeviceDRMPropertiesEXT}
		props.Next = unsafe.Pointer(&drm)
		id.GetPhysicalDeviceProperties2(phys, &props)
		label := fixedCString(props.Properties.DeviceName[:])
		t.Logf("device %s primary=%d:%d render=%d:%d", label, drm.PrimaryMajor, drm.PrimaryMinor, drm.RenderMajor, drm.RenderMinor)
		if drm.HasRender == 0 {
			t.Logf("UNSUPPORTED %s: no DRM render node", label)
			continue
		}
		required := []string{"VK_KHR_external_memory_fd", "VK_EXT_external_memory_dma_buf", "VK_EXT_image_drm_format_modifier", "VK_KHR_external_semaphore_fd"}
		missing := ""
		for _, e := range required {
			if !contains(exts, e) {
				missing = e
				break
			}
		}
		if missing != "" {
			t.Logf("UNSUPPORTED %s: %s", label, missing)
			continue
		}
		var count uint32
		id.GetPhysicalDeviceQueueFamilyProperties(phys, &count, nil)
		families := make([]vulkan.QueueFamilyProperties, count)
		id.GetPhysicalDeviceQueueFamilyProperties(phys, &count, &families[0])
		family := -1
		for i, f := range families {
			if f.QueueFlags&vulkan.QueueGraphicsBit != 0 {
				family = i
				break
			}
		}
		if family < 0 {
			t.Logf("UNSUPPORTED %s: graphics queue", label)
			continue
		}
		format := vulkan.FormatProperties2{SType: vulkan.StructureTypeFormatProperties2}
		modifiers := vulkan.DrmFormatModifierPropertiesListEXT{SType: vulkan.StructureTypeDRMFormatModifierPropertiesListEXT}
		format.Next = unsafe.Pointer(&modifiers)
		id.GetPhysicalDeviceFormatProperties2(phys, vulkan.FormatB8g8r8a8Unorm, &format)
		if modifiers.DrmFormatModifierCount == 0 {
			t.Logf("UNSUPPORTED %s: no B8G8R8A8_UNORM modifiers", label)
			continue
		}
		list := make([]vulkan.DrmFormatModifierPropertiesEXT, modifiers.DrmFormatModifierCount)
		modifiers.DrmFormatModifierProperties = &list[0]
		id.GetPhysicalDeviceFormatProperties2(phys, vulkan.FormatB8g8r8a8Unorm, &format)
		priority := float32(1)
		queue := vulkan.DeviceQueueCreateInfo{SType: vulkan.StructureTypeDeviceQueueCreateInfo, QueueFamilyIndex: uint32(family), QueueCount: 1, QueuePriorities: &priority}
		extNames := make([][]byte, len(required))
		extPtrs := make([]*byte, len(required))
		for i, e := range required {
			extNames[i] = cStringBytes(e)
			extPtrs[i] = &extNames[i][0]
		}
		timeline := vulkan.PhysicalDeviceTimelineSemaphoreFeatures{SType: vulkan.StructureTypePhysicalDeviceTimelineSemaphoreFeatures, TimelineSemaphore: 1}
		dc := vulkan.DeviceCreateInfo{SType: vulkan.StructureTypeDeviceCreateInfo, Next: unsafe.Pointer(&timeline), QueueCreateInfoCount: 1, QueueCreateInfos: &queue, EnabledExtensionCount: uint32(len(required)), PpEnabledExtensionNames: &extPtrs[0]}
		var device vulkan.Device
		result := id.CreateDevice(phys, &dc, nil, &device)
		runtime.KeepAlive(extNames)
		if result != vulkan.Success {
			t.Logf("UNSUPPORTED %s: CreateDevice timeline/extensions: %s", label, vulkan.ResultString(result))
			continue
		}
		dd, err := vulkan.LoadDeviceDispatch(id, device)
		if err != nil {
			id.GetDeviceProcAddr(device, nil)
			t.Fatalf("BINDING: LoadDeviceDispatch: %v", err)
		}
		func() {
			defer dd.DestroyDevice(device, nil)
			if !dd.HasGetSemaphoreFdKHR() || !dd.HasGetMemoryFdKHR() || !dd.HasGetImageDrmFormatModifierPropertiesEXT() {
				t.Fatalf("BINDING: advertised extension dispatch missing on %s", label)
			}
			st := vulkan.SemaphoreTypeCreateInfo{SType: vulkan.StructureTypeSemaphoreTypeCreateInfo, SemaphoreType: vulkan.SemaphoreTypeTimeline}
			exportSem := vulkan.ExportSemaphoreCreateInfo{SType: vulkan.StructureTypeExportSemaphoreCreateInfo, Next: unsafe.Pointer(&st), HandleTypes: vulkan.ExternalSemaphoreHandleTypeOpaqueFDBit}
			si := vulkan.SemaphoreCreateInfo{SType: vulkan.StructureTypeSemaphoreCreateInfo, Next: unsafe.Pointer(&exportSem)}
			var sem vulkan.Semaphore
			mustResult(t, "CreateSemaphore timeline", dd.CreateSemaphore(device, &si, nil, &sem))
			defer dd.DestroySemaphore(device, sem, nil)
			getSem := vulkan.SemaphoreGetFdInfoKHR{SType: vulkan.StructureTypeSemaphoreGetFDInfoKHR, Semaphore: sem, HandleType: vulkan.ExternalSemaphoreHandleTypeOpaqueFDBit}
			var semFD int32 = -1
			mustResult(t, "GetSemaphoreFdKHR timeline OPAQUE_FD", dd.GetSemaphoreFdKHR(device, &getSem, &semFD))
			defer os.NewFile(uintptr(semFD), "timeline-opaque-fd").Close()
			t.Logf("%s: timeline OPAQUE_FD export OK (fd=%d; syncobj identity NOT established)", label, semFD)
			exported := false
			for _, mod := range list {
				if mod.DrmFormatModifierPlaneCount != 1 || mod.DrmFormatModifierTilingFeatures&vulkan.FormatFeatureColorAttachmentBit == 0 {
					continue
				}
				modifier := mod.DrmFormatModifier
				mc := vulkan.ImageDrmFormatModifierListCreateInfoEXT{SType: vulkan.StructureTypeImageDRMFormatModifierListCreateInfoEXT, DrmFormatModifierCount: 1, DrmFormatModifiers: &modifier}
				external := vulkan.ExternalMemoryImageCreateInfo{SType: vulkan.StructureTypeExternalMemoryImageCreateInfo, Next: unsafe.Pointer(&mc), HandleTypes: vulkan.ExternalMemoryHandleTypeDMABUFBitEXT}
				imageInfo := vulkan.ImageCreateInfo{SType: vulkan.StructureTypeImageCreateInfo, Next: unsafe.Pointer(&external), ImageType: vulkan.ImageType2d, Format: vulkan.FormatB8g8r8a8Unorm, Extent: vulkan.Extent3D{Width: 64, Height: 64, Depth: 1}, MipLevels: 1, ArrayLayers: 1, Samples: vulkan.SampleCount1Bit, Tiling: vulkan.ImageTilingDRMFormatModifierEXT, Usage: vulkan.ImageUsageColorAttachmentBit, SharingMode: vulkan.SharingModeExclusive}
				var image vulkan.Image
				if dd.CreateImage(device, &imageInfo, nil, &image) != vulkan.Success {
					continue
				}
				func() {
					defer dd.DestroyImage(device, image, nil)
					var req vulkan.MemoryRequirements
					dd.GetImageMemoryRequirements(device, image, &req)
					var mp vulkan.PhysicalDeviceMemoryProperties
					id.GetPhysicalDeviceMemoryProperties(phys, &mp)
					memType := -1
					for i := uint32(0); i < mp.MemoryTypeCount; i++ {
						if req.MemoryTypeBits&(1<<i) != 0 && mp.MemoryTypes[i].PropertyFlags&vulkan.MemoryPropertyDeviceLocalBit != 0 {
							memType = int(i)
							break
						}
					}
					if memType < 0 {
						return
					}
					dedicated := vulkan.MemoryDedicatedAllocateInfo{SType: vulkan.StructureTypeMemoryDedicatedAllocateInfo, Image: image}
					exp := vulkan.ExportMemoryAllocateInfo{SType: vulkan.StructureTypeExportMemoryAllocateInfo, Next: unsafe.Pointer(&dedicated), HandleTypes: vulkan.ExternalMemoryHandleTypeDMABUFBitEXT}
					alloc := vulkan.MemoryAllocateInfo{SType: vulkan.StructureTypeMemoryAllocateInfo, Next: unsafe.Pointer(&exp), AllocationSize: req.Size, MemoryTypeIndex: uint32(memType)}
					var mem vulkan.DeviceMemory
					if dd.AllocateMemory(device, &alloc, nil, &mem) != vulkan.Success {
						return
					}
					defer dd.FreeMemory(device, mem, nil)
					mustResult(t, "BindImageMemory", dd.BindImageMemory(device, image, mem, 0))
					get := vulkan.MemoryGetFdInfoKHR{SType: vulkan.StructureTypeMemoryGetFDInfoKHR, Memory: mem, HandleType: vulkan.ExternalMemoryHandleTypeDMABUFBitEXT}
					var fd int32 = -1
					mustResult(t, "GetMemoryFdKHR DMA_BUF", dd.GetMemoryFdKHR(device, &get, &fd))
					defer os.NewFile(uintptr(fd), "dmabuf").Close()
					p := vulkan.ImageDrmFormatModifierPropertiesEXT{SType: vulkan.StructureTypeImageDRMFormatModifierPropertiesEXT}
					mustResult(t, "GetImageDrmFormatModifierPropertiesEXT", dd.GetImageDrmFormatModifierPropertiesEXT(device, image, &p))
					t.Logf("%s: DMA-BUF B8G8R8A8_UNORM modifier=0x%x fd=%d exported", label, p.DrmFormatModifier, fd)
					exported = true
				}()
				if exported {
					break
				}
			}
			if !exported {
				t.Logf("UNSUPPORTED %s: no exportable one-plane color modifier", label)
			} else {
				passed++
			}
		}()
	}
	if passed == 0 {
		t.Skip("UNSUPPORTED: no device completed both exports")
	}
	t.Logf("positive hardware export devices: %d", passed)
}
func contains(items []string, needle string) bool {
	for _, s := range items {
		if s == needle {
			return true
		}
	}
	return false
}
func mustResult(t *testing.T, op string, r vulkan.Result) {
	t.Helper()
	if r != vulkan.Success {
		t.Fatalf("%s: %s", op, fmt.Sprint(vulkan.ResultString(r)))
	}
}

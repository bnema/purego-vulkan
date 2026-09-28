package integration_test

import (
	"fmt"
	"os"
	"runtime"
	"strings"

	"github.com/bnema/purego"
	"testing"
	"unsafe"

	"github.com/bnema/purego-vulkan/vulkan"
)

// Run with VEYA_VULKAN_PROBE=1; set VEYA_REQUIRE_POSITIVE=1 to require a successful export.
func TestVeyaHardwareExport(t *testing.T) {
	if os.Getenv("VEYA_VULKAN_PROBE") != "1" {
		t.Skip("set VEYA_VULKAN_PROBE=1")
	}
	if err := vulkan.Init(); err != nil {
		if strings.Contains(err.Error(), "resolve vkGetInstanceProcAddr") || strings.Contains(err.Error(), "dispatch") {
			t.Fatalf("BINDING_FAILURE: Init: %v", err)
		}
		t.Skipf("LOADER_UNAVAILABLE: %v", err)
	}
	gd := vulkan.Global()
	if gd == nil || !gd.HasCreateInstance() {
		t.Fatal("BINDING_FAILURE: vkCreateInstance missing")
	}
	app := cStringBytes("Vulkan hardware probe")
	ai := vulkan.ApplicationInfo{SType: vulkan.StructureTypeApplicationInfo, ApplicationName: &app[0], ApiVersion: (1 << 22) | (3 << 12)}
	ci := vulkan.InstanceCreateInfo{SType: vulkan.StructureTypeInstanceCreateInfo, ApplicationInfo: &ai}
	var instance vulkan.Instance
	mustResult(t, "CreateInstance", gd.CreateInstance(&ci, nil, &instance))
	id, err := vulkan.LoadInstanceDispatch(instance)
	if err != nil {
		if vulkan.VkDestroyInstance != nil {
			vulkan.VkDestroyInstance(instance, nil)
		}
		t.Fatalf("BINDING_FAILURE: LoadInstanceDispatch: %v", err)
	}
	defer id.DestroyInstance(instance, nil)
	if !id.HasGetPhysicalDeviceFeatures2() || !id.HasGetPhysicalDeviceImageFormatProperties2() {
		t.Fatalf("BINDING_FAILURE: capability query dispatch missing: features2=%t imageFormat2=%t", id.HasGetPhysicalDeviceFeatures2(), id.HasGetPhysicalDeviceImageFormatProperties2())
	}
	// Some Vulkan 1.1+ loaders expose the promoted core name but not the KHR alias.
	getExternalSemaphoreProperties := id.GetPhysicalDeviceExternalSemaphorePropertiesKHR
	if !id.HasGetPhysicalDeviceExternalSemaphorePropertiesKHR() {
		name := cStringBytes("vkGetPhysicalDeviceExternalSemaphoreProperties")
		addr := vulkan.VkGetInstanceProcAddr(instance, &name[0])
		if addr == 0 {
			t.Fatal("BINDING_FAILURE: external semaphore capability query missing")
		}
		purego.RegisterFunc(&getExternalSemaphoreProperties, uintptr(addr))
	}
	devices := enumeratePhysicalDevices(t, id, instance)
	if len(devices) == 0 {
		unsupportedProbe(t, "no Vulkan devices")
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
		features := vulkan.PhysicalDeviceFeatures2{SType: vulkan.StructureTypePhysicalDeviceFeatures2}
		timelineFeature := vulkan.PhysicalDeviceTimelineSemaphoreFeatures{SType: vulkan.StructureTypePhysicalDeviceTimelineSemaphoreFeatures}
		features.Next = unsafe.Pointer(&timelineFeature)
		id.GetPhysicalDeviceFeatures2(phys, &features)
		if timelineFeature.TimelineSemaphore == 0 {
			t.Logf("UNSUPPORTED %s: timelineSemaphore feature false", label)
			continue
		}
		semInfo := vulkan.PhysicalDeviceExternalSemaphoreInfo{SType: vulkan.StructureTypePhysicalDeviceExternalSemaphoreInfo, HandleType: vulkan.ExternalSemaphoreHandleTypeOpaqueFDBit}
		semType := vulkan.SemaphoreTypeCreateInfo{SType: vulkan.StructureTypeSemaphoreTypeCreateInfo, SemaphoreType: vulkan.SemaphoreTypeTimeline}
		semInfo.Next = unsafe.Pointer(&semType)
		semProps := vulkan.ExternalSemaphoreProperties{SType: vulkan.StructureTypeExternalSemaphoreProperties}
		getExternalSemaphoreProperties(phys, &semInfo, &semProps)
		if semProps.ExternalSemaphoreFeatures&vulkan.ExternalSemaphoreFeatureExportableBit == 0 || semProps.CompatibleHandleTypes&vulkan.ExternalSemaphoreHandleTypeOpaqueFDBit == 0 {
			t.Logf("UNSUPPORTED %s: timeline OPAQUE_FD semaphore export capability absent", label)
			continue
		}
		var count uint32
		id.GetPhysicalDeviceQueueFamilyProperties(phys, &count, nil)
		if count == 0 {
			t.Logf("UNSUPPORTED %s: no queue families", label)
			continue
		}
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
		// Query exportability for the exact format, usage, handle and each modifier.
		exportable := make([]vulkan.DrmFormatModifierPropertiesEXT, 0, len(list))
		for _, mod := range list {
			if mod.DrmFormatModifierPlaneCount != 1 || mod.DrmFormatModifierTilingFeatures&vulkan.FormatFeatureColorAttachmentBit == 0 {
				continue
			}
			modifierInfo := vulkan.PhysicalDeviceImageDrmFormatModifierInfoEXT{SType: vulkan.StructureTypePhysicalDeviceImageDRMFormatModifierInfoEXT, DrmFormatModifier: mod.DrmFormatModifier, SharingMode: vulkan.SharingModeExclusive}
			externalInfo := vulkan.PhysicalDeviceExternalImageFormatInfo{SType: vulkan.StructureTypePhysicalDeviceExternalImageFormatInfo, Next: unsafe.Pointer(&modifierInfo), HandleType: vulkan.ExternalMemoryHandleTypeDMABUFBitEXT}
			info := vulkan.PhysicalDeviceImageFormatInfo2{SType: vulkan.StructureTypePhysicalDeviceImageFormatInfo2, Next: unsafe.Pointer(&externalInfo), Format: vulkan.FormatB8g8r8a8Unorm, Type: vulkan.ImageType2d, Tiling: vulkan.ImageTilingDRMFormatModifierEXT, Usage: vulkan.ImageUsageColorAttachmentBit}
			externalProps := vulkan.ExternalImageFormatProperties{SType: vulkan.StructureTypeExternalImageFormatProperties}
			formatProps := vulkan.ImageFormatProperties2{SType: vulkan.StructureTypeImageFormatProperties2, Next: unsafe.Pointer(&externalProps)}
			r := id.GetPhysicalDeviceImageFormatProperties2(phys, &info, &formatProps)
			if r == vulkan.ErrorFormatNotSupported {
				continue
			}
			mustResult(t, "GetPhysicalDeviceImageFormatProperties2", r)
			if externalProps.ExternalMemoryProperties.ExternalMemoryFeatures&vulkan.ExternalMemoryFeatureExportableBit != 0 {
				exportable = append(exportable, mod)
			}
		}
		if len(exportable) == 0 {
			t.Logf("UNSUPPORTED %s: no exportable one-plane color modifier", label)
			continue
		}
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
			if result != vulkan.ErrorExtensionNotPresent && result != vulkan.ErrorFeatureNotPresent {
				t.Fatalf("BINDING_FAILURE %s: CreateDevice: %s", label, vulkan.ResultString(result))
			}
			t.Logf("UNSUPPORTED %s: CreateDevice timeline/extensions: %s", label, vulkan.ResultString(result))
			continue
		}
		dd, err := vulkan.LoadDeviceDispatch(id, device)
		if err != nil {
			// A partially loaded dispatch cannot destroy the device; resolve the core
			// command independently through the instance's vkGetDeviceProcAddr.
			if !id.HasGetDeviceProcAddr() {
				t.Fatalf("BINDING_FAILURE: cannot destroy device after LoadDeviceDispatch: %v", err)
			}
			name := cStringBytes("vkDestroyDevice")
			addr := id.GetDeviceProcAddr(device, &name[0])
			if addr == 0 {
				t.Fatalf("BINDING_FAILURE: vkDestroyDevice missing after LoadDeviceDispatch: %v", err)
			}
			var destroy func(vulkan.Device, *vulkan.AllocationCallbacks)
			purego.RegisterFunc(&destroy, uintptr(addr))
			destroy(device, nil)
			t.Fatalf("BINDING_FAILURE: LoadDeviceDispatch: %v", err)
		}
		func() {
			defer dd.DestroyDevice(device, nil)
			if !dd.HasGetSemaphoreFdKHR() || !dd.HasGetMemoryFdKHR() || !dd.HasGetImageDrmFormatModifierPropertiesEXT() {
				t.Fatalf("BINDING_FAILURE: advertised extension dispatch missing on %s", label)
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
			for _, mod := range exportable {
				if mod.DrmFormatModifierPlaneCount != 1 || mod.DrmFormatModifierTilingFeatures&vulkan.FormatFeatureColorAttachmentBit == 0 {
					continue
				}
				modifier := mod.DrmFormatModifier
				mc := vulkan.ImageDrmFormatModifierListCreateInfoEXT{SType: vulkan.StructureTypeImageDRMFormatModifierListCreateInfoEXT, DrmFormatModifierCount: 1, DrmFormatModifiers: &modifier}
				external := vulkan.ExternalMemoryImageCreateInfo{SType: vulkan.StructureTypeExternalMemoryImageCreateInfo, Next: unsafe.Pointer(&mc), HandleTypes: vulkan.ExternalMemoryHandleTypeDMABUFBitEXT}
				imageInfo := vulkan.ImageCreateInfo{SType: vulkan.StructureTypeImageCreateInfo, Next: unsafe.Pointer(&external), ImageType: vulkan.ImageType2d, Format: vulkan.FormatB8g8r8a8Unorm, Extent: vulkan.Extent3D{Width: 64, Height: 64, Depth: 1}, MipLevels: 1, ArrayLayers: 1, Samples: vulkan.SampleCount1Bit, Tiling: vulkan.ImageTilingDRMFormatModifierEXT, Usage: vulkan.ImageUsageColorAttachmentBit, SharingMode: vulkan.SharingModeExclusive}
				var image vulkan.Image
				mustResult(t, "CreateImage exportable modifier", dd.CreateImage(device, &imageInfo, nil, &image))
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
						t.Fatalf("BINDING_FAILURE %s: exportable image has no device-local memory type", label)
					}
					dedicated := vulkan.MemoryDedicatedAllocateInfo{SType: vulkan.StructureTypeMemoryDedicatedAllocateInfo, Image: image}
					exp := vulkan.ExportMemoryAllocateInfo{SType: vulkan.StructureTypeExportMemoryAllocateInfo, Next: unsafe.Pointer(&dedicated), HandleTypes: vulkan.ExternalMemoryHandleTypeDMABUFBitEXT}
					alloc := vulkan.MemoryAllocateInfo{SType: vulkan.StructureTypeMemoryAllocateInfo, Next: unsafe.Pointer(&exp), AllocationSize: req.Size, MemoryTypeIndex: uint32(memType)}
					var mem vulkan.DeviceMemory
					mustResult(t, "AllocateMemory exportable image", dd.AllocateMemory(device, &alloc, nil, &mem))
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
				t.Fatalf("BINDING_FAILURE %s: queried exportable modifier did not export", label)
			} else {
				passed++
			}
		}()
	}
	if passed == 0 {
		unsupportedProbe(t, "no device completed both exports")
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
		t.Fatalf("BINDING_FAILURE %s: %s", op, fmt.Sprint(vulkan.ResultString(r)))
	}
}

func unsupportedProbe(t *testing.T, reason string) {
	t.Helper()
	if os.Getenv("VEYA_REQUIRE_POSITIVE") == "1" {
		t.Fatalf("UNSUPPORTED: %s (VEYA_REQUIRE_POSITIVE=1)", reason)
	}
	t.Skipf("UNSUPPORTED: %s", reason)
}

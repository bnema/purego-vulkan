# Veya presentation binding audit

The external consumer compile fixture is `integration/veya_contract_compile_test.go`;
`VEYA_VULKAN_PROBE=1 go test ./integration -run TestVeyaHardwareExport -v -count=1`
requires positive hardware exports (otherwise reports `UNSUPPORTED`, never a positive pass).
All symbols below are generated from the pinned Vulkan registry, not handwritten.

| Need | Generated symbols | Status |
|---|---|---|
| Loader, instance, devices, queue selection | `Init`, `Global`, `CreateInstance`, `LoadInstanceDispatch`, `EnumeratePhysicalDevices`, `EnumerateDeviceExtensionProperties`, `GetPhysicalDeviceQueueFamilyProperties`, `CreateDevice`, `LoadDeviceDispatch`, `GetDeviceQueue` | present |
| Match linux-dmabuf `main_device` | `GetPhysicalDeviceProperties2`, `PhysicalDeviceDrmPropertiesEXT` (`VK_EXT_physical_device_drm`) | present; compare both primary and render major/minor with decoded `dev_t` |
| Format/modifier capabilities | `GetPhysicalDeviceFormatProperties2`, `DrmFormatModifierPropertiesListEXT`, `DrmFormatModifierPropertiesList2EXT`, `GetPhysicalDeviceImageFormatProperties2`, `PhysicalDeviceImageDrmFormatModifierInfoEXT`, `FormatB8g8r8a8Unorm` | present; intersect with compositor feedback and require one plane |
| Image export | `ExternalMemoryImageCreateInfo`, `ImageDrmFormatModifierListCreateInfoEXT`, `ImageDrmFormatModifierExplicitCreateInfoEXT`, `CreateImage`, `GetImageMemoryRequirements2`, `MemoryDedicatedAllocateInfo`, `ExportMemoryAllocateInfo`, `AllocateMemory`, `BindImageMemory2`, `GetMemoryFdKHR`, `GetMemoryFdPropertiesKHR`, `GetImageDrmFormatModifierPropertiesEXT` | present; DMA_BUF export positive on both RADV devices |
| Timeline/external FD | `PhysicalDeviceTimelineSemaphoreFeatures`, `SemaphoreTypeCreateInfo`, `ExportSemaphoreCreateInfo`, `CreateSemaphore`, `SemaphoreGetFdInfoKHR`, `GetSemaphoreFdKHR`, `ImportSemaphoreFdInfoKHR`, `ImportSemaphoreFdKHR`, `WaitSemaphores`, `SignalSemaphore`, `GetSemaphoreCounterValue`, `ExternalSemaphoreHandleTypeOpaqueFDBit`, `ExternalSemaphoreHandleTypeSyncFDBit` | present; timeline OPAQUE_FD export positive on both RADV devices; syncobj conversion not proven |
| Transfers, readback and synchronization | `CreateBuffer`, `BindBufferMemory`, `MapMemory`, `CmdCopyImageToBuffer`, `CmdCopyBufferToImage`, `CmdPipelineBarrier2`/`KHR`, `DependencyInfo`, `ImageMemoryBarrier2`, `QueueSubmit2`/`KHR` | present |
| Drawing | `CmdBeginRendering`, `CmdEndRendering`, `RenderingInfo`, `CreateShaderModule`, `CreateGraphicsPipelines`, `CreatePipelineLayout`, `CmdPushConstants`, `CreateDescriptorSetLayout`, `CreateDescriptorPool`, `AllocateDescriptorSets`, `UpdateDescriptorSets`, `CreateSampler`, `CmdDraw` | present |
| Validation and teardown | `CreateDebugUtilsMessengerEXT`, `DestroyDebugUtilsMessengerEXT`, `DeviceWaitIdle`, `QueueWaitIdle`, `DestroyImage`, `FreeMemory`, `DestroySemaphore`, `DestroyDevice`, `DestroyInstance` | present; callback registration/lifetime and validation output not hardware-tested |

`wp_linux_drm_syncobj` **imports a DRM syncobj FD**, not an arbitrary Vulkan
OPAQUE_FD. Do not pass an OPAQUE_FD directly merely because RADV exported it.
On RADV the export may refer to a DRM syncobj, but the probe currently establishes
only successful FD export, **not** FD_TO_HANDLE acceptance or timeline-point
sharing with the compositor. The portable bridge is: Vulkan binary semaphore
SYNC_FD export/import at the appropriate submit boundary, a DRM syncobj created
on the matching render node, and DRM SYNCOBJ_FD_TO_HANDLE (flag
IMPORT_SYNC_FILE), SYNCOBJ_TRANSFER for timeline-point movement, and
SYNCOBJ_HANDLE_TO_FD (without EXPORT_SYNC_FILE) to send the syncobj FD to
Wayland. For release, transfer the compositor's timeline point to a temporary
syncobj then export a SYNC_FILE and import it temporarily into a Vulkan binary
semaphore (or otherwise wait for the DRM timeline point). Implement and test
these ioctls using `golang.org/x/sys/unix` on the selected render node in Veya;
verify FD ownership and kernel support. Never equate Vulkan timeline values
with syncobj points without an explicit proven bridge.

Open risks: compositor format/modifier intersection, explicit timeline-point
interop, synchronization validation, and end-to-end buffer lifecycle remain P5
integration work. The probe does not assert compositor acceptance, syncobj FD
identity, or rendering correctness.

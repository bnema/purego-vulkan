# purego-vulkan

Linux-first Vulkan bindings for Go without cgo.

`purego-vulkan` loads the system Vulkan loader (`libvulkan.so.1`) through `github.com/bnema/purego` and generates its raw Vulkan API surface from a pinned Khronos `vk.xml` snapshot.

The package is intentionally low-level. It exposes Vulkan ABI types, constants, command signatures, and explicit global/instance/device dispatch tables for compositor and renderer code that wants to own its rendering architecture.

## Status

Early v0.x implementation. The initial target is the Go Wayland compositor renderer path on Linux/DRM. Other platforms, high-level rendering helpers, swapchain management, and scene abstractions are out of scope for this milestone.

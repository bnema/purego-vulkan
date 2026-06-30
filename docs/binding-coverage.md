# Vulkan binding coverage

`cmd/vulkangen` reads the pinned Khronos registry at `registry/vk.xml` and supports three named coverage profiles:

| Profile | Intended use | Commands | Types | Constants | Extensions |
| --- | --- | ---: | ---: | ---: | ---: |
| `renderer` | Low-level offscreen/compositor renderer subset | 103 | 340 | 810 | 14 |
| `wsi` | Default generated package: renderer subset plus Linux surface/swapchain and image readback | 124 | 379 | 865 | 19 |
| `complete` | Audit / broad command generation for the pinned registry | 857 | 1074 | 1724 | 0 |

The pinned registry currently contains 864 command declarations. The complete profile selects 857 unique Vulkan command names after duplicate Vulkan/Vulkan SC registry variants are normalized away.

## Default profile

Committed generated files use the `wsi` profile. It includes:

- the renderer command set from `renderer`;
- required core memory-property queries (`vkGetPhysicalDeviceMemoryProperties` and `vkGetPhysicalDeviceMemoryProperties2`);
- required core `vkCmdPipelineBarrier` coverage plus optional synchronization2 extension commands;
- `vkCmdCopyImageToBuffer` for image readback from offscreen examples;
- `VK_KHR_surface` and `VK_KHR_swapchain`;
- Linux WSI platform extensions: `VK_KHR_wayland_surface`, `VK_KHR_xcb_surface`, and `VK_KHR_xlib_surface`.

WSI and other extension commands are loaded as optional dispatch fields because availability depends on enabled instance/device extensions and driver support. Core renderer commands remain required for the corresponding dispatch loader.

## Complete profile

Use the complete profile when auditing command coverage against the registry or when a consumer needs command declarations beyond the default WSI surface:

```sh
go run ./cmd/vulkangen --registry ./registry/vk.xml --out . --profile complete
```

The complete profile marks commands outside the renderer baseline optional so `LoadInstanceDispatch` and `LoadDeviceDispatch` do not require every extension command to resolve on a given system. It selects all command signatures and their type/constant closure, but it does not emit every extension metadata constant because several registry extension aliases collapse to duplicate Go names in a single package. Extension metadata for the committed WSI surface is emitted by the default `wsi` profile.

## Intentionally unsupported

The package remains low-level Vulkan bindings only. It does not provide swapchain ownership, renderer framework code, render-pass builders, descriptor allocators, scene abstractions, debug callback helpers, or platform window integration. Consumers create native windows and pass platform handles into the generated Vulkan structs explicitly.

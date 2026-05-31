# Vulkan registry snapshot

`vk.xml` is pinned from the Khronos Vulkan registry:

- Source: https://raw.githubusercontent.com/KhronosGroup/Vulkan-Docs/main/xml/vk.xml
- Snapshot date: 2026-05-31

Update with:

```sh
rtk curl -L https://raw.githubusercontent.com/KhronosGroup/Vulkan-Docs/main/xml/vk.xml -o registry/vk.xml
rtk go generate ./...
rtk go test ./...
```

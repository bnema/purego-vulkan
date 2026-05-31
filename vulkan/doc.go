// Package vulkan exposes low-level Vulkan bindings loaded through purego.
//
// The package is Linux-first and intentionally close to Vulkan's C ABI. Raw
// calls use explicit pointers, Vulkan-sized scalar types, and caller-managed
// object lifetimes. Higher-level renderer policy belongs in consuming projects.
package vulkan

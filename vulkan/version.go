package vulkan

import "fmt"

func MakeVersion(major, minor, patch uint32) uint32 {
	return MakeAPIVersion(0, major, minor, patch)
}

func MakeAPIVersion(variant, major, minor, patch uint32) uint32 {
	return variant<<29 | major<<22 | minor<<12 | patch
}

func VersionVariant(version uint32) uint32 {
	return version >> 29
}

func VersionMajor(version uint32) uint32 {
	return (version >> 22) & 0x7f
}

func VersionMinor(version uint32) uint32 {
	return (version >> 12) & 0x3ff
}

func VersionPatch(version uint32) uint32 {
	return version & 0xfff
}

func FormatVersion(version uint32) string {
	return fmt.Sprintf("%d.%d.%d", VersionMajor(version), VersionMinor(version), VersionPatch(version))
}

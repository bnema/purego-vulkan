package capi

import "bnema.dev/purego"

// RegisterFunc binds fptr to the C function address addr.
func RegisterFunc(fptr any, addr uintptr) {
	purego.RegisterFunc(fptr, addr)
}

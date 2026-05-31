package capi

import "testing"

func TestRegisterRequiredTreatsZeroAddressAsMissing(t *testing.T) {
	var fn func()
	err := registerRequired([]string{"vkMissing"}, 0, func(uintptr, string) (uintptr, error) {
		return 0, nil
	}, map[string]any{"vkMissing": &fn})
	if err == nil {
		t.Fatal("registerRequired() error = nil for zero address")
	}
	if fn != nil {
		t.Fatal("registerRequired() registered a zero-address function")
	}
}

func TestRegisterOptionalTreatsZeroAddressAsMissing(t *testing.T) {
	var fn func()
	registerOptional([]string{"vkOptionalMissing"}, 0, func(uintptr, string) (uintptr, error) {
		return 0, nil
	}, map[string]any{"vkOptionalMissing": &fn})
	if fn != nil {
		t.Fatal("registerOptional() registered a zero-address function")
	}
}

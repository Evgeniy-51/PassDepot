//go:build !windows

package singleinst

// Acquire на не-Windows не ограничивает число экземпляров.
func Acquire() bool { return true }

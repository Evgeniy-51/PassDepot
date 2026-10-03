//go:build windows

package singleinst

import (
	"errors"
	"unsafe"

	"golang.org/x/sys/windows"
)

// Per-session mutex: one instance per logged-in user, portable and installed share it.
const mutexName = `Local\PassDepot.SingleInstance`

const (
	swRestore = 9
	hwndTop   = 0
)

var (
	modUser32               = windows.NewLazySystemDLL("user32.dll")
	procFindWindowW         = modUser32.NewProc("FindWindowW")
	procShowWindow          = modUser32.NewProc("ShowWindow")
	procIsIconic            = modUser32.NewProc("IsIconic")
	procSetForegroundWindow = modUser32.NewProc("SetForegroundWindow")
	procSetWindowPos        = modUser32.NewProc("SetWindowPos")

	// Held until process exit so the mutex stays owned.
	heldMutex windows.Handle
)

// Acquire — true, если это первый экземпляр. Иначе поднимает уже открытое окно.
func Acquire() bool {
	name, err := windows.UTF16PtrFromString(mutexName)
	if err != nil {
		return true
	}
	h, err := windows.CreateMutex(nil, false, name)
	if errors.Is(err, windows.ERROR_ALREADY_EXISTS) {
		if h != 0 {
			_ = windows.CloseHandle(h)
		}
		activateExisting()
		return false
	}
	if err != nil {
		return true
	}
	heldMutex = h
	return true
}

func activateExisting() {
	title, err := windows.UTF16PtrFromString("PassDepot")
	if err != nil {
		return
	}
	hwnd, _, _ := procFindWindowW.Call(0, uintptr(unsafe.Pointer(title)))
	if hwnd == 0 {
		return
	}
	iconic, _, _ := procIsIconic.Call(hwnd)
	if iconic != 0 {
		_, _, _ = procShowWindow.Call(hwnd, swRestore)
	}
	const swpNoSize, swpNoMove, swpShowWindow = 0x0001, 0x0002, 0x0040
	_, _, _ = procSetWindowPos.Call(hwnd, hwndTop, 0, 0, 0, 0, swpNoSize|swpNoMove|swpShowWindow)
	_, _, _ = procSetForegroundWindow.Call(hwnd)
}

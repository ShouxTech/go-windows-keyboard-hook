package main

import (
	"fmt"
	"go-windows-keyboard-hook/vk_codes"
	"syscall"
	"unsafe"

	"golang.org/x/sys/windows"
)

type HOOKPROC func(int, uintptr, uintptr) uintptr

type KBDLLHOOKSTRUCT struct {
	VkCode      uint32
	ScanCode    uint32
	Flags       uint32
	Time        uint32
	DwExtraInfo uintptr
}

type MSG struct {
	Hwnd    uintptr
	Message uint32
	WParam  uintptr
	LParam  uintptr
	Time    uint32
	Pt      struct {
		X, Y int32
	}
}

const (
	WH_KEYBOARD_LL = 13
	WM_KEYDOWN     = 256
)

var (
	user32                  = windows.NewLazySystemDLL("user32.dll")
	procSetWindowsHookExA   = user32.NewProc("SetWindowsHookExA")
	procCallNextHookEx      = user32.NewProc("CallNextHookEx")
	procUnhookWindowsHookEx = user32.NewProc("UnhookWindowsHookEx")
	procGetMessage          = user32.NewProc("GetMessageW")
	procTranslateMessage    = user32.NewProc("TranslateMessage")
	procDispatchMessage     = user32.NewProc("DispatchMessageW")
)

var keyboardHook uintptr

func SetWindowsHookExA(idHook int, lpfn HOOKPROC, hmod uintptr, dwThreadId uint32) uintptr {
	ret, _, _ := procSetWindowsHookExA.Call(
		uintptr(idHook),
		syscall.NewCallback(lpfn),
		hmod,
		uintptr(dwThreadId),
	)
	return ret
}

func CallNextHookEx(hhk uintptr, nCode int, wParam uintptr, lParam uintptr) uintptr {
	ret, _, _ := procCallNextHookEx.Call(
		uintptr(hhk),
		uintptr(nCode),
		uintptr(wParam),
		uintptr(lParam),
	)
	return ret
}

func UnhookWindowsHookEx(hhk uintptr) bool {
	ret, _, _ := procUnhookWindowsHookEx.Call(
		uintptr(hhk),
	)
	return ret != 0
}

func GetMessage(msg *MSG, hwnd uintptr, msgFilterMin uint32, msgFilterMax uint32) int {
	ret, _, _ := procGetMessage.Call(
		uintptr(unsafe.Pointer(msg)),
		hwnd,
		uintptr(msgFilterMin),
		uintptr(msgFilterMax))
	return int(ret)
}

func TranslateMessage(msg *MSG) bool {
	ret, _, _ := procTranslateMessage.Call(uintptr(unsafe.Pointer(msg)))
	return ret != 0
}

func DispatchMessage(msg *MSG) uintptr {
	ret, _, _ := procDispatchMessage.Call(uintptr(unsafe.Pointer(msg)))
	return ret
}

func start() {
	keyboardHook = SetWindowsHookExA(WH_KEYBOARD_LL, func(nCode int, wParam uintptr, lParam uintptr) uintptr {
		if nCode == 0 && wParam == WM_KEYDOWN {
			kbdstruct := (*KBDLLHOOKSTRUCT)(unsafe.Pointer(lParam))

			fmt.Printf("Key pressed: %q\n", byte(kbdstruct.VkCode))

			if kbdstruct.VkCode == vk_codes.VK_H {
				fmt.Println("ello world!")
			}
		}
		return CallNextHookEx(keyboardHook, nCode, wParam, lParam)
	}, 0, 0)
	defer UnhookWindowsHookEx(keyboardHook)

	// Message loop required in order for hook callback to work.
	// There's also strange slow system-wide typing behavior if you remove the message loop.
	var msg MSG
	for GetMessage(&msg, 0, 0, 0) != 0 {
		TranslateMessage(&msg)
		DispatchMessage(&msg)
	}
}

func main() {
	fmt.Println("Listening for key presses...")
	go start()

	select {} // Keep the program running.
}

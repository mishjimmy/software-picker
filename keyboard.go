package main

import (
	"context"
	"fmt"
	"syscall"
	"time"
	"unsafe"

	"github.com/lxn/win"
	"github.com/wailsapp/wails/v2/pkg/runtime"
)

// Global variable to store our window handle
var mainWindowHandle uintptr

func SetMainWindowHandle(hwnd uintptr) {
	mainWindowHandle = hwnd
	fmt.Printf("Set main window handle to: %v\n", hwnd)
}

func startKeyboardMonitor(ctx context.Context) {
	user32 := syscall.NewLazyDLL("user32.dll")
	getAsyncKeyState := user32.NewProc("GetAsyncKeyState")
	getForegroundWindow := user32.NewProc("GetForegroundWindow")
	findWindow := user32.NewProc("FindWindowW")

	// Get our window handle
	hwnd, _, _ := findWindow.Call(0, uintptr(unsafe.Pointer(syscall.StringToUTF16Ptr("LDLauncher"))))
	SetMainWindowHandle(hwnd)

	// Variables for debouncing
	var lastUpPress time.Time
	var lastDownPress time.Time
	var lastEnterPress time.Time
	var lastEscapePress time.Time
	const debounceTime = 150 * time.Millisecond // Adjust this value to control the delay

	go func() {
		for {
			now := time.Now()

			// Get foreground window
			foregroundWnd, _, _ := getForegroundWindow.Call()

			// Only process keys if our window is the foreground window
			if foregroundWnd == 0 || foregroundWnd != mainWindowHandle {
				time.Sleep(100 * time.Millisecond) // Sleep longer when not focused
				continue
			}

			// Check for up arrow
			ret, _, _ := getAsyncKeyState.Call(uintptr(win.VK_UP))
			if ret&0x8000 != 0 && now.Sub(lastUpPress) > debounceTime {
				fmt.Println("Up arrow pressed")
				runtime.EventsEmit(ctx, "keydown", "ArrowUp")
				lastUpPress = now
				time.Sleep(debounceTime) // Add delay after key press
			}

			// Check for down arrow
			ret, _, _ = getAsyncKeyState.Call(uintptr(win.VK_DOWN))
			if ret&0x8000 != 0 && now.Sub(lastDownPress) > debounceTime {
				fmt.Println("Down arrow pressed")
				runtime.EventsEmit(ctx, "keydown", "ArrowDown")
				lastDownPress = now
				time.Sleep(debounceTime) // Add delay after key press
			}

			// Check for enter
			ret, _, _ = getAsyncKeyState.Call(uintptr(win.VK_RETURN))
			if ret&0x8000 != 0 && now.Sub(lastEnterPress) > debounceTime {
				fmt.Println("Enter pressed")
				runtime.EventsEmit(ctx, "keydown", "Enter")
				lastEnterPress = now
				time.Sleep(debounceTime) // Add delay after key press
			}

			// Check for escape
			ret, _, _ = getAsyncKeyState.Call(uintptr(win.VK_ESCAPE))
			if ret&0x8000 != 0 && now.Sub(lastEscapePress) > debounceTime {
				fmt.Println("Escape pressed")
				runtime.EventsEmit(ctx, "keydown", "Escape")
				lastEscapePress = now
				time.Sleep(debounceTime) // Add delay after key press
			}

			// Small delay to prevent CPU overuse
			time.Sleep(10 * time.Millisecond)
		}
	}()
} 
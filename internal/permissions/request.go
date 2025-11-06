package permissions

import (
	"fmt"
	"os/exec"
	"time"
)

// RequestFlow guides the user through granting permissions
func RequestFlow() error {
	fmt.Println("🔐 rekap permission setup")
	fmt.Println()
	fmt.Println("rekap needs certain permissions to provide full functionality.")
	fmt.Println("Let's check what's available and help you enable missing permissions.")
	fmt.Println()

	// Check current status
	caps := Check()
	
	// Full Disk Access
	fmt.Println("📊 Full Disk Access (Screen Time data)")
	fmt.Println("   Enables: App usage tracking, screen-on time, focus streaks")
	if caps.FullDiskAccess {
		fmt.Println("   ✓ Already granted")
	} else {
		fmt.Println("   ✗ Not granted")
		fmt.Println()
		fmt.Println("   To grant Full Disk Access:")
		fmt.Println("   1. System Settings will open to Privacy & Security")
		fmt.Println("   2. Click 'Full Disk Access' in the sidebar")
		fmt.Println("   3. Enable 'rekap' or your terminal app")
		fmt.Println()
		fmt.Print("   Press Enter to open System Settings...")
		fmt.Scanln()
		
		// Open System Settings to Privacy & Security
		exec.Command("open", "x-apple.systempreferences:com.apple.preference.security?Privacy_AllFiles").Run()
		
		// Wait for user to grant permission
		fmt.Println()
		fmt.Println("   Waiting for permission to be granted...")
		fmt.Println("   (This window will auto-update when detected)")
		fmt.Println()
		
		waitForPermission("Full Disk Access", func() bool {
			return checkFullDiskAccess()
		})
	}
	fmt.Println()

	// Accessibility
	fmt.Println("♿ Accessibility (UI element access)")
	fmt.Println("   Enables: Frontmost app detection (fallback method)")
	if caps.Accessibility {
		fmt.Println("   ✓ Already granted")
	} else {
		fmt.Println("   ✗ Not granted")
		fmt.Println()
		fmt.Println("   To grant Accessibility:")
		fmt.Println("   1. System Settings will open to Privacy & Security")
		fmt.Println("   2. Click 'Accessibility' in the sidebar")
		fmt.Println("   3. Enable 'rekap' or your terminal app")
		fmt.Println()
		fmt.Print("   Press Enter to open System Settings...")
		fmt.Scanln()
		
		exec.Command("open", "x-apple.systempreferences:com.apple.preference.security?Privacy_Accessibility").Run()
		
		fmt.Println()
		fmt.Println("   Waiting for permission to be granted...")
		fmt.Println()
		
		waitForPermission("Accessibility", func() bool {
			return checkAccessibility()
		})
	}
	fmt.Println()

	// Final status
	finalCaps := Check()
	fmt.Println("✅ Setup complete!")
	fmt.Println()
	fmt.Println("Current capabilities:")
	fmt.Println(FormatCapabilities(finalCaps))
	fmt.Println()
	fmt.Println("Run 'rekap' to see your activity summary.")
	fmt.Println("Run 'rekap doctor' anytime to check permissions.")
	
	return nil
}

// waitForPermission polls for a permission to be granted
func waitForPermission(name string, checkFunc func() bool) {
	ticker := time.NewTicker(1 * time.Second)
	defer ticker.Stop()
	
	timeout := time.After(60 * time.Second)
	
	for {
		select {
		case <-ticker.C:
			if checkFunc() {
				fmt.Printf("   ✓ %s granted!\n", name)
				return
			}
		case <-timeout:
			fmt.Printf("   ⏱️  Timeout waiting for %s\n", name)
			fmt.Printf("   You can grant it later and run 'rekap init' again\n")
			return
		}
	}
}

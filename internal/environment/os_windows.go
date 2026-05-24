//go:build windows

package environment

import (
	"fmt"
	"syscall"
	"unsafe"
)

func detectOSVersion() string {
	// RtlGetVersion via ntdll — no subprocess.
	ntdll := syscall.NewLazyDLL("ntdll.dll")
	proc := ntdll.NewProc("RtlGetVersion")

	type osVersionInfoEx struct {
		Size             uint32
		MajorVersion     uint32
		MinorVersion     uint32
		BuildNumber      uint32
		PlatformId       uint32
		CSDVersion       [128]uint16
		ServicePackMajor uint16
		ServicePackMinor uint16
		SuiteMask        uint16
		ProductType      byte
		Reserved         byte
	}

	var info osVersionInfoEx
	info.Size = uint32(unsafe.Sizeof(info))
	ret, _, _ := proc.Call(uintptr(unsafe.Pointer(&info)))
	if ret != 0 {
		return ""
	}
	return fmt.Sprintf("%d.%d.%d", info.MajorVersion, info.MinorVersion, info.BuildNumber)
}

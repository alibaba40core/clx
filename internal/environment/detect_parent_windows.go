//go:build windows

package environment

import (
	"os"
	"path/filepath"
	"strings"
	"syscall"
	"unsafe"
)

const (
	th32csSnapProcess = 0x00000002
	maxPath           = 260
)

type processEntry32 struct {
	Size            uint32
	Usage           uint32
	ProcessID       uint32
	DefaultHeapID   uintptr
	ModuleID        uint32
	Threads         uint32
	ParentProcessID uint32
	PriClassBase    int32
	Flags           uint32
	ExeFile         [maxPath]uint16
}

var (
	modKernel32                  = syscall.NewLazyDLL("kernel32.dll")
	procCreateToolhelp32Snapshot = modKernel32.NewProc("CreateToolhelp32Snapshot")
	procProcess32FirstW          = modKernel32.NewProc("Process32FirstW")
	procProcess32NextW             = modKernel32.NewProc("Process32NextW")
	procCloseHandle                = modKernel32.NewProc("CloseHandle")
)

// parentProcessBaseName returns the executable file name of this process's parent.
func parentProcessBaseName() string {
	ppid := uint32(os.Getppid())
	if ppid == 0 {
		return ""
	}
	snap, _, _ := procCreateToolhelp32Snapshot.Call(th32csSnapProcess, 0)
	if snap == uintptr(syscall.InvalidHandle) {
		return ""
	}
	defer procCloseHandle.Call(snap)

	var entry processEntry32
	entry.Size = uint32(unsafe.Sizeof(entry))
	ret, _, _ := procProcess32FirstW.Call(snap, uintptr(unsafe.Pointer(&entry)))
	for ret != 0 {
		if entry.ProcessID == ppid {
			return strings.ToLower(filepath.Base(syscall.UTF16ToString(entry.ExeFile[:])))
		}
		ret, _, _ = procProcess32NextW.Call(snap, uintptr(unsafe.Pointer(&entry)))
	}
	return ""
}

// shellFromParentExecutable maps a parent exe base name to a shell id, or "" if unknown.
func shellFromParentExecutable(base string) string {
	switch {
	case strings.Contains(base, "pwsh"):
		return "pwsh"
	case strings.Contains(base, "powershell"):
		return "powershell"
	case base == "cmd.exe":
		return "cmd"
	case strings.Contains(base, "mintty"), base == "winpty-agent.exe":
		return "bash"
	case strings.Contains(base, "bash"), strings.Contains(base, "sh"):
		return detectShellUnix()
	default:
		return ""
	}
}

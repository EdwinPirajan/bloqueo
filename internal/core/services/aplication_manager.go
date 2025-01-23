package services

import (
	"fmt"
	"unsafe"

	"github.com/fatih/color"
	"golang.org/x/sys/windows"
)

type ProcessInfo struct {
	Name string
	ID   uint32
}

type ApplicationManager interface {
	SuspendProcess(handle windows.Handle) error
	ResumeProcess(handle windows.Handle) error
	ListApplicationsInCurrentSession() ([]ProcessInfo, error)
	GetProcessHandlesInCurrentSession(processName string) ([]windows.Handle, error)
	Intersect(a, b []ProcessInfo) []ProcessInfo
	EqualProcessSlices(a, b []ProcessInfo) bool
}

type windowsApplicationManager struct {
	systemManager SystemManager
}

func NewWindowsApplicationManager(systemManager SystemManager) ApplicationManager {
	return &windowsApplicationManager{systemManager: systemManager}
}

func (am *windowsApplicationManager) SuspendProcess(handle windows.Handle) error {
	color.Red("Suspendiendo proceso")
	ntSuspendProcess := windows.NewLazySystemDLL("ntdll.dll").NewProc("NtSuspendProcess")
	r1, _, e1 := ntSuspendProcess.Call(uintptr(handle))
	if r1 != 0 {
		if e1 != nil && e1 != windows.ERROR_SUCCESS {
			return fmt.Errorf("failed to suspend process: %v", e1)
		}
	}
	return nil
}

func (am *windowsApplicationManager) ResumeProcess(handle windows.Handle) error {
	color.Green("Reanudando proceso")
	ntResumeProcess := windows.NewLazySystemDLL("ntdll.dll").NewProc("NtResumeProcess")
	r1, _, e1 := ntResumeProcess.Call(uintptr(handle))
	if r1 != 0 {
		if e1 != nil && e1 != windows.ERROR_SUCCESS {
			return fmt.Errorf("failed to resume process: %v", e1)
		}
	}
	return nil
}

func (am *windowsApplicationManager) ListApplicationsInCurrentSession() ([]ProcessInfo, error) {
	snapshot, err := windows.CreateToolhelp32Snapshot(windows.TH32CS_SNAPPROCESS, 0)
	if err != nil {
		return nil, err
	}
	defer windows.CloseHandle(snapshot)

	var entry windows.ProcessEntry32
	entry.Size = uint32(unsafe.Sizeof(entry))
	var apps []ProcessInfo

	currentSessionID, err := am.systemManager.GetCurrentSessionID()
	if err != nil {
		return nil, err
	}

	for {
		err = windows.Process32Next(snapshot, &entry)
		if err != nil {
			break
		}
		var sessionID uint32
		err = windows.ProcessIdToSessionId(uint32(entry.ProcessID), &sessionID)
		if err != nil {
			continue
		}
		if sessionID == currentSessionID {
			apps = append(apps, ProcessInfo{Name: windows.UTF16ToString(entry.ExeFile[:]), ID: uint32(entry.ProcessID)})
		}
	}

	if len(apps) == 0 {
		return nil, fmt.Errorf("no applications found in the current session")
	}

	return apps, nil
}

func (am *windowsApplicationManager) GetProcessHandlesInCurrentSession(processName string) ([]windows.Handle, error) {
	snapshot, err := windows.CreateToolhelp32Snapshot(windows.TH32CS_SNAPPROCESS, 0)
	if err != nil {
		return nil, err
	}
	defer windows.CloseHandle(snapshot)

	var entry windows.ProcessEntry32
	entry.Size = uint32(unsafe.Sizeof(entry))
	var handles []windows.Handle

	currentSessionID, err := am.systemManager.GetCurrentSessionID()
	if err != nil {
		return nil, err
	}

	for {
		err = windows.Process32Next(snapshot, &entry)
		if err != nil {
			break
		}
		if windows.UTF16ToString(entry.ExeFile[:]) == processName {
			var sessionID uint32
			err := windows.ProcessIdToSessionId(uint32(entry.ProcessID), &sessionID)
			if err != nil {
				continue
			}
			if sessionID == currentSessionID {
				handle, err := windows.OpenProcess(windows.PROCESS_SUSPEND_RESUME|windows.PROCESS_QUERY_INFORMATION, false, entry.ProcessID)
				if err != nil {
					fmt.Printf("Error opening process %d: %v\n", entry.ProcessID, err)
					continue
				}
				fmt.Printf("Found process %s (PID: %d) in current session.\n", processName, entry.ProcessID)
				handles = append(handles, handle)
			}
		}
	}

	if len(handles) == 0 {
		return nil, fmt.Errorf("no instances of process %s found in the current session", processName)
	}

	return handles, nil
}

func (am *windowsApplicationManager) Intersect(a, b []ProcessInfo) []ProcessInfo {
	m := make(map[string]bool)
	for _, item := range b {
		m[item.Name] = true
	}
	var result []ProcessInfo
	for _, item := range a {
		if _, ok := m[item.Name]; ok {
			result = append(result, item)
		}
	}
	return result
}

func (am *windowsApplicationManager) EqualProcessSlices(a, b []ProcessInfo) bool {
	if len(a) != len(b) {
		return false
	}
	m := make(map[uint32]ProcessInfo)
	for _, item := range a {
		m[item.ID] = item
	}
	for _, item := range b {
		if _, ok := m[item.ID]; !ok {
			return false
		}
	}
	return true
}

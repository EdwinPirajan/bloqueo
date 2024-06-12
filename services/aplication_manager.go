package services

import (
	"fmt"
	"sync"
	"unsafe"

	"github.com/fatih/color"
	"golang.org/x/sys/windows"
)

type ProcessInfo struct {
	Name string
	ID   uint32
}

type ApplicationManager interface {
	BlockApplication(processName string) error
	UnblockApplication(processName string) error
	SuspendProcess(handle windows.Handle) error
	ResumeProcess(handle windows.Handle) error
	ListApplications() ([]ProcessInfo, error)
	Intersect(a, b []ProcessInfo) []ProcessInfo
	EqualProcessSlices(a, b []ProcessInfo) bool
	GetProcessHandles(processName string) ([]windows.Handle, error)
	GetProcessSessionID(processID uint32) (uint32, error)
}

type windowsApplicationManager struct {
	mu            sync.Mutex
	systemManager SystemManager
}

func NewWindowsApplicationManager(systemManager SystemManager) ApplicationManager {
	return &windowsApplicationManager{systemManager: systemManager}
}

func suspendProcess(handle windows.Handle) error {
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

func resumeProcess(handle windows.Handle) error {
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

func getProcessHandles(systemManager SystemManager, processName string) ([]windows.Handle, error) {
	snapshot, err := windows.CreateToolhelp32Snapshot(windows.TH32CS_SNAPPROCESS, 0)
	if err != nil {
		return nil, err
	}
	defer windows.CloseHandle(snapshot)

	var entry windows.ProcessEntry32
	entry.Size = uint32(unsafe.Sizeof(entry))
	var handles []windows.Handle
	currentSessionID, err := systemManager.GetCurrentSessionID()
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

func (s *windowsApplicationManager) ListApplications() ([]ProcessInfo, error) {
	snapshot, err := windows.CreateToolhelp32Snapshot(windows.TH32CS_SNAPPROCESS, 0)
	if err != nil {
		return nil, err
	}
	defer windows.CloseHandle(snapshot)

	var entry windows.ProcessEntry32
	entry.Size = uint32(unsafe.Sizeof(entry))
	var apps []ProcessInfo

	currentSessionID, err := s.systemManager.GetCurrentSessionID()
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

func (s *windowsApplicationManager) BlockApplication(processName string) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	fmt.Printf("Attempting to block process: %s\n", processName)
	handles, err := getProcessHandles(s.systemManager, processName)
	if err != nil {
		return err
	}
	for _, handle := range handles {
		fmt.Printf("Suspending process handle: %v\n", handle)
		err := suspendProcess(handle)
		if err != nil {
			fmt.Printf("Failed to block process %s: %v\n", processName, err)
			return err
		}
	}
	return nil
}

func (s *windowsApplicationManager) UnblockApplication(processName string) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	fmt.Printf("Attempting to unblock process: %s\n", processName)
	handles, err := getProcessHandles(s.systemManager, processName)
	if err != nil {
		return err
	}
	for _, handle := range handles {
		fmt.Printf("Resuming process handle: %v\n", handle)
		err := resumeProcess(handle)
		if err != nil {
			fmt.Printf("Failed to unblock process %s: %v\n", processName, err)
			return err
		}
	}
	return nil
}

func (s *windowsApplicationManager) SuspendProcess(handle windows.Handle) error {
	return suspendProcess(handle)
}

func (s *windowsApplicationManager) ResumeProcess(handle windows.Handle) error {
	return resumeProcess(handle)
}

func (s *windowsApplicationManager) Intersect(a, b []ProcessInfo) []ProcessInfo {
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

func (s *windowsApplicationManager) EqualProcessSlices(a, b []ProcessInfo) bool {
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

func (s *windowsApplicationManager) GetProcessHandles(processName string) ([]windows.Handle, error) {
	return getProcessHandles(s.systemManager, processName)
}

func (s *windowsApplicationManager) GetProcessSessionID(processID uint32) (uint32, error) {
	var sessionID uint32
	err := windows.ProcessIdToSessionId(processID, &sessionID)
	if err != nil {
		return 0, err
	}
	return sessionID, nil
}

package services

import (
	"fmt"
	"sync"
	"unsafe"

	"github.com/fatih/color"
	"golang.org/x/sys/windows"
)

type ApplicationManager interface {
	BlockApplication(processName string) error
	UnblockApplication(processName string) error
	SuspendProcess(handle windows.Handle) error
	ResumeProcess(handle windows.Handle) error
	ListApplications() ([]string, error)
	Intersect(a, b []string) []string
	EqualStringSlices(a, b []string) bool
	GetProcessHandles(processName string) ([]windows.Handle, error)
	GetProcessSessionID(processID uint32) (uint32, error)
}

type windowsApplicationManager struct {
	mu sync.Mutex
}

func NewWindowsApplicationManager() ApplicationManager {
	return &windowsApplicationManager{}
}

func GetCurrentSessionID() (uint32, error) {
	var sessionID uint32
	processID := windows.GetCurrentProcessId()
	err := windows.ProcessIdToSessionId(processID, &sessionID)
	if err != nil {
		return 0, err
	}
	return sessionID, nil
}

func suspendProcess(handle windows.Handle) error {
	color.Red("Suspendiendo proceso")
	ntSuspendProcess := windows.NewLazySystemDLL("ntdll.dll").NewProc("NtSuspendProcess")
	_, _, err := ntSuspendProcess.Call(uintptr(handle))
	if err != nil && err.Error() != "The operation completed successfully." {
		return fmt.Errorf("failed to suspend process: %v", err)
	}
	return nil
}

func resumeProcess(handle windows.Handle) error {
	color.Green("Reanudando proceso")
	ntResumeProcess := windows.NewLazySystemDLL("ntdll.dll").NewProc("NtResumeProcess")
	_, _, err := ntResumeProcess.Call(uintptr(handle))
	if err != nil && err.Error() != "The operation completed successfully." {
		return fmt.Errorf("failed to resume process: %v", err)
	}
	return nil
}

func getProcessHandles(processName string) ([]windows.Handle, error) {
	snapshot, err := windows.CreateToolhelp32Snapshot(windows.TH32CS_SNAPPROCESS, 0)
	if err != nil {
		return nil, err
	}
	defer windows.CloseHandle(snapshot)

	var entry windows.ProcessEntry32
	entry.Size = uint32(unsafe.Sizeof(entry))
	var handles []windows.Handle
	currentSessionID, err := GetCurrentSessionID()
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
					continue
				}
				handles = append(handles, handle)
			}
		}
	}

	if len(handles) == 0 {
		return nil, fmt.Errorf("no instances of process %s found in the current session", processName)
	}

	return handles, nil
}

func (s *windowsApplicationManager) ListApplications() ([]string, error) {
	snapshot, err := windows.CreateToolhelp32Snapshot(windows.TH32CS_SNAPPROCESS, 0)
	if err != nil {
		return nil, err
	}
	defer windows.CloseHandle(snapshot)

	var entry windows.ProcessEntry32
	entry.Size = uint32(unsafe.Sizeof(entry))
	var apps []string

	currentSessionID, err := GetCurrentSessionID()
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
			apps = append(apps, windows.UTF16ToString(entry.ExeFile[:]))
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
	handles, err := getProcessHandles(processName)
	if err != nil {
		return err
	}
	for _, handle := range handles {
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
	handles, err := getProcessHandles(processName)
	if err != nil {
		return err
	}
	for _, handle := range handles {
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

func (s *windowsApplicationManager) Intersect(a, b []string) []string {
	m := make(map[string]bool)
	for _, item := range a {
		m[item] = true
	}
	var result []string
	for _, item := range b {
		if _, ok := m[item]; ok {
			result = append(result, item)
		}
	}
	return result
}

func (s *windowsApplicationManager) EqualStringSlices(a, b []string) bool {
	if len(a) != len(b) {
		return false
	}
	m := make(map[string]int)
	for _, item := range a {
		m[item]++
	}
	for _, item := range b {
		m[item]--
		if m[item] < 0 {
			return false
		}
	}
	return true
}

func (s *windowsApplicationManager) GetProcessHandles(processName string) ([]windows.Handle, error) {
	return getProcessHandles(processName)
}

func (s *windowsApplicationManager) GetProcessSessionID(processID uint32) (uint32, error) {
	var sessionID uint32
	err := windows.ProcessIdToSessionId(processID, &sessionID)
	if err != nil {
		return 0, err
	}
	return sessionID, nil
}

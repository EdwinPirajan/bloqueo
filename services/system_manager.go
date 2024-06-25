package services

import (
	"fmt"
	"unsafe"

	"golang.org/x/sys/windows"
)

const (
	WTS_CURRENT_SERVER_HANDLE = 0
	WTS_CURRENT_SERVER_NAME   = ""
	WTS_CURRENT_SESSION       = 0xFFFFFFFF

	WTSSessionStateActive = 0x00000000
	WTSSessionStateLocked = 0x00000005
)

type WTS_SESSION_INFO struct {
	SessionId      uint32
	WinStationName *uint16
	State          uint32
}

var (
	modwtsapi32              = windows.NewLazySystemDLL("wtsapi32.dll")
	procWTSEnumerateSessions = modwtsapi32.NewProc("WTSEnumerateSessionsW")
	procWTSFreeMemory        = modwtsapi32.NewProc("WTSFreeMemory")
)

func WTSEnumerateSessions(serverHandle windows.Handle, reserved, version uint32, sessionInfo **WTS_SESSION_INFO, count *uint32) error {
	ret, _, err := procWTSEnumerateSessions.Call(
		uintptr(serverHandle),
		uintptr(reserved),
		uintptr(version),
		uintptr(unsafe.Pointer(sessionInfo)),
		uintptr(unsafe.Pointer(count)),
	)
	if ret == 0 {
		return err
	}
	return nil
}

func WTSFreeMemory(memory *WTS_SESSION_INFO) {
	procWTSFreeMemory.Call(uintptr(unsafe.Pointer(memory)))
}

type SystemManager interface {
	EnableDebugPrivilege() error
	GetCurrentSessionID() (uint32, error)
	GetSessionMutexName() (string, error)
	CheckForDuplicateInstance() (windows.Handle, error)
	GetCurrentActiveSessionID() (uint32, error)
}

type windowsSystemManager struct{}

func NewWindowsSystemManager() SystemManager {
	return &windowsSystemManager{}
}

func (s *windowsSystemManager) EnableDebugPrivilege() error {
	var hToken windows.Token
	err := windows.OpenProcessToken(windows.CurrentProcess(), windows.TOKEN_ADJUST_PRIVILEGES|windows.TOKEN_QUERY, &hToken)
	if err != nil {
		return err
	}
	defer hToken.Close()

	var tkp windows.Tokenprivileges
	tkp.PrivilegeCount = 1
	tkp.Privileges[0].Attributes = windows.SE_PRIVILEGE_ENABLED

	name, err := windows.UTF16PtrFromString("SeDebugPrivilege")
	if err != nil {
		return err
	}
	err = windows.LookupPrivilegeValue(nil, name, &tkp.Privileges[0].Luid)
	if err != nil {
		return err
	}

	err = windows.AdjustTokenPrivileges(hToken, false, &tkp, 0, nil, nil)
	if err != nil {
		return err
	}

	return nil
}

func (s *windowsSystemManager) GetCurrentSessionID() (uint32, error) {
	var sessionID uint32
	processID := windows.GetCurrentProcessId()
	err := windows.ProcessIdToSessionId(processID, &sessionID)
	if err != nil {
		return 0, err
	}
	return sessionID, nil
}

func (s *windowsSystemManager) GetSessionMutexName() (string, error) {
	sessionID, err := s.GetCurrentSessionID()
	if err != nil {
		return "", err
	}
	return fmt.Sprintf("ScrapeBlockerMutex_Session_%d", sessionID), nil
}

func (s *windowsSystemManager) CheckForDuplicateInstance() (windows.Handle, error) {
	mutexName, err := s.GetSessionMutexName()
	if err != nil {
		return 0, err
	}
	mutexHandle, err := windows.CreateMutex(nil, false, windows.StringToUTF16Ptr(mutexName))
	if err != nil {
		return 0, err
	}
	if err = windows.GetLastError(); err == windows.ERROR_ALREADY_EXISTS {
		return 0, fmt.Errorf("another instance of the application is already running in this session")
	}
	return mutexHandle, nil
}

func (s *windowsSystemManager) GetCurrentActiveSessionID() (uint32, error) {
	var sessionInfo *WTS_SESSION_INFO
	var count uint32

	err := WTSEnumerateSessions(WTS_CURRENT_SERVER_HANDLE, 0, 1, &sessionInfo, &count)
	if err != nil {
		return 0, err
	}
	defer WTSFreeMemory(sessionInfo)

	sessions := unsafe.Slice(sessionInfo, count)
	for _, session := range sessions {
		if session.State == WTSSessionStateActive {
			return session.SessionId, nil
		}
	}

	return 0, fmt.Errorf("no active session found")
}

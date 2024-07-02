package services

import (
	"golang.org/x/sys/windows"
)

type SystemManager interface {
	EnableDebugPrivilege() error
	GetCurrentSessionID() (uint32, error)
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

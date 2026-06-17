//go:build windows

package pool

import (
	"fmt"
	"os/exec"
	"syscall"
	"unsafe"
)

const (
	jobObjectExtendedLimitInformationClass = 9
	jobObjectLimitKillOnJobClose           = 0x2000
	processSetQuota                        = 0x0100
	processTerminate                       = 0x0001
	processQueryLimitedInformation         = 0x1000
)

var (
	kernel32                    = syscall.NewLazyDLL("kernel32.dll")
	procCreateJobObjectW        = kernel32.NewProc("CreateJobObjectW")
	procSetInformationJobObject = kernel32.NewProc("SetInformationJobObject")
	procQueryInformationJob     = kernel32.NewProc("QueryInformationJobObject")
	procAssignProcessToJob      = kernel32.NewProc("AssignProcessToJobObject")
	procCloseHandle             = kernel32.NewProc("CloseHandle")
	procGlobalMemoryStatusEx    = kernel32.NewProc("GlobalMemoryStatusEx")
)

type memoryStatusEx struct {
	Length               uint32
	MemoryLoad           uint32
	TotalPhys            uint64
	AvailPhys            uint64
	TotalPageFile        uint64
	AvailPageFile        uint64
	TotalVirtual         uint64
	AvailVirtual         uint64
	AvailExtendedVirtual uint64
}

type ioCounters struct {
	ReadOperationCount  uint64
	WriteOperationCount uint64
	OtherOperationCount uint64
	ReadTransferCount   uint64
	WriteTransferCount  uint64
	OtherTransferCount  uint64
}

type jobObjectBasicLimitInformation struct {
	PerProcessUserTimeLimit int64
	PerJobUserTimeLimit     int64
	LimitFlags              uint32
	MinimumWorkingSetSize   uintptr
	MaximumWorkingSetSize   uintptr
	ActiveProcessLimit      uint32
	Affinity                uintptr
	PriorityClass           uint32
	SchedulingClass         uint32
}

type jobObjectExtendedLimitInformation struct {
	BasicLimitInformation jobObjectBasicLimitInformation
	IoInfo                ioCounters
	ProcessMemoryLimit    uintptr
	JobMemoryLimit        uintptr
	PeakProcessMemoryUsed uintptr
	PeakJobMemoryUsed     uintptr
}

type processMonitor struct {
	job uintptr
}

func attachProcessMonitor(cmd *exec.Cmd) (processMonitor, error) {
	if cmd.Process == nil {
		return processMonitor{}, fmt.Errorf("process monitor requires a started process")
	}
	job, _, err := procCreateJobObjectW.Call(0, 0)
	if job == 0 {
		return processMonitor{}, fmt.Errorf("create Windows job object: %w", err)
	}
	info := jobObjectExtendedLimitInformation{}
	info.BasicLimitInformation.LimitFlags = jobObjectLimitKillOnJobClose
	ok, _, err := procSetInformationJobObject.Call(
		job,
		uintptr(jobObjectExtendedLimitInformationClass),
		uintptr(unsafe.Pointer(&info)),
		unsafe.Sizeof(info),
	)
	if ok == 0 {
		procCloseHandle.Call(job)
		return processMonitor{}, fmt.Errorf("set Windows job object limits: %w", err)
	}
	process, err := syscall.OpenProcess(processSetQuota|processTerminate|processQueryLimitedInformation, false, uint32(cmd.Process.Pid))
	if err != nil {
		procCloseHandle.Call(job)
		return processMonitor{}, fmt.Errorf("open process for Windows job object: %w", err)
	}
	defer syscall.CloseHandle(process)
	ok, _, err = procAssignProcessToJob.Call(job, uintptr(process))
	if ok == 0 {
		procCloseHandle.Call(job)
		return processMonitor{}, fmt.Errorf("assign process to Windows job object: %w", err)
	}
	return processMonitor{job: job}, nil
}

func (m processMonitor) close() {
	if m.job != 0 {
		procCloseHandle.Call(m.job)
	}
}

func (m processMonitor) peakJobMemoryBytes() int64 {
	if m.job == 0 {
		return 0
	}
	var snapshot jobObjectExtendedLimitInformation
	ok, _, _ := procQueryInformationJob.Call(
		m.job,
		uintptr(jobObjectExtendedLimitInformationClass),
		uintptr(unsafe.Pointer(&snapshot)),
		unsafe.Sizeof(snapshot),
		0,
	)
	if ok == 0 {
		return 0
	}
	return int64(snapshot.PeakJobMemoryUsed)
}

type systemMemoryMonitor struct{}

func (systemMemoryMonitor) Status() (MemoryStatus, error) {
	status := memoryStatusEx{Length: uint32(unsafe.Sizeof(memoryStatusEx{}))}
	ok, _, err := procGlobalMemoryStatusEx.Call(uintptr(unsafe.Pointer(&status)))
	if ok == 0 {
		return MemoryStatus{}, fmt.Errorf("GlobalMemoryStatusEx: %w", err)
	}
	return MemoryStatus{
		TotalMemoryMB: int(status.TotalPhys / 1024 / 1024),
		FreeMemoryMB:  int(status.AvailPhys / 1024 / 1024),
		TotalCommitMB: int(status.TotalPageFile / 1024 / 1024),
		FreeCommitMB:  int(status.AvailPageFile / 1024 / 1024),
	}, nil
}

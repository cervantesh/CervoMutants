//go:build !windows

package pool

import "os/exec"

type processMonitor struct{}

func attachProcessMonitor(_ *exec.Cmd) (processMonitor, error) {
	return processMonitor{}, nil
}

func (processMonitor) close()                    {}
func (processMonitor) peakJobMemoryBytes() int64 { return 0 }

type systemMemoryMonitor struct{}

func (systemMemoryMonitor) Status() (MemoryStatus, error) {
	return MemoryStatus{}, nil
}

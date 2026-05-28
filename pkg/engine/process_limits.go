package engine

import (
	"errors"

	"gitea.cervbox.synology.me/CervoSoft/cervo-mutant/pkg/config"
)

var errProcessLimitUnsupported = errors.New("process resource limits are not supported on this platform")

func hasProcessLimits(resources config.Resources) bool {
	return resources.MaxProcessMemoryMB > 0 || resources.MaxProcesses > 0
}

func noopProcessLimitCleanup() {}

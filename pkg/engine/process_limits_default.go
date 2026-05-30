//go:build !windows

package engine

import (
	"os/exec"

	"gitea.cervbox.synology.me/CervoSoft/cervo-mutant/pkg/config"
)

func applyProcessLimits(cmd *exec.Cmd, resources config.Resources) (func(), error) {
	return noopProcessLimitCleanup, nil
}

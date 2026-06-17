//go:build !windows

package engine

import (
	"os/exec"

	"github.com/cervantesh/cervo-mutant/pkg/config"
)

func applyProcessLimits(cmd *exec.Cmd, resources config.Resources) (func(), error) {
	return noopProcessLimitCleanup, nil
}

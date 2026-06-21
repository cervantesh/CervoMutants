package engine

import (
	"errors"

	"github.com/cervantesh/cervo-mutants/pkg/internal/gotestenv"
	internalmutation "github.com/cervantesh/cervo-mutants/pkg/internal/mutationfs"
	"github.com/cervantesh/cervo-mutants/pkg/isolate"
)

func (s *runSession) prepareMutation(mutant Mutant, command []string) (string, []string, func(), error) {
	if s.engine.cfg.Execution.Isolation == "overlay" {
		return prepareOverlayMutation(mutant, command, s.engine.cfg.Execution.TempRoot)
	}
	workdir, err := isolate.CopyModuleWithRoot(mutant.Module, s.engine.cfg.Execution.TempRoot)
	if err != nil {
		return "", nil, noopCleanup, err
	}
	cleanup := func() { _ = isolate.Cleanup(workdir) }
	targetFile, err := isolate.ContainedTargetPath(mutant.Module, workdir, mutant.File)
	if err != nil {
		cleanup()
		return "", nil, noopCleanup, err
	}
	if err := applyDiffReplacement(targetFile, mutant); err != nil {
		cleanup()
		return "", nil, noopCleanup, err
	}
	return workdir, command, cleanup, nil
}

func prepareOverlayMutation(mutant Mutant, command []string, tempRoot string) (string, []string, func(), error) {
	workdir, overlayPath, cleanup, err := internalmutation.PrepareOverlay(mutant.Module, mutant.File, mutant.StartOffset, mutant.EndOffset, mutant.Original, mutant.Mutated, tempRoot)
	if err != nil {
		return "", nil, noopCleanup, err
	}
	return workdir, gotestenv.WithOverlayFlag(command, overlayPath), cleanup, nil
}

func noopCleanup() {
	// No temporary resources were allocated before the failure.
}

func applyDiffReplacement(path string, mutant Mutant) error {
	if mutant.StartOffset < 0 || mutant.EndOffset <= mutant.StartOffset {
		return errors.New("mutant patch offsets are invalid")
	}
	return internalmutation.ApplyReplacement(path, mutant.StartOffset, mutant.EndOffset, mutant.Original, mutant.Mutated)
}

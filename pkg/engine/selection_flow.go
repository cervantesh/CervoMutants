package engine

import "github.com/cervantesh/cervo-mutants/pkg/internal/gotestenv"

func (s *runSession) selectTests(mutant Mutant) TestPlan {
	command := append([]string{}, s.engine.cfg.Tests.Command...)
	if len(command) == 0 {
		command = []string{"go", "test", "./..."}
	}
	lineCovered, fileCovered := s.coverageSignal(mutant)
	if s.engine.cfg.Selection.Prefilter && !fileCovered {
		return TestPlan{Command: command, Reason: "coverage prefilter did not match mutant file", CoversMutant: false, CoverageSource: "package-mode-prefilter"}
	}
	if s.engine.cfg.Selection.Mode == "package" && gotestenv.IsGoTestCommand(command) && mutant.Package != "" {
		command = gotestenv.PackageScopedCommand(command, mutant.Package)
		source := "unknown"
		if s.engine.cfg.Selection.Prefilter {
			source = "package-mode-prefilter"
		}
		return TestPlan{Command: command, Reason: "package selected from mutant file", CoversMutant: true, CoverageSource: source}
	}
	if s.engine.cfg.Selection.Mode == "coverage" {
		if lineCovered && gotestenv.IsGoTestCommand(command) && mutant.Package != "" {
			command = gotestenv.PackageScopedCommand(command, mutant.Package)
			return TestPlan{Command: command, Reason: "coverage profile matched mutant line", CoversMutant: true, CoverageSource: "coverage-mode"}
		}
		if fileCovered && gotestenv.IsGoTestCommand(command) && mutant.Package != "" {
			command = gotestenv.PackageScopedCommand(command, mutant.Package)
			return TestPlan{Command: command, Reason: "coverage profile matched mutant file; package fallback selected", CoversMutant: true, CoverageSource: "coverage-mode-file-fallback"}
		}
		return TestPlan{Command: command, Reason: "coverage profile did not match mutant file", CoversMutant: false, CoverageSource: "coverage-mode"}
	}
	return TestPlan{Command: command, Reason: "all tests selected", CoversMutant: true, CoverageSource: "unknown"}
}

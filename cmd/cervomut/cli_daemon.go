package main

import (
	"context"
	"flag"
	"fmt"
	"io"
	"os"

	"github.com/cervantesh/cervo-mutants/pkg/daemon"
	"github.com/cervantesh/cervo-mutants/pkg/engine"
)

const daemonExperimentalEnvVar = "CERVOMUT_EXPERIMENTAL_DAEMON"

var serveDaemonFn = func(ctx context.Context, in io.Reader, out io.Writer, runner engine.Runner) error {
	return daemon.ServeJSONLines(ctx, in, out, runner)
}

func backgroundContext() context.Context {
	return context.Background()
}

func cmdDaemonMode(command string, args []string) error {
	fs := flag.NewFlagSet(command, flag.ContinueOnError)
	experimental := fs.Bool("experimental", false, "acknowledge that this daemon/worker protocol is experimental and unsupported")
	maxOutputBytes := fs.Int("max-output-bytes", 12000, "maximum bytes of test output captured per mutant result")
	if err := fs.Parse(reorderFlags(args, map[string]bool{"max-output-bytes": true})); err != nil {
		if err == flag.ErrHelp {
			return nil
		}
		return err
	}
	if !*experimental && os.Getenv(daemonExperimentalEnvVar) != "1" {
		return fmt.Errorf("%s is experimental and not a supported compatibility surface yet; rerun with --experimental or set %s=1", command, daemonExperimentalEnvVar)
	}
	return serveDaemonFn(backgroundContext(), os.Stdin, os.Stdout, daemon.WorkerRunner{MaxOutputBytes: *maxOutputBytes})
}

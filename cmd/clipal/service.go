package main

import (
	"context"
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/lansespirit/Clipal/internal/config"
	"github.com/lansespirit/Clipal/internal/service"
)

func runService(args []string) {
	fs := flag.NewFlagSet("service", flag.ContinueOnError)
	fs.SetOutput(os.Stderr)

	configDir := fs.String("config-dir", "", "Configuration directory (default: ~/.clipal)")
	binaryPath := fs.String("bin", "", "Path to clipal binary (default: current executable)")
	force := fs.Bool("force", false, "Reinstall/update the system service if it already exists")
	dryRun := fs.Bool("dry-run", false, "Print actions without executing them")
	timeout := fs.Duration("timeout", 30*time.Second, "Overall timeout for service manager commands")

	// macOS launchd (optional)
	stdoutPath := fs.String("stdout", "", "macOS: launchd StandardOutPath (optional)")
	stderrPath := fs.String("stderr", "", "macOS: launchd StandardErrorPath (optional)")

	if err := fs.Parse(args); err != nil {
		os.Exit(2)
	}
	rest := fs.Args()
	if len(rest) < 1 {
		printServiceUsage()
		os.Exit(2)
	}

	action, err := service.ParseAction(rest[0])
	if err != nil {
		fmt.Fprintf(os.Stderr, "clipal service: %v\n", err)
		printServiceUsage()
		os.Exit(2)
	}

	cfgDir := *configDir
	if cfgDir == "" {
		cfgDir = config.GetConfigDir()
	}

	bin := *binaryPath
	if bin == "" {
		exe, exeErr := os.Executable()
		if exeErr != nil {
			fmt.Fprintf(os.Stderr, "clipal service: failed to determine executable path: %v\n", exeErr)
			os.Exit(1)
		}
		if resolved, resolvedErr := filepath.EvalSymlinks(exe); resolvedErr == nil {
			exe = resolved
		}
		bin = exe
	}

	mgr := service.DefaultManager()
	opts := service.Options{
		ConfigDir:  cfgDir,
		BinaryPath: bin,
		Force:      *force,
		DryRun:     *dryRun,
		StdoutPath: *stdoutPath,
		StderrPath: *stderrPath,
	}

	plan, err := mgr.Plan(action, opts)
	if err != nil {
		fmt.Fprintf(os.Stderr, "clipal service %s failed: %v\n", action, err)
		os.Exit(1)
	}

	ctx, cancel := context.WithTimeout(context.Background(), *timeout)
	defer cancel()

	out, err := service.ExecutePlan(ctx, plan, opts.DryRun)
	if err != nil {
		if out != "" {
			fmt.Fprint(os.Stderr, out)
		}
		fmt.Fprintf(os.Stderr, "clipal service %s failed: %v\n", action, err)
		os.Exit(1)
	}

	if out != "" {
		fmt.Fprint(os.Stdout, out)
	}
}

func printServiceUsage() {
	fmt.Fprintln(os.Stderr, "usage: clipal service [flags] <install|uninstall|start|stop|restart|status>")
	fmt.Fprintln(os.Stderr, "")
	fmt.Fprintln(os.Stderr, "examples:")
	fmt.Fprintln(os.Stderr, "  clipal service install --config-dir ~/.clipal")
	fmt.Fprintln(os.Stderr, "  clipal service restart")
	fmt.Fprintln(os.Stderr, "  clipal service status")
}

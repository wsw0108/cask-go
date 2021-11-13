package main

import (
	"bytes"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
)

func main() {
	var emacs string
	emacs = os.Getenv("CASK_EMACS")
	if emacs == "" {
		emacs = os.Getenv("EMACS")
	}
	if emacs == "" {
		emacs = "emacs"
	}
	// cask := os.Args[0]
	var subcommand string
	var extras []string
	if len(os.Args) > 1 {
		subcommand = os.Args[1]
		extras = os.Args[2:]
	} else {
		subcommand = "install"
	}
	if subcommand == "exec" {
		if len(os.Args) < 3 {
			// TODO: emulate
			fmt.Println("no enough args")
			os.Exit(-1)
		}
		extras = os.Args[3:]
	}
	script, err := scriptPath()
	if err != nil {
		os.Exit(-1)
	}
	// TODO: subcommand to bootstrap/clone cask to some where
	switch subcommand {
	case "emacs", "exec":
		var cmd *exec.Cmd
		if subcommand == "emacs" {
			cmd = exec.Command(emacs, extras...)
		} else if subcommand == "exec" {
			cmd = exec.Command(os.Args[2], extras...)
		}
		// FIXME: copy all envs first
		envHome := os.Getenv("HOME")
		if envHome != "" {
			cmd.Env = append(cmd.Env, fmt.Sprintf("HOME=%s", envHome))
		}
		cmd.Env = append(cmd.Env, fmt.Sprintf("TMPDIR=%s", os.TempDir()))
		cmd.Env = append(cmd.Env, fmt.Sprintf("EMACS=%s", emacs))
		{
			var buf bytes.Buffer
			c, err := runCaskCli(emacs, script, "load-path", nil, &buf, nil)
			if err != nil {
				os.Exit(c.ProcessState.ExitCode())
			}
			value := strings.TrimSuffix(buf.String(), "\r\n")
			cmd.Env = append(cmd.Env, fmt.Sprintf("EMACSLOADPATH=%s", value))
		}
		{
			var buf bytes.Buffer
			c, err := runCaskCli(emacs, script, "path", nil, &buf, nil)
			if err != nil {
				os.Exit(c.ProcessState.ExitCode())
			}
			value := strings.TrimSuffix(buf.String(), "\r\n")
			cmd.Env = append(cmd.Env, fmt.Sprintf("PATH=%s", value))
		}
		cmd.Stdout = os.Stdout
		cmd.Stderr = os.Stderr
		if err := cmd.Run(); err != nil {
			os.Exit(cmd.ProcessState.ExitCode())
		}
	default:
		cmd, err := runCaskCli(emacs, script, subcommand, extras, os.Stdout, os.Stderr)
		if err != nil {
			os.Exit(cmd.ProcessState.ExitCode())
		}
	}
}

func scriptPath() (string, error) {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return "", err
	}
	srcDir := filepath.Join(homeDir, ".cask")
	script := filepath.Join(srcDir, "cask-cli.el")
	return script, nil
}

func runCaskCli(emacs string, script string, subcommand string, args []string, stdout io.Writer, stderr io.Writer) (*exec.Cmd, error) {
	var argv []string
	argv = append(argv, "-Q", "--script", script, "--", subcommand)
	argv = append(argv, args...)
	cmd := exec.Command(emacs, argv...)
	cmd.Stdout = stdout
	cmd.Stderr = stderr
	return cmd, cmd.Run()
}

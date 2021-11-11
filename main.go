package main

import (
	"bytes"
	"fmt"
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
	cask := os.Args[0]
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
			fmt.Println("no enough args")
			os.Exit(-1)
		}
		extras = os.Args[3:]
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
			c := exec.Command(cask, "load-path")
			var buf bytes.Buffer
			c.Stdout = &buf
			if err := c.Run(); err != nil {
				fmt.Println(err)
				os.Exit(-1)
			}
			value := strings.TrimSuffix(buf.String(), "\r\n")
			cmd.Env = append(cmd.Env, fmt.Sprintf("EMACSLOADPATH=%s", value))
		}
		{
			c := exec.Command(cask, "path")
			var buf bytes.Buffer
			c.Stdout = &buf
			if err := c.Run(); err != nil {
				fmt.Println(err)
				os.Exit(-1)
			}
			value := strings.TrimSuffix(buf.String(), "\r\n")
			cmd.Env = append(cmd.Env, fmt.Sprintf("PATH=%s", value))
		}
		cmd.Stdout = os.Stdout
		cmd.Stderr = os.Stderr
		if err := cmd.Run(); err != nil {
			fmt.Println(err)
			os.Exit(-1)
		}
	default:
		homeDir, err := os.UserHomeDir()
		if err != nil {
			fmt.Println(err)
			os.Exit(-1)
		}
		srcDir := filepath.Join(homeDir, ".cask")
		script := filepath.Join(srcDir, "cask-cli.el")
		var args []string
		args = append(args, "-Q", "--script", script, "--", subcommand)
		args = append(args, extras...)
		cmd := exec.Command(emacs, args...)
		cmd.Stdout = os.Stdout
		cmd.Stderr = os.Stderr
		if err := cmd.Run(); err != nil {
			fmt.Println(err)
			os.Exit(-1)
		}
	}
}

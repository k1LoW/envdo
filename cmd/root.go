/*
Copyright © 2025 Ken'ichiro Oyama <k1lowxb@gmail.com>

Permission is hereby granted, free of charge, to any person obtaining a copy
of this software and associated documentation files (the "Software"), to deal
in the Software without restriction, including without limitation the rights
to use, copy, modify, merge, publish, distribute, sublicense, and/or sell
copies of the Software, and to permit persons to whom the Software is
furnished to do so, subject to the following conditions:

The above copyright notice and this permission notice shall be included in
all copies or substantial portions of the Software.

THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE
AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM,
OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN
THE SOFTWARE.
*/
package cmd

import (
	"errors"
	"fmt"
	"os"

	"os/exec"

	"github.com/k1LoW/envdo/env"
	"github.com/k1LoW/envdo/version"
	"github.com/spf13/cobra"
)

var profile string

// rootCmd represents the base command when called without any subcommands.
var rootCmd = &cobra.Command{
	Use:   "envdo",
	Short: "Execute commands with environment variables from .env files",
	Long: `envdo loads environment variables from .env files and executes commands with them.

It searches for .env files in the current directory and $XDG_CONFIG_HOME/envdo directory.
Current directory values take priority over config directory values.

Examples:
  envdo -- echo $MY_VAR
  envdo --profile production -- node app.js
  envdo -p dev -- npm start`,
	SilenceUsage: true,
	Version:      version.Version,
	RunE: func(cmd *cobra.Command, args []string) error {
		// Load environment variables
		envs, err := env.LoadEnvFiles(profile)
		if err != nil {
			return fmt.Errorf("failed to load environment variables: %w", err)
		}

		// If no arguments, print the loaded environment variables
		if len(args) == 0 {
			for key, value := range envs {
				fmt.Printf("export %s=%s\n", key, value)
			}
			return nil
		}

		// Prepare environment for command execution
		cmdEnvs := os.Environ()
		for key, value := range envs {
			cmdEnvs = append(cmdEnvs, fmt.Sprintf("%s=%s", key, value))
		}

		// Execute the command
		command := args[0]
		c := exec.Command(command, args[1:]...)
		c.Stdin = os.Stdin
		c.Stdout = os.Stdout
		c.Stderr = os.Stderr
		c.Env = cmdEnvs
		cmd.SilenceErrors = true
		if err := c.Run(); err != nil {
			var exitError *exec.ExitError
			if errors.As(err, &exitError) {
				os.Exit(exitError.ExitCode())
			}
			return err
		}
		return nil
	},
}

// Execute adds all child commands to the root command and sets flags appropriately.
// This is called by main.main(). It only needs to happen once to the rootCmd.
func Execute() {
	if err := rootCmd.Execute(); err != nil {
		os.Exit(1)
	}
}

func init() {
	rootCmd.Flags().StringVarP(&profile, "profile", "p", "", "profile name")
}

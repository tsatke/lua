package main

import (
	"fmt"
	"github.com/spf13/cobra"
	"github.com/tsatke/lua"
	"os"
)

var (
	// Version can be set with the Go linker.
	Version string = "master"
	// AppName is the name of this app, as displayed in the help
	// text of the root command.
	AppName = "lua"
)

var (
	rootCmd = &cobra.Command{
		Use:  AppName,
		Args: cobra.MinimumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			script := args[0]

			wd, err := os.Getwd()
			if err != nil {
				return err
			}

			e := lua.NewEngine(
				lua.WithStdin(os.Stdin),
				lua.WithStdout(os.Stdout),
				lua.WithStderr(os.Stderr),
				lua.WithScannerType(lua.ScannerTypeInMemory),
				lua.WithWorkingDirectory(wd),
			)

			_, err = e.EvalFile(script)
			if err != nil {
				return err
			}

			return nil
		},
	}
)

func main() {
	if err := rootCmd.Execute(); err != nil {
		_, _ = fmt.Fprintf(os.Stderr, "%s", err)
		os.Exit(1)
	}
}

/*
Copyright © 2026 JACOB ARTHURS
*/
package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
)

var (
	flagPyPIPackage string
)

var pypiCmd = &cobra.Command{
	Use:   "pypi",
	Short: "Publish binaries to PyPI",
	RunE: func(cmd *cobra.Command, args []string) error {
		return fmt.Errorf("pypi publishing not yet implemented")
	},
}

func init() {
	pypiCmd.Flags().StringVar(&flagPyPIPackage, "package", "", "PyPI package name")

	pypiCmd.MarkFlagRequired("package")
}

/*
Copyright © 2024 NAME HERE <EMAIL ADDRESS>
*/
package cmd

import (
	"os"
	"regexp"
	"strings"

	"maze/cmd/ux"

	"github.com/fatih/color"
	"github.com/spf13/cobra"
)

// rootCmd represents the base command when called without any subcommands
var rootCmd = &cobra.Command{
	Use:   "maze",
	Short: ux.ShortTextCli,
	Long:  ux.LongTextCli,

	Run: func(cmd *cobra.Command, args []string) {
		cmd.Help()
	},
}

// This is called by main.main(). It only needs to happen once to the rootCmd.
func Execute() {
	err := rootCmd.Execute()
	if err != nil {
		os.Exit(1)
	}
}
func addSubCommands() {

	rootCmd.AddCommand(planCmd)
	rootCmd.AddCommand(configureCmd)
}
func init() {

	addSubCommands()
	rootCmd.CompletionOptions.DisableDefaultCmd = true
	rootCmd.SetOutput(color.Output)
	cobra.AddTemplateFunc("StyleHeading", color.RGB(5, 98, 254).SprintFunc())
	usageTemplate := rootCmd.UsageTemplate()

	usageTemplate = strings.NewReplacer(
		`Usage:`, `{{StyleHeading "Usage:"}}`,
		`Aliases:`, `{{StyleHeading "Aliases:"}}`,
		`Available Commands:`, `{{StyleHeading "Available Commands:"}}`,
		`Global Flags:`, `{{StyleHeading "Global Flags:"}}`,
		// The following one steps on "Global Flags:"
		`Flags:`, `{{StyleHeading "Flags:"}}`,
	).Replace(usageTemplate)
	re := regexp.MustCompile(`(?m)^Flags:\s*$`)
	usageTemplate = re.ReplaceAllLiteralString(usageTemplate, `{{StyleHeading "Flags:"}}`)
	rootCmd.SetUsageTemplate(usageTemplate)

}

/*
Copyright © 2024 NAME HERE <EMAIL ADDRESS>
*/
package cmd

import (
	"fmt"
	"regexp"
	"strings"

	"maze/cmd/ux"

	"github.com/fatih/color"
	"github.com/spf13/cobra"
)

var (
	listProfileBool   bool
	deleteProfileBool bool
)

// configureCmd represents the configure command
var configureCmd = &cobra.Command{
	Use:   "configure",
	Short: ux.ShortTextConfigure,
	Long:  ux.LongTextConfigure,
	Run: func(cmd *cobra.Command, args []string) {
		mazeConfigure()
	},
}

func mazeConfigure() error {

	style := color.New()
	style.AddRGB(5, 98, 254)
	style.Add(color.Bold)
	fmt.Println()
	style.Printf(ux.MazeLogo)

	fmt.Println()
	if listProfileBool {
		listProfiles()
	} else if deleteProfileBool {
		deleteProfiles()
	} else {
		addNewProfile()
	}

	return nil
}
func listProfiles() {
	fmt.Println("Profile list: ")

	profileList, err := ux.LoadProfiles()
	if err != nil {
		fmt.Println("Error loading profiles")
	}
	for profileName := range profileList {
		fmt.Println(" -", profileName)
	}
}
func deleteProfiles() {
	fmt.Println("Type a profile to delete: ")
	profileName := ""
	fmt.Scanln(&profileName)
	if profileName == "" {
		fmt.Println("Exiting, no profiles deleted")
	}
	err := ux.DeleteProfile(profileName)
	if err != nil {
		fmt.Println(err)
	}
	return
}

func addNewProfile() {
	profileName := ""
	token := ""

	fmt.Println("Enter profile name (press enter for default):")
	fmt.Scanln(&profileName)
	if profileName == "" {
		profileName = "default"
	}
	fmt.Println("Enter your authentication token:")
	fmt.Scanln(&token)
	if token == "" {
		for ok := true; ok; ok = token == "" {
			fmt.Println("Token cannot be empty")
			fmt.Println("Enter your authentication token:")
			fmt.Scanln(&token)
		}
	}

	profile := ux.Profile{
		ProfileName: profileName,
		AuthToken:   token,
	}

	// Save the profile.
	if err := ux.SaveProfile(profile); err != nil {
		fmt.Println("Error saving profile:", err)
		return
	}
	return
}

func init() {

	configureCmd.SetOutput(color.Output)
	configureCmd.Flags().BoolVarP(&listProfileBool, "list", "l", false, "List the existing configured profiles")
	configureCmd.Flags().BoolVarP(&deleteProfileBool, "delete", "d", false, "delete an existing profiles")

	cobra.AddTemplateFunc("StyleHeading", color.RGB(5, 98, 254).SprintFunc())
	usageTemplate := planCmd.UsageTemplate()

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
	configureCmd.SetUsageTemplate(usageTemplate)

}

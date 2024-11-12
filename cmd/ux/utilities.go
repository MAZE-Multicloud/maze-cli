package ux

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"

	"github.com/fatih/color"
)

// - - - Collection of UX related functions - - -

var BlueColor = color.RGB(5, 98, 254).SprintFunc()

var MazeLogo = BlueColor(`
    __  ______ _____   ______     ________    ____
   /  |/  /   /__  /  / ____/    / ____/ /   /  _/
  / /|_/ / /| | / /  / __/______/ /   / /    / /  
 / /  / / ___ |/ /__/ /__/_____/ /___/ /____/ /   
/_/  /_/_/  |_/____/_____/     \____/_____/___/   

`)

var ShortTextCli = `Maze-cli allows users of maze to interact with the maze api in an easy to use tool!`
var LongTextCli = fmt.Sprintf(`%s

%s`, MazeLogo, ShortTextCli,
)

var ShortTextPlan = `Plan creates a canvas using provided teraform files`
var LongTextPlan = fmt.Sprintf(`%s

%s`, MazeLogo, ShortTextPlan,
)

var ShortTextConfigure = `Configure your authentication credentials`
var LongTextConfigure = fmt.Sprintf(`%s

%s`, MazeLogo, ShortTextConfigure,
)

type Profile struct {
	ProfileName string `json:"profileName"`
	AuthToken   string `json:"authToken"`
}

// GetProfileDir returns the path to the user's home directory.
func GetProfileDir() (string, error) {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return "", fmt.Errorf("could not determine home directory: %w", err)
	}
	filePath := filepath.Join(homeDir, ".maze")

	// Create the hidden folder if it doesn't exist.
	if err := os.MkdirAll(filePath, os.ModePerm); err != nil {
		return "", fmt.Errorf("could not create directory %s: %w", homeDir, err)
	}
	filePath = filepath.Join(filePath, "profiles.json")

	return filePath, nil
}

// GetAuthToken retrieves the auth token for a specific profile.
func GetAuthToken(profileName string) (string, error) {
	// Load existing profiles.
	profiles, err := LoadProfiles()
	if err != nil {
		return "", fmt.Errorf("could not load profiles: %w", err)
	}

	// Retrieve the auth token for the specified profile.
	authToken, exists := profiles[profileName]
	if !exists {
		return "", fmt.Errorf("profile %s not found", profileName)
	}

	return authToken, nil
}

// SaveProfile saves a profile to a file in a hidden folder in the user's home directory.
func SaveProfile(profile Profile) error {

	filePath, err := GetProfileDir()
	if err != nil {
		return err
	}
	// Load existing profiles if the file already exists.
	profiles, err := LoadProfiles()
	if err != nil {
		return fmt.Errorf("could not load existing profiles: %w", err)
	}

	// Add or update the profile in the list of profiles.
	profiles[profile.ProfileName] = profile.AuthToken

	// Marshal the profiles map to JSON.
	jsonData, err := json.MarshalIndent(profiles, "", "  ")
	if err != nil {
		return fmt.Errorf("could not marshal profiles to JSON: %w", err)
	}

	// Write the JSON data to the file.
	if err := os.WriteFile(filePath, jsonData, 0644); err != nil {
		return fmt.Errorf("could not write to file %s: %w", filePath, err)
	}

	fmt.Printf("Profile data successfully saved for profile %s\n", profile.ProfileName)
	return nil
}

// DeleteProfile removes a profile from the profiles.json file by name.
func DeleteProfile(profileName string) error {
	filePath, err := GetProfileDir()
	if err != nil {
		return err
	}

	// Load existing profiles.
	profiles, err := LoadProfiles()
	if err != nil {
		return fmt.Errorf("could not load profiles: %w", err)
	}

	// Check if the profile exists.
	if _, exists := profiles[profileName]; !exists {
		return fmt.Errorf("profile %s does not exist", profileName)
	}

	// Delete the profile from the map.
	delete(profiles, profileName)

	// Save the updated profiles back to the file.
	jsonData, err := json.MarshalIndent(profiles, "", "  ")
	if err != nil {
		return fmt.Errorf("could not marshal profiles to JSON: %w", err)
	}

	if err := os.WriteFile(filePath, jsonData, 0644); err != nil {
		return fmt.Errorf("could not write to file %s: %w", filePath, err)
	}

	fmt.Printf("Profile %s successfully deleted\n", profileName)
	return nil
}

// loadProfiles loads the profiles from the specified file path.
func LoadProfiles() (map[string]string, error) {
	profiles := make(map[string]string)
	filePath, err := GetProfileDir()

	// Check if the profiles file exists.
	if _, err := os.Stat(filePath); os.IsNotExist(err) {
		return profiles, nil // Return empty map if file doesn't exist.
	}

	// Read the profiles file.
	data, err := os.ReadFile(filePath)
	if err != nil {
		return nil, fmt.Errorf("could not read file %s: %w", filePath, err)
	}

	// Unmarshal the JSON data into the profiles map.
	if err := json.Unmarshal(data, &profiles); err != nil {
		return nil, fmt.Errorf("could not unmarshal profiles JSON: %w", err)
	}

	return profiles, nil
}

// Function to print formatted output
func PrintFormatted(text string, styling []string) {
	// Creating map with available styles
	styleMap := map[string]color.Attribute{
		"white":  color.FgHiWhite,
		"blue":   color.FgHiBlue,
		"green":  color.FgHiGreen,
		"red":    color.FgHiRed,
		"yellow": color.FgHiYellow,
		"bold":   color.Bold,
	}

	// Defining new blank color object
	style := color.New()

	// Applying all styles found in "styling" arg
	for _, item := range styling {
		style.Add(styleMap[item])
	}

	// Printing passed text with perviously configured styles
	style.Printf(text)
}

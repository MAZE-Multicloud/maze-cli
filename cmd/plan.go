/*
Copyright © 2024 NAME HERE <EMAIL ADDRESS>
*/
package cmd

import (
	"bytes"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"math"
	"mime/multipart"
	"net/http"
	"os"
	"path/filepath"
	"regexp"
	"strings"

	"errors"
	"maze/cmd/ux"
	"time"

	"github.com/fatih/color"
	"github.com/spf13/cobra"
)

// Define struct to match the JSON structure.
type Check struct {
	CheckID     string `json:"check_id"`
	CheckName   string `json:"check_name"`
	CheckResult struct {
		Result        string   `json:"result"`
		EvaluatedKeys []string `json:"evaluated_keys"`
	} `json:"check_result"`
	FilePath string `json:"file_path"`
	Resource string `json:"resource"`
}

type Results struct {
	FailedChecks []Check `json:"failed_checks"`
}

type Summary struct {
	Passed         int    `json:"passed"`
	Failed         int    `json:"failed"`
	Skipped        int    `json:"skipped"`
	ParsingErrors  int    `json:"parsing_errors"`
	ResourceCount  int    `json:"resource_count"`
	CheckovVersion string `json:"checkov_version"`
}

type Response struct {
	CheckType string  `json:"check_type"`
	Results   Results `json:"results"`
	Summary   Summary `json:"summary"`
}

var (
	dirPath       string
	url           string
	name          string
	generateImage bool
	token         string
	profileName   string
	provider      string
)

// planCmd represents the plan command
var planCmd = &cobra.Command{
	Use:   "plan",
	Short: ux.ShortTextPlan,
	Long:  ux.LongTextPlan,
	Run: func(cmd *cobra.Command, args []string) {
		mazePlan()
	},
}

func mazePlan() error {
	style := color.New()
	style.AddRGB(5, 98, 254)
	style.Add(color.Bold)
	fmt.Println()
	style.Printf(ux.MazeLogo)
	style.Println("Terraform plan")
	fmt.Printf("We are going to create a project in maze using terraform provided from %s and the project will be named %s\n\nUsing "+url+" as the server (to change use -u)\n\n\n", dirPath, name)

	time.Sleep(2000 * time.Millisecond)

	if provider != "aws" && provider != "azure" && provider != "gcp" {
		fmt.Println("Enter the provider you are using - aws/azure/gcp (use --provider to pass in provider):")
		fmt.Scanln(&provider)
		if provider != "aws" && provider != "azure" && provider != "gcp" {
			for ok := true; ok; ok = provider != "aws" && provider != "azure" && provider != "gcp" {
				fmt.Println("Provider needs to be one provided")
				fmt.Println("Enter the provider you are using - aws/azure/gcp:")
				fmt.Scanln(&provider)
			}
		}
	}
	success := authStep()

	if !success {
		return nil
	}

	if provider == "gcp" {
		provider = "google"
	} else if provider == "azure" {
		provider = "azurerm"
	}
	time.Sleep(500 * time.Millisecond)

	body, writer := readFileStep()

	time.Sleep(1000 * time.Millisecond)
	success, path := sendFiles(body, writer)
	folderPath := filepath.Join(dirPath, "maze-output")
	if _, err := os.Stat(folderPath); errors.Is(err, os.ErrNotExist) {
		err := os.Mkdir(folderPath, os.ModePerm)
		if err != nil {
			fmt.Println(err)
		}
	}

	time.Sleep(1000 * time.Millisecond)
	formatStep(string(path))
	time.Sleep(1000 * time.Millisecond)

	validateStep(string(path))

	time.Sleep(1000 * time.Millisecond)

	complianceStep(string(path))

	time.Sleep(1000 * time.Millisecond)

	success, projectIdBytes := planStep(string(path))
	projectId := string(projectIdBytes)

	if !success {
		return nil
	}

	time.Sleep(1000 * time.Millisecond)

	costStep(projectId)

	time.Sleep(1000 * time.Millisecond)
	if generateImage {
		imageStep(projectId, dirPath)
		time.Sleep(2000 * time.Millisecond)
	}

	deleteFilesStep(string(path))

	fmt.Println("")

	fmt.Println("Go to this url to view your canvas:")
	time.Sleep(800 * time.Millisecond)

	style.Println("----------------------------------------------------------------------")

	fmt.Println(url + "/projects/" + projectId + "/canvas")
	style.Println("----------------------------------------------------------------------")

	fmt.Println("")

	return nil
}

func authStep() (success bool) {
	//Auth ----------------------------------------------

	// Create a new HTTP request with the multipart data.
	authReq, err := http.NewRequest("GET", url+"/api/user/", nil)
	if err != nil {
		return false
	}
	token, err = ux.GetAuthToken(profileName)
	if err != nil {
		println(err)
		return false
	}

	//Auth ----------------------------------------------
	authReq.Header.Set("Authorization", token)

	var authStartSpinner = ux.NewSpinner("Authenticating", "Authenticated", "Authentication failed", false)
	authStartSpinner.Start()

	// Execute the request.
	client := &http.Client{}
	resp, err := client.Do(authReq)
	if err != nil {
		authStartSpinner.Fail()
		return false

	}
	// Check the response status.
	if resp.StatusCode != http.StatusOK {
		authStartSpinner.Fail()
		fmt.Println("Check your authentication token or selected profile and try again")
		return false

	}
	authStartSpinner.Success()
	return true
}

func readFileStep() (body bytes.Buffer, writer *multipart.Writer) {

	writer = multipart.NewWriter(&body)
	// Get the boundary string from the writer to use in the Content-Type header.

	// Walk through the directory and add all files to the multipart form.
	err := filepath.Walk(dirPath, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		// Skip the "maze-output" directory.
		if info.IsDir() && info.Name() == "maze-output" {
			return filepath.SkipDir
		}

		// Skip directories; only process files.
		if !info.IsDir() {
			// Open the file.
			if strings.HasPrefix(filepath.Ext(info.Name()), ".tf") || strings.HasPrefix(filepath.Ext(info.Name()), ".tfvars") || strings.HasPrefix(filepath.Ext(info.Name()), ".tfstate") {

				fileToUpload, err := os.Open(path)
				if err != nil {
					return fmt.Errorf("failed to open file %s: %v", path, err)
				}
				defer fileToUpload.Close()

				// Create a form file field for the file with the key "file".
				part, err := writer.CreateFormFile("file", info.Name())
				if err != nil {
					return fmt.Errorf("failed to create form file: %v", err)
				}

				// Copy the file content to the multipart form field.
				if _, err := io.Copy(part, fileToUpload); err != nil {
					return fmt.Errorf("failed to copy file content: %v", err)
				}

				// Compute the relative path of the file to the base directory.
				relativePath, err := filepath.Rel(dirPath, path)
				if err != nil {
					return fmt.Errorf("failed to compute relative path: %v", err)
				}

				// Create a form field for the file path using the relative path.
				if err := writer.WriteField("originalPath", relativePath); err != nil {
					return fmt.Errorf("failed to write path field: %v", err)
				}
			}
		}
		return nil
	})
	if err != nil {
		return
	}

	// Close the multipart writer to finalize the form data.
	if err := writer.Close(); err != nil {
		return
	}
	return
}

func sendFiles(body bytes.Buffer, writer *multipart.Writer) (success bool, bodyBytes []byte) {
	success = false

	boundary := writer.Boundary()

	// Create a new HTTP request with the multipart data.

	submitReq, err := http.NewRequest("POST", url+"/api/cli/submitFiles/", &body)

	if err != nil {
		return
	}

	submitReq.Header.Set("Content-Type", fmt.Sprintf("multipart/form-data; boundary=%s", boundary))

	// Set the Authorization header with the provided token.
	submitReq.Header.Set("Authorization", token)
	var uploadStartSpinner = ux.NewSpinner("Uploading files", "Files uploaded", "Upload failed", false)
	uploadStartSpinner.Start()
	// Execute the request.
	client := &http.Client{}
	resp, err := client.Do(submitReq)
	if err != nil {
		return
	}
	bodyBytes, err = io.ReadAll(resp.Body)
	if err != nil {
		fmt.Println(err)
		return
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusOK || resp.StatusCode == http.StatusCreated {
		success = true
	} else {
		uploadStartSpinner.Fail()
		fmt.Println(resp)
		return

	}
	uploadStartSpinner.Success()
	return
}

func complianceStep(path string) {

	//Compliance  ----------------------------------------------
	var complianceStartSpinner = ux.NewSpinner("Starting compliance testing", "Compliance testing complete", "Authentication failed", false)
	complianceStartSpinner.Start()
	submitReq, err := http.NewRequest("GET", url+"/api/cli/compliance/"+path, nil)

	// Create a new HTTP request with the multipart data.
	if err != nil {
		return
	}

	// Set the Content-Type header with the boundary.

	submitReq.Header.Set("Content-Type", "application/json")
	// Set the Authorization header with the provided token.
	submitReq.Header.Set("Authorization", token)

	// Execute the request.
	client := &http.Client{}
	resp, err := client.Do(submitReq)
	if err != nil {
		complianceStartSpinner.Fail()
		return
	}
	bodyBytes, err := io.ReadAll(resp.Body)

	if err != nil {
		complianceStartSpinner.Fail()
		fmt.Println(err)
		return
	}

	var data Response
	if err := json.Unmarshal(bodyBytes, &data); err != nil {
		fmt.Printf("Compliance testing failed: %v", err)
	}
	complianceStartSpinner.Success()
	time.Sleep(1000 * time.Millisecond)

	fmt.Printf("Summary:\n")
	fmt.Printf("    Passed: %d\n", data.Summary.Passed)
	fmt.Printf("    Failed: %d\n", data.Summary.Failed)

	filePath := filepath.Join(dirPath, "maze-output", "maze_compliance_results.json")
	if err := os.WriteFile(filePath, bodyBytes, 0644); err != nil {
		fmt.Printf("Failed to write response to file: %v", err)
	}
	time.Sleep(1000 * time.Millisecond)

	fmt.Printf("Full compliance test saved to %s\n", filePath)

	return
}

func validateStep(path string) (success bool) {

	var validateStartSpinner = ux.NewSpinner("Validation starting", "Validation complete", "Validation failed", false)
	validateStartSpinner.Start()
	body := []byte(`{"provider":"` + provider + `"}`)

	validateReq, err := http.NewRequest("GET", url+"/api/cli/tfvalidate/"+path, bytes.NewBuffer(body))
	if err != nil {
		validateStartSpinner.Fail()
		return false
	}

	//Auth ----------------------------------------------
	validateReq.Header.Set("Content-Type", "application/json")
	validateReq.Header.Set("Authorization", token)

	// Execute the request.
	client := &http.Client{}
	resp, err := client.Do(validateReq)
	bodyBytes, err := io.ReadAll(resp.Body)
	if err != nil {
		validateStartSpinner.Fail()

		return false

	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK && resp.StatusCode != http.StatusCreated {
		validateStartSpinner.Fail()

		fmt.Println("validate service unavailable")
		return false

	}
	validateStartSpinner.Success()
	time.Sleep(500 * time.Millisecond)

	fmt.Println(string(bodyBytes))
	return true

}

func formatStep(path string) (success bool) {

	var formatStartSpinner = ux.NewSpinner("Formatting starting", "Formatting complete", "Formatting failed", false)
	formatStartSpinner.Start()

	formatReq, err := http.NewRequest("GET", url+"/api/cli/tfformat/"+path, nil)
	if err != nil {
		formatStartSpinner.Fail()
		return false
	}

	//Auth ----------------------------------------------
	formatReq.Header.Set("Content-Type", "application/json")
	formatReq.Header.Set("Authorization", token)
	time.Sleep(1000 * time.Millisecond)

	// Execute the request.
	client := &http.Client{}
	resp, err := client.Do(formatReq)
	if err != nil {
		formatStartSpinner.Fail()

		return false

	}
	//bodyBytes, err := io.ReadAll(resp.Body)

	//defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK && resp.StatusCode != http.StatusCreated {
		formatStartSpinner.Fail()

		fmt.Println("Format service unavailable")
		return false

	}
	formatStartSpinner.Success()
	time.Sleep(500 * time.Millisecond)

	return true

}

func planStep(path string) (success bool, bodyBytes []byte) {

	success = false
	body := []byte(`{"name":"test","description":"test","provider":"` + provider + `"}`)
	var planStartSpinner = ux.NewSpinner("Plan starting", "Plan started", "Plan failed to start", false)
	planStartSpinner.Start()
	time.Sleep(1000 * time.Millisecond)
	planStartSpinner.Success()

	var planProgressSpinner = ux.NewSpinner("Plan in progress", "Plan finished", "Plan failed", true)
	planProgressSpinner.Start()

	submitReq, err := http.NewRequest("POST", url+"/api/cli/tfplan/"+path, bytes.NewBuffer(body))

	// Create a new HTTP request with the multipart data.
	if err != nil {
		return
	}

	// Set the Content-Type header with the boundary.

	submitReq.Header.Set("Content-Type", "application/json")
	// Set the Authorization header with the provided token.
	submitReq.Header.Set("Authorization", token)

	// Execute the request.
	client := &http.Client{}
	resp, err := client.Do(submitReq)
	if err != nil {
		planProgressSpinner.Fail()
		return
	}
	bodyBytes, err = io.ReadAll(resp.Body)
	if err != nil {
		planProgressSpinner.Fail()
		fmt.Println(err)
		return
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusOK || resp.StatusCode == http.StatusCreated {
		planProgressSpinner.Success()
		success = true
	} else {
		fmt.Println(resp)
		return

	}
	// var planProcessingSpinner = ux.NewSpinner("Plan processing", "Plan processed", "Plan failed", false)
	// planProcessingSpinner.Start()
	// time.Sleep(1000 * time.Millisecond)
	// planProcessingSpinner.Success()

	return
}

func deleteFilesStep(path string) (success bool) {
	authReq, err := http.NewRequest("GET", url+"/api/cli/deletefiles/"+path, nil)
	if err != nil {
		return false
	}

	//Auth ----------------------------------------------
	authReq.Header.Set("Authorization", token)

	// Execute the request.
	client := &http.Client{}
	resp, err := client.Do(authReq)

	if err != nil {
		return false

	}
	if resp.StatusCode != http.StatusOK {
		return false

	}
	return true
}

func costStep(projectId string) (success bool) {
	//Cost ----------------------------------------------

	var costStartSpinner = ux.NewSpinner("Calculating cloud cost", "Cost calculated", "Cost calculation failed", false)
	costStartSpinner.Start()

	authReq, err := http.NewRequest("GET", url+"/api/cost/"+projectId, nil)
	if err != nil {
		return false
	}

	//Auth ----------------------------------------------
	authReq.Header.Set("Authorization", token)

	// Execute the request.
	client := &http.Client{}
	resp, err := client.Do(authReq)
	bodyBytes, err := io.ReadAll(resp.Body)
	if err != nil {
		costStartSpinner.Fail()
		return false

	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		costStartSpinner.Fail()
		fmt.Println("cost service unavailable")
		return false

	}
	type Resources struct {
		HourlyCost  float64 `json:"hourlyCost"`
		Description string  `json:"description"`
		Currency    string  `json:"currency"`
		Resource    string  `json:"resource"`
	}
	type Cost struct {
		TotalCost float64     `json:"totalCost"`
		Resources []Resources `json:"resources"`
	}
	var cost Cost
	err = json.Unmarshal(bodyBytes, &cost)
	if err != nil {
		fmt.Println("Error:", err)
		return false
	}
	time.Sleep(2000 * time.Millisecond)

	costStartSpinner.Success()
	time.Sleep(1000 * time.Millisecond)

	fmt.Println(fmt.Sprint("    Daily cost: $", roundFloat(cost.TotalCost*24, 2)))
	fmt.Println(fmt.Sprint("    Monthly cost: $", roundFloat(cost.TotalCost*730, 2)))
	// Check the response status.

	time.Sleep(1000 * time.Millisecond)

	return true
}

func imageStep(projectId string, path string) {

	filePath := filepath.Join(path, "maze-output", "maze_canvas_image.png")

	var imageStartSpinner = ux.NewSpinner("Generating canvas image", "Image saved to: "+filePath, "Generating canvas image failed", false)
	imageStartSpinner.Start()

	authReq, err := http.NewRequest("GET", url+"/api/cli/canvasimage/"+projectId, nil)
	if err != nil {
		return
	}

	//Auth ----------------------------------------------
	authReq.Header.Set("Authorization", token)

	// Execute the request.
	client := &http.Client{}
	resp, err := client.Do(authReq)
	imageData, err := io.ReadAll(resp.Body)
	if err != nil {
		imageStartSpinner.Fail()
		return

	}

	imgBytes, err := base64.StdEncoding.DecodeString(string(imageData))
	if err != nil {
		fmt.Println(err)

		imageStartSpinner.Fail()
		return
	}

	err = os.WriteFile(filePath, imgBytes, 0644)
	if err != nil {
		fmt.Println(err)
		imageStartSpinner.Fail()

		return
	}
	imageStartSpinner.Success()
}

func roundFloat(val float64, precision uint) float64 {
	ratio := math.Pow(10, float64(precision))
	return math.Round(val*ratio) / ratio
}
func init() {

	// Here you will define your flags and configuration settings.
	planCmd.Flags().StringVarP(&dirPath, "dir", "d", "./", "The directory for terraform files")
	planCmd.Flags().StringVarP(&profileName, "profile", "p", "default", "The profile to get the auth token from")
	planCmd.Flags().StringVarP(&url, "url", "u", "https://maze-multicloud.com", "The url for the maze instance you are using")
	planCmd.Flags().StringVarP(&name, "name", "n", "default", "The name for the generated project")
	planCmd.Flags().StringVarP(&provider, "provider", "", "", "The name of the provider, e.g. AWS or AZURE")
	planCmd.Flags().BoolVarP(&generateImage, "image", "i", false, "Generate an image of the canvas and save to terraform files folder after plan is complete")

	planCmd.SetOutput(color.Output)
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
	planCmd.SetUsageTemplate(usageTemplate)

}

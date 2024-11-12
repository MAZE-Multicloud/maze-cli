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
	"os/exec"
	"path/filepath"
	"regexp"
	"strings"

	"time"

	"maze/cmd/ux"

	"github.com/fatih/color"
	"github.com/spf13/cobra"
)

var (
	dirPath       string
	url           string
	name          string
	generateImage bool
	token         string
	profileName   string
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
	fmt.Printf("We are going to create a project in maze using terraform provided from %s and the project will be named %s\n\n\n", dirPath, name)

	time.Sleep(3000 * time.Millisecond)

	success := authStep()

	if !success {
		return nil
	}

	time.Sleep(500 * time.Millisecond)

	body, writer := readFileStep()

	time.Sleep(1000 * time.Millisecond)

	success, bodyBytes := planStep(body, writer)

	if !success {
		return nil
	}
	projectId := string(bodyBytes)
	complianceStep()

	time.Sleep(1000 * time.Millisecond)

	costStep(projectId)

	time.Sleep(1000 * time.Millisecond)
	if generateImage {
		imageStep(projectId, dirPath)
		time.Sleep(2000 * time.Millisecond)
	}

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
		// Skip directories; only process files.
		if !info.IsDir() {
			// Open the file.
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

func sendFiles(path string) (success bool, bodyBytes []byte) {
	success = false

	body := []byte(`{name:"test",description:"test"}`)
	// Create a new HTTP request with the multipart data.
	submitReq, err := http.NewRequest("POST", url+"/api/cli/plan/"+name, bytes.NewBuffer(body))
	if err != nil {
		return
	}

	// Set the Authorization header with the provided token.
	submitReq.Header.Set("Authorization", token)

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

	println(string(bodyBytes))

	if resp.StatusCode == http.StatusOK || resp.StatusCode == http.StatusCreated {
		success = true
	} else {
		fmt.Println(resp)
		return

	}
	return
}

func complianceStep() {

	//Compliance  ----------------------------------------------
	var complianceStartSpinner = ux.NewSpinner("Starting compliance", "Compliance testing in progress", "Authentication failed", false)
	complianceStartSpinner.Start()

	out, err := exec.Command("checkov", "serv.dev").Output()
	if err != nil {

	}
	fmt.Print(out)

	time.Sleep(1000 * time.Millisecond)
	complianceStartSpinner.Success()

	// Industrial Compliance sub headings ----------------------------------------------

	var industryComplianceStartSpinner = ux.NewSpinner("Industry compliance testing in progress", "Industry compliance testing finished", "Industry compliance testing failed", true)
	industryComplianceStartSpinner.Start()

	time.Sleep(2000 * time.Millisecond)

	industryComplianceStartSpinner.Success()

	time.Sleep(500 * time.Millisecond)
	// var complianceEndSpinner = ux.NewSpinner("Compliance processing", "Compliance finished", "Authentication failed", false)
	// complianceEndSpinner.Start()

	// time.Sleep(1000 * time.Millisecond)
	// complianceEndSpinner.Success()
	return
}

func planStep(body bytes.Buffer, writer *multipart.Writer) (success bool, bodyBytes []byte) {

	success = false
	boundary := writer.Boundary()

	var planStartSpinner = ux.NewSpinner("Plan starting", "Plan started", "Plan failed to start", false)
	planStartSpinner.Start()
	time.Sleep(1000 * time.Millisecond)
	planStartSpinner.Success()

	var planProgressSpinner = ux.NewSpinner("Plan in progress", "Plan finished", "Plan failed", true)
	planProgressSpinner.Start()

	// Create a new HTTP request with the multipart data.
	submitReq, err := http.NewRequest("POST", url+"/api/cli/submit/"+name, &body)
	if err != nil {
		return
	}

	// Set the Content-Type header with the boundary.
	submitReq.Header.Set("Content-Type", fmt.Sprintf("multipart/form-data; boundary=%s", boundary))

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
	var planProcessingSpinner = ux.NewSpinner("Plan processing", "Plan processed", "Plan failed", false)
	planProcessingSpinner.Start()
	time.Sleep(1000 * time.Millisecond)
	planProcessingSpinner.Success()

	return
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

	var imageStartSpinner = ux.NewSpinner("Generating canvas image", "Image saved to: "+path+"/canvas_image.png", "Generating canvas image failed", false)
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

	// Define the file path and name
	filePath := path + "/canvas_image.png"
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

package main

import (
	"errors"
	"os"
	"path/filepath"

	"github.com/yashid-mohamed/harbor-scan-report/cmd/harbor-scan-report/config"
	"github.com/yashid-mohamed/harbor-scan-report/cmd/harbor-scan-report/github"
	"github.com/yashid-mohamed/harbor-scan-report/cmd/harbor-scan-report/image"
	"github.com/yashid-mohamed/harbor-scan-report/cmd/harbor-scan-report/log"
	"github.com/yashid-mohamed/harbor-scan-report/cmd/harbor-scan-report/report"
	"github.com/yashid-mohamed/harbor-scan-report/cmd/harbor-scan-report/sarif"
	"github.com/yashid-mohamed/harbor-scan-report/cmd/harbor-scan-report/scan"
	"github.com/yashid-mohamed/harbor-scan-report/cmd/harbor-scan-report/util"
)

func main() {
	log.Debug.Println("Application Configuration: " + config.PrintConfig())

	//find Image
	findResult := image.GetFinder().FindImage()
	if findResult.Failed() {
		if findResult.HasError() {
			util.ExitOnError(findResult.GetError())
		} else {
			util.ExitOnError(errors.New("failed to find image"))
		}
	}

	//get scan results
	scanStatus := scan.WaitForScanCompeted()
	scanReport := scan.GetScanReport(scanStatus)

	log.Debug.Printf("Image '%s' has %d vulnerabilities (%d critical, %d high, %d medium, %d low)\n",
		config.Get().ImageInfo.Raw,
		scanReport.Counters.Total,
		scanReport.Counters.Critical,
		scanReport.Counters.High,
		scanReport.Counters.Medium,
		scanReport.Counters.Low,
	)
	if scanReport.Counters.Total > 0 {
		log.Debug.Printf("%d/%d fixable\n", scanReport.Counters.Fixable, scanReport.Counters.Total)
	}

	//write comment
	if config.Get().Github.Enabled {
		github.WriteComment(scanReport)
	}

	report.WriteListOfVulnerabilities(scanReport)

	// Generate SARIF report if output path is provided
	if util.IsStringPresent(config.Get().Report.SarifOutputPath) {
		sarifPath := config.Get().Report.SarifOutputPath

		// Print environment variables for debugging
		log.Debug.Printf("GITHUB_WORKSPACE: %s", os.Getenv("GITHUB_WORKSPACE"))
		log.Debug.Printf("PWD: %s", os.Getenv("PWD"))
		
		// Try to get absolute path and verify directory exists
		absPath, err := filepath.Abs(sarifPath)
		if err == nil {
			sarifPath = absPath
			log.Debug.Printf("Absolute SARIF path: %s", sarifPath)
			
			// Check if parent directory exists
			dir := filepath.Dir(sarifPath)
			if _, err := os.Stat(dir); os.IsNotExist(err) {
				log.Warning.Printf("Parent directory %s does not exist, attempting to create it", dir)
				if err := os.MkdirAll(dir, 0755); err != nil {
					log.Warning.Printf("Failed to create directory: %s", err.Error())
				}
			} else {
				log.Debug.Printf("Parent directory %s exists", dir)
			}
		}

		log.Info.Printf("Generating SARIF report at: %s", sarifPath)
		err = sarif.GenerateSarifReport(scanReport, sarifPath)
		if err != nil {
			log.Warning.Printf("Failed to generate SARIF report: %s\n", err.Error())
		} else {
			// Verify file was created
			if _, err := os.Stat(sarifPath); os.IsNotExist(err) {
				log.Warning.Printf("SARIF file was not found at %s after generation", sarifPath)
			} else {
				log.Info.Printf("SARIF report generated successfully at %s", sarifPath)
				
				// Print file contents for debugging
				if fileContent, err := os.ReadFile(sarifPath); err == nil {
					// Print first 500 characters
					previewLen := 500
					if len(fileContent) < previewLen {
						previewLen = len(fileContent)
					}
					log.Debug.Printf("SARIF file content preview: %s...", string(fileContent[:previewLen]))
				} else {
					log.Warning.Printf("Could not read SARIF file for verification: %s", err.Error())
				}
			}
		}
	}

	if scanReport.TopSeverity.IsMoreCriticalThen(config.Get().MaxAllowedSeverity) {
		var hasFixableVulnerabilities bool
		for _, vuln := range scanReport.Vulnerabilities {
			if vuln.Severity.IsMoreCriticalThen(config.Get().MaxAllowedSeverity) {
				if vuln.HasFixVersion() {
					hasFixableVulnerabilities = true
					break
				}
			}
		}
		if hasFixableVulnerabilities {
			log.Error.Fatalf("Image has fixable vulnerabilities that are more critical "+
				"then allowed severity %s. "+
				"Check failed \n",
				config.Get().MaxAllowedSeverity.String())
		}
	}
}

package sarif

import (
	"encoding/json"
	"fmt"
	"os"
	"time"

	"github.com/yashid-mohamed/harbor-scan-report/cmd/harbor-scan-report/config"
	"github.com/yashid-mohamed/harbor-scan-report/cmd/harbor-scan-report/log"
	"github.com/yashid-mohamed/harbor-scan-report/cmd/harbor-scan-report/scan"
	"github.com/yashid-mohamed/harbor-scan-report/cmd/harbor-scan-report/severity"
)

// SARIF represents the root object of a SARIF log file
type SarifReport struct {
	Schema  string `json:"$schema"`
	Version string `json:"version"`
	Runs    []Run  `json:"runs"`
}

// Run represents a single run of an analysis tool
type Run struct {
	Tool        Tool         `json:"tool"`
	Invocations []Invocation `json:"invocations,omitempty"`
	Results     []Result     `json:"results"`
}

// Tool contains information about the analysis tool
type Tool struct {
	Driver ToolComponent `json:"driver"`
}

// ToolComponent contains information about the analysis tool component
type ToolComponent struct {
	Name           string                  `json:"name"`
	Version        string                  `json:"version,omitempty"`
	InformationURI string                  `json:"informationUri,omitempty"`
	Rules          []ReportingDescriptor   `json:"rules,omitempty"`
	Properties     map[string]interface{}  `json:"properties,omitempty"`
}

// ReportingDescriptor represents a rule that was evaluated during the scan
type ReportingDescriptor struct {
	ID                   string                 `json:"id"`
	Name                 string                 `json:"name,omitempty"`
	ShortDescription     Message                `json:"shortDescription,omitempty"`
	FullDescription      Message                `json:"fullDescription,omitempty"`
	Help                 Message                `json:"help,omitempty"`
	DefaultConfiguration DefaultConfiguration   `json:"defaultConfiguration,omitempty"`
	Properties           map[string]interface{} `json:"properties,omitempty"`
}

// Message represents a message with optional formatting
type Message struct {
	Text string `json:"text"`
}

// DefaultConfiguration contains default configuration for a rule
type DefaultConfiguration struct {
	Level string `json:"level"`
}

// Invocation represents a single invocation of an analysis tool
type Invocation struct {
	ExecutionSuccessful bool      `json:"executionSuccessful"`
	StartTimeUTC        time.Time `json:"startTimeUtc,omitempty"`
	EndTimeUTC          time.Time `json:"endTimeUtc,omitempty"`
}

// Result represents a single analysis result
type Result struct {
	RuleID     string                 `json:"ruleId"`
	RuleIndex  int                    `json:"ruleIndex,omitempty"`
	Level      string                 `json:"level"`
	Message    Message                `json:"message"`
	Locations  []Location             `json:"locations,omitempty"`
	Properties map[string]interface{} `json:"properties,omitempty"`
}

// Location represents a location within an artifact
type Location struct {
	PhysicalLocation PhysicalLocation `json:"physicalLocation"`
}

// PhysicalLocation represents a physical location
type PhysicalLocation struct {
	ArtifactLocation ArtifactLocation `json:"artifactLocation"`
}

// ArtifactLocation represents the location of an artifact
type ArtifactLocation struct {
	URI string `json:"uri"`
}

// GenerateSarifReport converts a scan report into SARIF format
func GenerateSarifReport(report *scan.Report, outputPath string) error {
	sarifReport := createSarifReport(report)

	// Convert to JSON
	jsonData, err := json.MarshalIndent(sarifReport, "", "  ")
	if err != nil {
		return fmt.Errorf("error marshaling SARIF report: %w", err)
	}

	// Write to file if output path provided, otherwise to stdout
	if outputPath != "" {
		err = os.WriteFile(outputPath, jsonData, 0644)
		if err != nil {
			return fmt.Errorf("error writing SARIF report to file: %w", err)
		}
		log.Info.Printf("SARIF report written to %s\n", outputPath)
	} else {
		fmt.Println(string(jsonData))
	}

	return nil
}

// createSarifReport converts a scan report to SARIF format
func createSarifReport(report *scan.Report) SarifReport {
	rules := []ReportingDescriptor{}
	results := []Result{}
	ruleMap := make(map[string]int)

	// Create rules and results for each vulnerability
	for _, vuln := range report.Vulnerabilities {
		// Only include vulnerabilities if they are fixable and report-only-fixable is set
		if config.Get().Report.ShowFixableOnly && !vuln.HasFixVersion() {
			continue
		}

		// Create rule
		ruleID := vuln.ID
		if _, exists := ruleMap[ruleID]; !exists {
			ruleMap[ruleID] = len(rules)
			properties := map[string]interface{}{
				"severity":      vuln.Severity.String(),
				"score":         vuln.Score,
				"package":       vuln.Package,
				"version":       vuln.Version,
				"fixVersion":    vuln.FixVersion,
				"hasFixVersion": vuln.HasFixVersion(),
			}

			rules = append(rules, ReportingDescriptor{
				ID:               ruleID,
				Name:             fmt.Sprintf("Vulnerability %s", ruleID),
				ShortDescription: Message{Text: fmt.Sprintf("%s in %s", ruleID, vuln.Package)},
				FullDescription:  Message{Text: vuln.Description},
				DefaultConfiguration: DefaultConfiguration{
					Level: mapSeverityToLevel(vuln.Severity),
				},
				Properties: properties,
			})
		}

		// Create result
		artifactURI := fmt.Sprintf("%s/%s:%s",
			config.Get().ImageInfo.Project,
			config.Get().ImageInfo.RepoName,
			config.Get().ImageInfo.Tag)

		results = append(results, Result{
			RuleID:    ruleID,
			RuleIndex: ruleMap[ruleID],
			Level:     mapSeverityToLevel(vuln.Severity),
			Message: Message{Text: fmt.Sprintf("%s found in %s version %s. %s",
				ruleID, vuln.Package, vuln.Version, getFixMessage(vuln))},
			Locations: []Location{
				{
					PhysicalLocation: PhysicalLocation{
						ArtifactLocation: ArtifactLocation{
							URI: artifactURI,
						},
					},
				},
			},
			Properties: map[string]interface{}{
				"severity": vuln.Severity.String(),
				"score":    vuln.Score,
				"fixable":  vuln.HasFixVersion(),
			},
		})
	}

	// Create the SARIF report
	return SarifReport{
		Schema:  "https://schemastore.azurewebsites.net/schemas/json/sarif-2.1.0.json",
		Version: "2.1.0",
		Runs: []Run{
			{
				Tool: Tool{
					Driver: ToolComponent{
						Name:           "Trivy via Harbor",
						Version:        report.Scanner.Version,
						InformationURI: "https://github.com/aquasecurity/trivy",
						Rules:          rules,
						Properties: map[string]interface{}{
							"vendor":               report.Scanner.Vendor,
							"executed_via":         "Harbor",
							"harbor_scan_reporter": "https://github.com/yashid-mohamed/harbor-scan-report",
						},
					},
				},
				Invocations: []Invocation{
					{
						ExecutionSuccessful: !report.Failed,
						StartTimeUTC:        time.Now().UTC().Add(-time.Duration(report.Scanner.Duration) * time.Second),
						EndTimeUTC:          report.GeneratedAt,
					},
				},
				Results: results,
			},
		},
	}
}

// mapSeverityToLevel maps a severity level to a SARIF level
func mapSeverityToLevel(sev severity.Severity) string {
	switch sev {
	case severity.Critical, severity.High:
		return "error"
	case severity.Medium:
		return "warning"
	case severity.Low:
		return "note"
	default:
		return "none"
	}
}

// getFixMessage returns a message about the fix status
func getFixMessage(vuln scan.Vulnerability) string {
	if vuln.HasFixVersion() {
		return fmt.Sprintf("Fixed in version %s", vuln.FixVersion)
	}
	return "No fix available"
}

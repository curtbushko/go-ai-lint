package reporters

import (
	"encoding/json"
	"io"

	"github.com/curtbushko/go-ai-lint/internal/domain"
)

// SARIF 2.1.0 schema types for IDE integration.
// See: https://docs.oasis-open.org/sarif/sarif/v2.1.0/sarif-v2.1.0.html

// SARIFLog is the root SARIF object.
type SARIFLog struct {
	Version string     `json:"version"`
	Schema  string     `json:"$schema"`
	Runs    []SARIFRun `json:"runs"`
}

// SARIFRun represents a single run of an analysis tool.
type SARIFRun struct {
	Tool    SARIFTool     `json:"tool"`
	Results []SARIFResult `json:"results"`
}

// SARIFTool describes the analysis tool.
type SARIFTool struct {
	Driver SARIFToolComponent `json:"driver"`
}

// SARIFToolComponent describes a tool component.
type SARIFToolComponent struct {
	Name           string            `json:"name"`
	Version        string            `json:"version,omitempty"`
	InformationURI string            `json:"informationUri,omitempty"`
	Rules          []SARIFRule       `json:"rules,omitempty"`
	Properties     map[string]string `json:"properties,omitempty"`
}

// SARIFRule describes a rule (analyzer).
type SARIFRule struct {
	ID               string            `json:"id"`
	Name             string            `json:"name,omitempty"`
	ShortDescription SARIFMessage      `json:"shortDescription,omitempty"`
	FullDescription  SARIFMessage      `json:"fullDescription,omitempty"`
	Help             SARIFMessage      `json:"help,omitempty"`
	DefaultConfig    SARIFRuleConfig   `json:"defaultConfiguration,omitempty"`
	Properties       map[string]string `json:"properties,omitempty"`
}

// SARIFRuleConfig describes the default configuration for a rule.
type SARIFRuleConfig struct {
	Level string `json:"level,omitempty"`
}

// SARIFMessage is a message with text content.
type SARIFMessage struct {
	Text string `json:"text,omitempty"`
}

// SARIFResult represents a single issue found.
type SARIFResult struct {
	RuleID    string          `json:"ruleId"`
	Level     string          `json:"level"`
	Message   SARIFMessage    `json:"message"`
	Locations []SARIFLocation `json:"locations,omitempty"`
}

// SARIFLocation describes where an issue was found.
type SARIFLocation struct {
	PhysicalLocation SARIFPhysicalLocation `json:"physicalLocation"`
}

// SARIFPhysicalLocation describes the physical file location.
type SARIFPhysicalLocation struct {
	ArtifactLocation SARIFArtifactLocation `json:"artifactLocation"`
	Region           SARIFRegion           `json:"region,omitempty"`
}

// SARIFArtifactLocation identifies the artifact (file).
type SARIFArtifactLocation struct {
	URI string `json:"uri"`
}

// SARIFRegion describes a region within an artifact.
type SARIFRegion struct {
	StartLine   int `json:"startLine,omitempty"`
	StartColumn int `json:"startColumn,omitempty"`
	EndLine     int `json:"endLine,omitempty"`
	EndColumn   int `json:"endColumn,omitempty"`
}

// SARIFReporter outputs issues in SARIF 2.1.0 format for IDE integration.
type SARIFReporter struct {
	w io.Writer
}

// NewSARIFReporter creates a new SARIF reporter.
func NewSARIFReporter(w io.Writer) *SARIFReporter {
	return &SARIFReporter{w: w}
}

// Report writes issues in SARIF format.
func (r *SARIFReporter) Report(issues []domain.Issue) error {
	log := SARIFLog{
		Version: "2.1.0",
		Schema:  "https://json.schemastore.org/sarif-2.1.0.json",
		Runs: []SARIFRun{
			{
				Tool: SARIFTool{
					Driver: SARIFToolComponent{
						Name:           "go-ai-lint",
						Version:        "0.1.0",
						InformationURI: "https://github.com/curtbushko/go-ai-lint",
						Rules:          r.buildRules(issues),
					},
				},
				Results: r.buildResults(issues),
			},
		},
	}

	encoder := json.NewEncoder(r.w)
	encoder.SetIndent("", "  ")
	return encoder.Encode(log)
}

// buildRules creates unique rules from the issues.
func (r *SARIFReporter) buildRules(issues []domain.Issue) []SARIFRule {
	seen := make(map[string]bool)
	var rules []SARIFRule

	for _, issue := range issues {
		if seen[issue.ID] {
			continue
		}
		seen[issue.ID] = true

		rule := SARIFRule{
			ID:   issue.ID,
			Name: issue.Name,
			ShortDescription: SARIFMessage{
				Text: issue.Message,
			},
			DefaultConfig: SARIFRuleConfig{
				Level: severityToLevel(issue.Severity),
			},
			Properties: map[string]string{
				"category": string(issue.Category),
			},
		}

		// Add full description from Why field
		if issue.Why != "" {
			rule.FullDescription = SARIFMessage{
				Text: issue.Why,
			}
		}

		// Add help text combining Why and Fix
		helpText := r.buildHelpText(issue)
		if helpText != "" {
			rule.Help = SARIFMessage{
				Text: helpText,
			}
		}

		rules = append(rules, rule)
	}

	return rules
}

// buildHelpText creates help text from Why and Fix fields.
func (r *SARIFReporter) buildHelpText(issue domain.Issue) string {
	var text string
	if issue.Why != "" {
		text = "Why: " + issue.Why
	}
	if issue.Fix != "" {
		if text != "" {
			text += "\n\n"
		}
		text += "Fix: " + issue.Fix
	}
	return text
}

// buildResults creates SARIF results from issues.
func (r *SARIFReporter) buildResults(issues []domain.Issue) []SARIFResult {
	results := make([]SARIFResult, len(issues))

	for i, issue := range issues {
		results[i] = SARIFResult{
			RuleID: issue.ID,
			Level:  severityToLevel(issue.Severity),
			Message: SARIFMessage{
				Text: issue.Message,
			},
			Locations: []SARIFLocation{
				{
					PhysicalLocation: SARIFPhysicalLocation{
						ArtifactLocation: SARIFArtifactLocation{
							URI: issue.Position.Filename,
						},
						Region: SARIFRegion{
							StartLine:   issue.Position.Line,
							StartColumn: issue.Position.Column,
							EndLine:     issue.Position.EndLine,
							EndColumn:   issue.Position.EndColumn,
						},
					},
				},
			},
		}
	}

	return results
}

// severityToLevel maps domain severity to SARIF level.
func severityToLevel(s domain.Severity) string {
	switch s {
	case domain.SeverityCritical, domain.SeverityHigh:
		return "error"
	case domain.SeverityMedium:
		return "warning"
	case domain.SeverityLow:
		return "note"
	default:
		return "warning"
	}
}

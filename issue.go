// Copyright 2026 Robert Boucher
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package main

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

// issueResponse represents the relevant fields from GET /rest/api/3/issue/{key}.
type issueResponse struct {
	Key    string `json:"key"`
	Self   string `json:"self"`
	Fields struct {
		Summary     string          `json:"summary"`
		Description json.RawMessage `json:"description"`
		Status      struct {
			Name string `json:"name"`
		} `json:"status"`
		Labels      []string     `json:"labels"`
		Attachments []Attachment `json:"attachment"`
		IssueType   struct {
			Name string `json:"name"`
		} `json:"issuetype"`
		Priority struct {
			Name string `json:"name"`
		} `json:"priority"`
		Assignee *struct {
			DisplayName string `json:"displayName"`
		} `json:"assignee"`
		Reporter *struct {
			DisplayName string `json:"displayName"`
		} `json:"reporter"`
	} `json:"fields"`
}

// GetIssue fetches a Jira issue, converts its description to markdown,
// downloads all attachments, and writes the output folder.
// Returns the path to the output folder.
func GetIssue(client *JiraClient, issueKey string, basePath string) (string, error) {
	// Fetch issue
	resp, err := client.Get("/issue/" + issueKey + "?fields=summary,description,status,labels,attachment,issuetype,priority,assignee,reporter")
	if err != nil {
		return "", fmt.Errorf("failed to fetch issue %s: %w", issueKey, err)
	}

	var issue issueResponse
	if err := ReadJSONBody(resp, &issue); err != nil {
		return "", fmt.Errorf("failed to read issue response: %w", err)
	}

	// Create output directory
	outputDir := filepath.Join(basePath, issue.Key)
	if err := os.MkdirAll(outputDir, 0755); err != nil {
		return "", fmt.Errorf("failed to create output directory: %w", err)
	}

	// Download all attachments
	mediaIDToFilename := make(map[string]string)
	attachmentPaths := make(map[string]string)

	if len(issue.Fields.Attachments) > 0 {
		for _, att := range issue.Fields.Attachments {
			mediaIDToFilename[att.ID] = att.Filename
		}
		attachmentPaths, err = DownloadAttachments(client, issue.Fields.Attachments, outputDir)
		if err != nil {
			return "", fmt.Errorf("failed to download attachments: %w", err)
		}
	}

	// Convert description ADF to Markdown
	descriptionMD := ""
	if issue.Fields.Description != nil && string(issue.Fields.Description) != "null" {
		descriptionMD, err = ADFToMarkdown(issue.Fields.Description, attachmentPaths, mediaIDToFilename)
		if err != nil {
			return "", fmt.Errorf("failed to convert description: %w", err)
		}
	}

	// Build ticket.md
	var md strings.Builder
	md.WriteString("# ")
	md.WriteString(issue.Key)
	md.WriteString(": ")
	md.WriteString(issue.Fields.Summary)
	md.WriteString("\n\n")

	// Metadata
	md.WriteString("**Status:** ")
	md.WriteString(issue.Fields.Status.Name)
	md.WriteString("\n")

	if issue.Fields.IssueType.Name != "" {
		md.WriteString("**Type:** ")
		md.WriteString(issue.Fields.IssueType.Name)
		md.WriteString("\n")
	}

	if issue.Fields.Priority.Name != "" {
		md.WriteString("**Priority:** ")
		md.WriteString(issue.Fields.Priority.Name)
		md.WriteString("\n")
	}

	if issue.Fields.Assignee != nil {
		md.WriteString("**Assignee:** ")
		md.WriteString(issue.Fields.Assignee.DisplayName)
		md.WriteString("\n")
	}

	if issue.Fields.Reporter != nil {
		md.WriteString("**Reporter:** ")
		md.WriteString(issue.Fields.Reporter.DisplayName)
		md.WriteString("\n")
	}

	if len(issue.Fields.Labels) > 0 {
		md.WriteString("**Labels:** ")
		md.WriteString(strings.Join(issue.Fields.Labels, ", "))
		md.WriteString("\n")
	}

	if descriptionMD != "" {
		md.WriteString("\n## Description\n\n")
		md.WriteString(descriptionMD)
		md.WriteString("\n")
	}

	// List attachments
	if len(issue.Fields.Attachments) > 0 {
		md.WriteString("\n## Attachments\n\n")
		for _, att := range issue.Fields.Attachments {
			md.WriteString("- [")
			md.WriteString(att.Filename)
			md.WriteString("](")
			md.WriteString(att.Filename)
			md.WriteString(")\n")
		}
	}

	// Write ticket.md
	ticketPath := filepath.Join(outputDir, "ticket.md")
	if err := os.WriteFile(ticketPath, []byte(md.String()), 0644); err != nil {
		return "", fmt.Errorf("failed to write ticket.md: %w", err)
	}

	return outputDir, nil
}

// CreateIssue creates a new Jira issue and returns the issue key and URL.
func CreateIssue(client *JiraClient, project string, summary string, description string) (string, string, error) {
	body := map[string]interface{}{
		"fields": map[string]interface{}{
			"project": map[string]string{
				"key": project,
			},
			"summary": summary,
			"issuetype": map[string]string{
				"name": "Task",
			},
			"description": json.RawMessage(TextToADF(description)),
		},
	}

	resp, err := client.Post("/issue", body)
	if err != nil {
		return "", "", fmt.Errorf("failed to create issue: %w", err)
	}

	var result struct {
		Key  string `json:"key"`
		Self string `json:"self"`
	}
	if err := ReadJSONBody(resp, &result); err != nil {
		return "", "", fmt.Errorf("failed to read create response: %w", err)
	}

	issueURL := strings.TrimRight(client.config.BaseURL, "/") + "/browse/" + result.Key
	return result.Key, issueURL, nil
}

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
	"flag"
	"fmt"
	"os"
)

// Version is set at build time via -ldflags "-X main.Version=...".
var Version = "dev"

func main() {
	// Flags
	listFlag := flag.Bool("l", false, "List issues matching a JQL query (requires -q)")
	jqlFlag := flag.String("q", "", "JQL query string (used with -l)")
	ticketFlag := flag.String("t", "", "Get ticket details by issue key (e.g. PROJ-123)")
	pathFlag := flag.String("p", "", "Base path for ticket output folder (default: ~/.go-jira-tool/issues/)")
	createFlag := flag.String("c", "", "Create a new ticket with the given summary")
	descFlag := flag.String("d", "", "Ticket description text (used with -c)")
	descPathFlag := flag.String("dp", "", "Path to a file containing the ticket description (used with -c)")
	transitionFlag := flag.String("transition", "", "Transition an issue to a new status (e.g. \"In Progress\")")
	commentFlag := flag.String("comment", "", "Add a comment to an issue")
	commentFileFlag := flag.String("cf", "", "Path to a file containing the comment body (used with -comment)")
	assignFlag := flag.String("assign", "", "Assign an issue to the current user")
	versionFlag := flag.Bool("version", false, "Print version and exit")

	flag.Usage = func() {
		fmt.Fprintf(os.Stderr, "Usage: go-jira-tool [flags]\n\n")
		fmt.Fprintf(os.Stderr, "A CLI tool for interacting with Jira.\n\n")
		fmt.Fprintf(os.Stderr, "Modes:\n")
		fmt.Fprintf(os.Stderr, "  List:       go-jira-tool -l -q \"<JQL query>\"\n")
		fmt.Fprintf(os.Stderr, "  Get:        go-jira-tool -t <ISSUE-KEY> [-p <output-path>]\n")
		fmt.Fprintf(os.Stderr, "  Create:     go-jira-tool -c \"<summary>\" [-d \"<description>\" | -dp <file>]\n")
		fmt.Fprintf(os.Stderr, "  Transition: go-jira-tool -transition <ISSUE-KEY> \"<status>\"\n")
		fmt.Fprintf(os.Stderr, "  Comment:    go-jira-tool -comment <ISSUE-KEY> \"<body>\" | -cf <file>\n")
		fmt.Fprintf(os.Stderr, "  Assign:     go-jira-tool -assign <ISSUE-KEY>\n")
		fmt.Fprintf(os.Stderr, "  Version:    go-jira-tool -version\n")
		fmt.Fprintf(os.Stderr, "\nFlags:\n")
		flag.PrintDefaults()
	}

	flag.Parse()

	if *versionFlag {
		fmt.Println(Version)
		return
	}

	// Determine which mode we're in
	mode := ""
	if *listFlag {
		mode = "list"
	}
	if *ticketFlag != "" {
		if mode != "" {
			fatal("cannot combine multiple mode flags (-l, -t, -c, -transition, -comment, -assign)")
		}
		mode = "get"
	}
	if *createFlag != "" {
		if mode != "" {
			fatal("cannot combine multiple mode flags (-l, -t, -c, -transition, -comment, -assign)")
		}
		mode = "create"
	}
	if *transitionFlag != "" {
		if mode != "" {
			fatal("cannot combine multiple mode flags (-l, -t, -c, -transition, -comment, -assign)")
		}
		mode = "transition"
	}
	if *commentFlag != "" {
		if mode != "" {
			fatal("cannot combine multiple mode flags (-l, -t, -c, -transition, -comment, -assign)")
		}
		mode = "comment"
	}
	if *assignFlag != "" {
		if mode != "" {
			fatal("cannot combine multiple mode flags (-l, -t, -c, -transition, -comment, -assign)")
		}
		mode = "assign"
	}

	if mode == "" {
		flag.Usage()
		os.Exit(1)
	}

	// Load config and create client
	cfg, err := LoadConfig()
	if err != nil {
		fatal("config error: %v", err)
	}
	client := NewJiraClient(cfg)

	switch mode {
	case "list":
		runList(client, cfg, *jqlFlag)
	case "get":
		runGet(client, *ticketFlag, *pathFlag)
	case "create":
		runCreate(client, cfg, *createFlag, *descFlag, *descPathFlag)
	case "transition":
		runTransition(client, *transitionFlag)
	case "comment":
		runComment(client, *commentFlag, *commentFileFlag)
	case "assign":
		runAssign(client, *assignFlag)
	}
}

func runList(client *JiraClient, cfg *Config, jql string) {
	if jql == "" {
		// If no JQL provided but we have a default project, list all issues in it
		if cfg.DefaultProject != "" {
			jql = "project = " + cfg.DefaultProject + " ORDER BY updated DESC"
		} else {
			fatal("-q (JQL query) is required when using -l")
		}
	}

	results, err := SearchIssues(client, jql)
	if err != nil {
		fatal("search failed: %v", err)
	}

	if err := PrintSearchResults(results); err != nil {
		fatal("output error: %v", err)
	}
}

func runGet(client *JiraClient, issueKey string, basePath string) {
	if basePath == "" {
		dir, err := IssuesDir()
		if err != nil {
			fatal("failed to resolve default issues directory: %v", err)
		}
		basePath = dir
	}
	outputDir, err := GetIssue(client, issueKey, basePath)
	if err != nil {
		fatal("failed to get issue: %v", err)
	}
	fmt.Println(outputDir)
}

func runCreate(client *JiraClient, cfg *Config, summary string, desc string, descPath string) {
	// Resolve description
	if descPath != "" {
		data, err := os.ReadFile(descPath)
		if err != nil {
			fatal("failed to read description file: %v", err)
		}
		desc = string(data)
	}
	if desc == "" {
		desc = summary
	}

	// Determine project
	project := cfg.DefaultProject
	if project == "" {
		fatal("default_project must be set in config.json to create tickets")
	}

	key, url, err := CreateIssue(client, project, summary, desc)
	if err != nil {
		fatal("failed to create issue: %v", err)
	}
	fmt.Printf("%s %s\n", key, url)
}

func runTransition(client *JiraClient, issueKey string) {
	args := flag.Args()
	if len(args) < 1 {
		fatal("-transition requires a target status as a positional argument\n  Usage: go-jira-tool -transition <ISSUE-KEY> \"<status>\"")
	}
	targetStatus := args[0]

	if err := TransitionIssue(client, issueKey, targetStatus); err != nil {
		fatal("transition failed: %v", err)
	}
	fmt.Printf("%s transitioned to %s\n", issueKey, targetStatus)
}

func runComment(client *JiraClient, issueKey string, commentFilePath string) {
	var body string
	if commentFilePath != "" {
		// File-based comment
		data, err := os.ReadFile(commentFilePath)
		if err != nil {
			fatal("failed to read comment file: %v", err)
		}
		body = string(data)
	} else {
		// Inline comment from positional arg
		args := flag.Args()
		if len(args) < 1 {
			fatal("-comment requires a comment body as a positional argument or -cf <file>\n  Usage: go-jira-tool -comment <ISSUE-KEY> \"<body>\"")
		}
		body = args[0]
	}

	if err := AddComment(client, issueKey, body); err != nil {
		fatal("comment failed: %v", err)
	}
	fmt.Printf("Comment added to %s\n", issueKey)
}

func runAssign(client *JiraClient, issueKey string) {
	if err := AssignIssue(client, issueKey); err != nil {
		fatal("assign failed: %v", err)
	}
	fmt.Printf("%s assigned to current user\n", issueKey)
}

func fatal(format string, args ...interface{}) {
	fmt.Fprintf(os.Stderr, "error: "+format+"\n", args...)
	os.Exit(1)
}

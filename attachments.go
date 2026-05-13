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
	"fmt"
	"io"
	"os"
	"path/filepath"
)

type Attachment struct {
	ID       string `json:"id"`
	Filename string `json:"filename"`
	Content  string `json:"content"` // download URL
	MimeType string `json:"mimeType"`
}

// DownloadAttachments downloads all attachments to outputDir and returns a map
// of filename to local file path.
func DownloadAttachments(client *JiraClient, attachments []Attachment, outputDir string) (map[string]string, error) {
	if err := os.MkdirAll(outputDir, 0755); err != nil {
		return nil, fmt.Errorf("failed to create output directory: %w", err)
	}

	localPaths := make(map[string]string)

	for _, att := range attachments {
		localPath := filepath.Join(outputDir, att.Filename)

		if err := downloadFile(client, att.Content, localPath); err != nil {
			return nil, fmt.Errorf("failed to download attachment %s: %w", att.Filename, err)
		}

		localPaths[att.Filename] = localPath
	}

	return localPaths, nil
}

func downloadFile(client *JiraClient, url, destPath string) error {
	resp, err := client.GetRaw(url)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return fmt.Errorf("download failed (HTTP %d) for %s", resp.StatusCode, url)
	}

	out, err := os.Create(destPath)
	if err != nil {
		return fmt.Errorf("failed to create file %s: %w", destPath, err)
	}
	defer out.Close()

	if _, err := io.Copy(out, resp.Body); err != nil {
		return fmt.Errorf("failed to write file %s: %w", destPath, err)
	}

	return nil
}

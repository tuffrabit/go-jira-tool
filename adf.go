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
	"strings"
)

// ADFNode represents a node in the Atlassian Document Format tree.
type ADFNode struct {
	Type    string          `json:"type"`
	Version int             `json:"version,omitempty"`
	Content []ADFNode       `json:"content,omitempty"`
	Text    string          `json:"text,omitempty"`
	Marks   []ADFMark       `json:"marks,omitempty"`
	Attrs   json.RawMessage `json:"attrs,omitempty"`
}

type ADFMark struct {
	Type  string          `json:"type"`
	Attrs json.RawMessage `json:"attrs,omitempty"`
}

// adfContext carries state needed during ADF-to-Markdown conversion.
type adfContext struct {
	// attachmentPaths maps attachment filenames to local relative paths.
	attachmentPaths map[string]string
	// mediaIDToFilename maps Jira media IDs to attachment filenames.
	mediaIDToFilename map[string]string
	// listDepth tracks nesting depth for indentation.
	listDepth int
}

// ADFToMarkdown converts an ADF document to a Markdown string.
// attachmentPaths maps filenames to local paths for media reference replacement.
// mediaIDToFilename maps Jira attachment IDs to filenames.
func ADFToMarkdown(adfJSON json.RawMessage, attachmentPaths map[string]string, mediaIDToFilename map[string]string) (string, error) {
	var doc ADFNode
	if err := json.Unmarshal(adfJSON, &doc); err != nil {
		return "", fmt.Errorf("failed to parse ADF document: %w", err)
	}

	ctx := &adfContext{
		attachmentPaths:   attachmentPaths,
		mediaIDToFilename: mediaIDToFilename,
	}

	var sb strings.Builder
	ctx.renderChildren(&sb, doc.Content, "\n\n")
	return strings.TrimSpace(sb.String()), nil
}

func (ctx *adfContext) renderChildren(sb *strings.Builder, children []ADFNode, separator string) {
	for i, child := range children {
		if i > 0 && separator != "" {
			sb.WriteString(separator)
		}
		ctx.renderNode(sb, child)
	}
}

func (ctx *adfContext) renderNode(sb *strings.Builder, node ADFNode) {
	switch node.Type {
	case "doc":
		ctx.renderChildren(sb, node.Content, "\n\n")

	case "paragraph":
		ctx.renderInlineChildren(sb, node.Content)

	case "heading":
		level := ctx.getAttrInt(node.Attrs, "level", 1)
		sb.WriteString(strings.Repeat("#", level))
		sb.WriteString(" ")
		ctx.renderInlineChildren(sb, node.Content)

	case "bulletList":
		ctx.renderList(sb, node.Content, false)

	case "orderedList":
		ctx.renderList(sb, node.Content, true)

	case "listItem":
		// Rendered by renderList
		ctx.renderListItemContent(sb, node.Content)

	case "codeBlock":
		lang := ctx.getAttrString(node.Attrs, "language", "")
		sb.WriteString("```")
		sb.WriteString(lang)
		sb.WriteString("\n")
		ctx.renderPlainTextChildren(sb, node.Content)
		sb.WriteString("\n```")

	case "blockquote":
		var inner strings.Builder
		ctx.renderChildren(&inner, node.Content, "\n\n")
		for _, line := range strings.Split(inner.String(), "\n") {
			sb.WriteString("> ")
			sb.WriteString(line)
			sb.WriteString("\n")
		}

	case "rule":
		sb.WriteString("---")

	case "table":
		ctx.renderTable(sb, node.Content)

	case "mediaSingle", "mediaGroup":
		for _, child := range node.Content {
			ctx.renderNode(sb, child)
		}

	case "media":
		ctx.renderMedia(sb, node)

	case "inlineCard":
		url := ctx.getAttrString(node.Attrs, "url", "")
		if url != "" {
			sb.WriteString("[")
			sb.WriteString(url)
			sb.WriteString("](")
			sb.WriteString(url)
			sb.WriteString(")")
		}

	case "hardBreak":
		sb.WriteString("\n")

	default:
		// Unknown node type — render any text content we can find
		ctx.renderInlineChildren(sb, node.Content)
	}
}

func (ctx *adfContext) renderInlineChildren(sb *strings.Builder, children []ADFNode) {
	for _, child := range children {
		switch child.Type {
		case "text":
			ctx.renderText(sb, child)
		case "hardBreak":
			sb.WriteString("\n")
		case "inlineCard":
			ctx.renderNode(sb, child)
		case "mention":
			text := ctx.getAttrString(child.Attrs, "text", "")
			if text != "" {
				sb.WriteString(text)
			}
		case "emoji":
			shortName := ctx.getAttrString(child.Attrs, "shortName", "")
			if shortName != "" {
				sb.WriteString(shortName)
			}
		case "media":
			ctx.renderMedia(sb, child)
		case "mediaSingle", "mediaGroup":
			for _, c := range child.Content {
				ctx.renderNode(sb, c)
			}
		default:
			// Try to extract text from unknown inline nodes
			if child.Text != "" {
				sb.WriteString(child.Text)
			} else {
				ctx.renderInlineChildren(sb, child.Content)
			}
		}
	}
}

func (ctx *adfContext) renderText(sb *strings.Builder, node ADFNode) {
	text := node.Text
	// Apply marks (bold, italic, code, link, etc.)
	for _, mark := range node.Marks {
		switch mark.Type {
		case "strong":
			text = "**" + text + "**"
		case "em":
			text = "_" + text + "_"
		case "code":
			text = "`" + text + "`"
		case "strike":
			text = "~~" + text + "~~"
		case "link":
			href := getAttrStringFromRaw(mark.Attrs, "href", "")
			if href != "" {
				text = "[" + text + "](" + href + ")"
			}
		}
	}
	sb.WriteString(text)
}

func (ctx *adfContext) renderList(sb *strings.Builder, items []ADFNode, ordered bool) {
	ctx.listDepth++
	indent := strings.Repeat("  ", ctx.listDepth-1)
	for i, item := range items {
		if i > 0 {
			sb.WriteString("\n")
		}
		sb.WriteString(indent)
		if ordered {
			sb.WriteString(fmt.Sprintf("%d. ", i+1))
		} else {
			sb.WriteString("- ")
		}
		ctx.renderListItemContent(sb, item.Content)
	}
	ctx.listDepth--
}

func (ctx *adfContext) renderListItemContent(sb *strings.Builder, content []ADFNode) {
	for i, child := range content {
		if child.Type == "paragraph" {
			if i > 0 {
				sb.WriteString("\n")
				sb.WriteString(strings.Repeat("  ", ctx.listDepth))
			}
			ctx.renderInlineChildren(sb, child.Content)
		} else if child.Type == "bulletList" || child.Type == "orderedList" {
			sb.WriteString("\n")
			ctx.renderNode(sb, child)
		} else {
			if i > 0 {
				sb.WriteString("\n")
			}
			ctx.renderNode(sb, child)
		}
	}
}

func (ctx *adfContext) renderTable(sb *strings.Builder, rows []ADFNode) {
	if len(rows) == 0 {
		return
	}

	// Collect all cell texts
	var table [][]string
	for _, row := range rows {
		var cells []string
		for _, cell := range row.Content {
			var cellSB strings.Builder
			ctx.renderChildren(&cellSB, cell.Content, " ")
			cells = append(cells, strings.TrimSpace(cellSB.String()))
		}
		table = append(table, cells)
	}

	if len(table) == 0 {
		return
	}

	// Determine column widths
	colCount := 0
	for _, row := range table {
		if len(row) > colCount {
			colCount = len(row)
		}
	}

	// Render header row
	sb.WriteString("| ")
	for i := 0; i < colCount; i++ {
		if i < len(table[0]) {
			sb.WriteString(table[0][i])
		}
		sb.WriteString(" | ")
	}
	sb.WriteString("\n")

	// Separator
	sb.WriteString("| ")
	for i := 0; i < colCount; i++ {
		sb.WriteString("---")
		sb.WriteString(" | ")
	}

	// Data rows
	for _, row := range table[1:] {
		sb.WriteString("\n| ")
		for i := 0; i < colCount; i++ {
			if i < len(row) {
				sb.WriteString(row[i])
			}
			sb.WriteString(" | ")
		}
	}
}

func (ctx *adfContext) renderMedia(sb *strings.Builder, node ADFNode) {
	mediaID := ctx.getAttrString(node.Attrs, "id", "")
	filename := ""

	// Try to resolve media ID to a filename
	if mediaID != "" && ctx.mediaIDToFilename != nil {
		filename = ctx.mediaIDToFilename[mediaID]
	}
	// Fallback: check alt text or other attrs
	if filename == "" {
		filename = ctx.getAttrString(node.Attrs, "alt", "")
	}

	if filename != "" {
		sb.WriteString("[")
		sb.WriteString(filename)
		sb.WriteString("](")
		sb.WriteString(filename)
		sb.WriteString(")")
	} else if mediaID != "" {
		sb.WriteString("[attachment:")
		sb.WriteString(mediaID)
		sb.WriteString("]")
	}
}

func (ctx *adfContext) renderPlainTextChildren(sb *strings.Builder, children []ADFNode) {
	for _, child := range children {
		if child.Text != "" {
			sb.WriteString(child.Text)
		}
	}
}

// Attribute helpers

func (ctx *adfContext) getAttrInt(attrs json.RawMessage, key string, defaultVal int) int {
	if attrs == nil {
		return defaultVal
	}
	var m map[string]json.RawMessage
	if err := json.Unmarshal(attrs, &m); err != nil {
		return defaultVal
	}
	raw, ok := m[key]
	if !ok {
		return defaultVal
	}
	var val int
	if err := json.Unmarshal(raw, &val); err != nil {
		return defaultVal
	}
	return val
}

func (ctx *adfContext) getAttrString(attrs json.RawMessage, key string, defaultVal string) string {
	return getAttrStringFromRaw(attrs, key, defaultVal)
}

func getAttrStringFromRaw(attrs json.RawMessage, key string, defaultVal string) string {
	if attrs == nil {
		return defaultVal
	}
	var m map[string]json.RawMessage
	if err := json.Unmarshal(attrs, &m); err != nil {
		return defaultVal
	}
	raw, ok := m[key]
	if !ok {
		return defaultVal
	}
	var val string
	if err := json.Unmarshal(raw, &val); err != nil {
		return defaultVal
	}
	return val
}

// TextToADF wraps a plain text string in a minimal ADF document structure.
func TextToADF(text string) json.RawMessage {
	doc := map[string]interface{}{
		"type":    "doc",
		"version": 1,
		"content": splitTextToADFParagraphs(text),
	}
	data, _ := json.Marshal(doc)
	return data
}

// splitTextToADFParagraphs splits text by double newlines into ADF paragraph nodes.
func splitTextToADFParagraphs(text string) []interface{} {
	paragraphs := strings.Split(text, "\n\n")
	var content []interface{}
	for _, p := range paragraphs {
		p = strings.TrimSpace(p)
		if p == "" {
			continue
		}
		// Split single newlines into hardBreak-separated text nodes
		lines := strings.Split(p, "\n")
		var inlineContent []interface{}
		for i, line := range lines {
			if i > 0 {
				inlineContent = append(inlineContent, map[string]interface{}{
					"type": "hardBreak",
				})
			}
			inlineContent = append(inlineContent, map[string]interface{}{
				"type": "text",
				"text": line,
			})
		}
		content = append(content, map[string]interface{}{
			"type":    "paragraph",
			"content": inlineContent,
		})
	}
	if len(content) == 0 {
		content = append(content, map[string]interface{}{
			"type":    "paragraph",
			"content": []interface{}{map[string]interface{}{"type": "text", "text": text}},
		})
	}
	return content
}

package main

import (
	"bufio"
	"bytes"
	"crypto/sha1"
	"encoding/json"
	"fmt"
	"math/rand"
	"os"
	"strings"

	"github.com/alecthomas/chroma"
	"github.com/alecthomas/chroma/formatters"
	"github.com/alecthomas/chroma/lexers"
	"github.com/alecthomas/chroma/styles"
	"github.com/spf13/cobra"
	"github.com/yuin/goldmark"
	"github.com/yuin/goldmark/ast"
	"github.com/yuin/goldmark/text"
)

// ANSI Constants
const (
	Reset      = "\033[0m"
	Green      = "\033[32m"
	HideCursor = "\033[?25l"
	ShowCursor = "\033[?25h"
	// Default terminal prompt
	DefaultPrompt = Green + "rosera@labdemo.app:~$ " + Reset
)

// Global map for special block handling.
// If the value is empty, it's treated as a raw response (output).
// If the value is "SKIP", the block is ignored.
// If the value is a string, it's treated as a comment-style request.
var specialPrefixes = map[string]string{
	"info":       "# Info: ",
	"warn":       "# Warn: ",
	"ignore":     "SKIP",
	"output":     "",
	"yaml":       "",
	"dockerfile": "",
	"json":       "",
//	"bash":       "",
//	"sh":         "",
	"sql":        "",
	"mermaid":    "SKIP",
}

type AsciinemaHeader struct {
	Version   int               `json:"version"`
	Width     int               `json:"width"`
	Height    int               `json:"height"`
	Timestamp int64             `json:"timestamp,omitempty"`
	Env       map[string]string `json:"env"`
}

type RequestResponse struct {
	ID       string `json:"id"`
	Request  string `json:"request"`
	Response string `json:"response"`
}

var (
	inputFile  string
	outputFile string
	prompt     string
	width      int
	height     int
	theme      string
)

func main() {
	var rootCmd = &cobra.Command{
		Use:   "mtv",
		Short: "A tool to convert Markdown/JSON to asciinema recordings",
	}

	var genCmd = &cobra.Command{
		Use:   "generate",
		Short: "Generate a .cast file from a JSON input file",
		Run:   runGenerate,
	}
	genCmd.Flags().StringVarP(&inputFile, "input", "i", "", "Input JSON file (required)")
	genCmd.Flags().StringVarP(&outputFile, "output", "o", "output.cast", "Output .cast file")
	genCmd.Flags().StringVarP(&prompt, "prompt", "p", DefaultPrompt, "Terminal prompt string")
	genCmd.Flags().IntVar(&width, "width", 100, "Terminal width")
	genCmd.Flags().IntVar(&height, "height", 30, "Terminal height")
	genCmd.MarkFlagRequired("input")

	var convCmd = &cobra.Command{
		Use:   "convert",
		Short: "Convert a Markdown file into a JSON request/response schema",
		Run:   runConvert,
	}
	convCmd.Flags().StringVarP(&inputFile, "input", "i", "", "Input Markdown file (required)")
	convCmd.Flags().StringVarP(&outputFile, "output", "o", "commands.json", "Output JSON file")
	convCmd.Flags().StringVarP(&theme, "theme", "t", "monokai", "Chroma highlighting theme")
	convCmd.MarkFlagRequired("input")

	rootCmd.AddCommand(genCmd, convCmd)

	if err := rootCmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}

func generateID(content string) string {
	h := sha1.New()
	h.Write([]byte(content))
	return fmt.Sprintf("%x", h.Sum(nil))[:8]
}

func highlightText(content, language, styleName string) string {
	lexer := lexers.Get(language)
	if lexer == nil {
		lexer = lexers.Fallback
	}
	lexer = chroma.Coalesce(lexer)

	style := styles.Get(styleName)
	if style == nil {
		style = styles.Fallback
	}

	formatter := formatters.TTY256
	var buf bytes.Buffer
	iterator, err := lexer.Tokenise(nil, content)
	if err != nil {
		return content
	}

	err = formatter.Format(&buf, style, iterator)
	if err != nil {
		return content
	}

	return buf.String()
}

func runConvert(cmd *cobra.Command, args []string) {
	source, err := os.ReadFile(inputFile)
	if err != nil {
		fmt.Printf("Error reading input file: %v\n", err)
		return
	}

	entries := convertMarkdownToEntries(source)
	jsonData, _ := json.MarshalIndent(entries, "", "  ")
	os.WriteFile(outputFile, jsonData, 0644)
	fmt.Printf("Successfully converted %s to %s (%d entries)\n", inputFile, outputFile, len(entries))
}

func runGenerate(cmd *cobra.Command, args []string) {
	source, err := os.ReadFile(inputFile)
	if err != nil {
		fmt.Printf("Error reading input file: %v\n", err)
		return
	}

	var entries []RequestResponse
	json.Unmarshal(source, &entries)

	out, _ := os.Create(outputFile)
	defer out.Close()

	writer := bufio.NewWriter(out)
	defer writer.Flush()

	header := AsciinemaHeader{
		Version: 2, Width: width, Height: height,
		Env: map[string]string{"TERM": "xterm-256color", "SHELL": "/bin/bash"},
	}
	headerJSON, _ := json.Marshal(header)
	writer.WriteString(string(headerJSON) + "\n")

	var currentTime float64 = 0.5
	processEntries(entries, writer, &currentTime)
	writeEvent(writer, currentTime+2.0, "")

	fmt.Printf("Successfully converted %s to %s\n", inputFile, outputFile)
}

func convertMarkdownToEntries(source []byte) []RequestResponse {
	var entries []RequestResponse
	md := goldmark.New()
	reader := text.NewReader(source)
	doc := md.Parser().Parse(reader)

	ast.Walk(doc, func(n ast.Node, entering bool) (ast.WalkStatus, error) {
		if entering && n.Kind() == ast.KindFencedCodeBlock {
			block := n.(*ast.FencedCodeBlock)
			info := string(block.Info.Value(source))
			infoParts := strings.Fields(info)

			var language, postfix string
			if len(infoParts) > 0 {
				language = strings.ToLower(infoParts[0])
			}
			if len(infoParts) > 1 {
				postfix = strings.ToLower(infoParts[1])
			} else if len(infoParts) == 1 {
				if _, ok := specialPrefixes[language]; ok {
					postfix = language
				}
			}

			prefix, isSpecial := specialPrefixes[postfix]
			if isSpecial && prefix == "SKIP" {
				return ast.WalkContinue, nil
			}

			// FIXED: Correctly extracting block content from segments
			var fullBlock strings.Builder
			lines := block.Lines()
			for i := 0; i < lines.Len(); i++ {
				segment := lines.At(i)
				fullBlock.Write(segment.Value(source))
			}
			content := strings.TrimSpace(fullBlock.String())

			if isSpecial && prefix == "" {
				highlighted := highlightText(content, language, theme)
				if len(entries) > 0 {
					if entries[len(entries)-1].Response != "" {
						entries[len(entries)-1].Response += "\r\n" + highlighted
					} else {
						entries[len(entries)-1].Response = highlighted
					}
				}
			} else if isSpecial && prefix != "" {
				contentLines := strings.Split(content, "\n")
				var commentedContent strings.Builder
				for idx, cl := range contentLines {
					commentedContent.WriteString(prefix + cl)
					if idx < len(contentLines)-1 {
						commentedContent.WriteString("\n")
					}
				}
				req := commentedContent.String()
				entries = append(entries, RequestResponse{ID: generateID(req), Request: req})

        } else {
         rawLines := strings.Split(content, "\n")
         for _, rl := range rawLines {
           trimmed := strings.TrimSpace(rl)
           if trimmed != "" {
             // Apply Chroma highlighting to the request command
             highlighted := highlightText(trimmed, language, theme)
             entries = append(entries, RequestResponse{ID: generateID(trimmed), Request: highlighted})
           }
         }
      }

		}
		return ast.WalkContinue, nil
	})
	return entries
}

func processEntries(entries []RequestResponse, writer *bufio.Writer, currentTime *float64) {
	writeEvent(writer, *currentTime, "\r\n")
	for _, entry := range entries {
		writeEvent(writer, *currentTime, ShowCursor+prompt)
		*currentTime += 0.5
		reqLines := strings.Split(entry.Request, "\n")
		for lineIdx, lineText := range reqLines {
			for _, char := range lineText {
				*currentTime += 0.04 + rand.Float64()*(0.12-0.04)
				writeEvent(writer, *currentTime, string(char))
			}
			*currentTime += 0.1
			writeEvent(writer, *currentTime, "\r\n")
			if lineIdx < len(reqLines)-1 {
				writeEvent(writer, *currentTime, prompt)
				*currentTime += 0.2
			}
		}
		if entry.Response != "" {
			*currentTime += 0.1
			writeEvent(writer, *currentTime, HideCursor)
			resp := strings.ReplaceAll(entry.Response, "\n", "\r\n")
			writeEvent(writer, *currentTime, resp+"\r\n\r\n")
			*currentTime += 0.5
			writeEvent(writer, *currentTime, ShowCursor)
		}
		*currentTime += 1.2
	}
}

func writeEvent(w *bufio.Writer, ts float64, data string) {
	event := []interface{}{ts, "o", data}
	eventJSON, _ := json.Marshal(event)
	w.WriteString(string(eventJSON) + "\n")
}

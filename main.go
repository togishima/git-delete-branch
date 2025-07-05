package main

import (
	"bytes"
	"embed"
	"encoding/json"
	"flag"
	"fmt"
	"os"
	"os/exec"
	"regexp"
	"strings"

	"github.com/AlecAivazis/survey/v2"
	"github.com/nicksnyder/go-i18n/v2/i18n"
	"golang.org/x/text/language"
)

//go:embed locales/*.json
var localeFS embed.FS

// ANSI escape code for colors
const (
	ColorGreen = "\033[32m"
	ColorRed   = "\033[31m"
	ColorReset = "\033[0m"
)

// Regex to remove ANSI color codes
var ansiStripper = regexp.MustCompile("\033[[0-9;]*m")

type BranchDetail struct {
	Name    string
	Hash    string
	Author  string
	Date    string
	Message string
}

// cleanBranchName removes color codes and merge indicators from a branch name
func cleanBranchName(branchName string) string {
	// First, remove ANSI color codes
	cleaned := ansiStripper.ReplaceAllString(branchName, "")
	// Then, remove the merge indicator (e.g., " (merged)" or " (unmerged)")
	parts := strings.SplitN(cleaned, " (", 2)
	return strings.TrimSpace(parts[0])
}

func getBranchDetail(branchName string) (BranchDetail, error) {
	cleanName := cleanBranchName(branchName)
	cmd := exec.Command("git", "log", "-1", "--pretty=format:%H%n%an%n%ad%n%s", cleanName)
	output, err := cmd.CombinedOutput()
	if err != nil {
		return BranchDetail{}, fmt.Errorf("git log failed: %w\n%s", err, string(output))
	}

	lines := strings.Split(strings.TrimSpace(string(output)), "\n")
	if len(lines) < 4 {
		return BranchDetail{}, fmt.Errorf("unexpected git log output: %s", string(output))
	}

	return BranchDetail{
		Name:    cleanName,
		Hash:    lines[0],
		Author:  lines[1],
		Date:    lines[2],
		Message: lines[3],
	}, nil
}

func main() {
	bundle := i18n.NewBundle(language.English)
	bundle.RegisterUnmarshalFunc("json", json.Unmarshal)
	bundle.LoadMessageFileFS(localeFS, "locales/en.json")
	bundle.LoadMessageFileFS(localeFS, "locales/ja.json")

	// Language option
	langFlag := flag.String("lang", "", "Specify the language (e.g., en, ja)")
	helpFlag := flag.Bool("h", false, "Show help")
	flag.BoolVar(helpFlag, "help", false, "Show help")

	// Internal flag for fzf preview
	getLogFlag := flag.String("get-log", "", "Internal flag to get log for a branch")

	flag.Parse()

	var lang string
	if *langFlag != "" {
		lang = *langFlag
	} else {
		lang = os.Getenv("LANG")
	}

	if !strings.HasPrefix(lang, "ja") {
		lang = "en"
	}

	localizer := i18n.NewLocalizer(bundle, lang)

	// Handle internal fzf preview request
	if *getLogFlag != "" {
		cleanName := cleanBranchName(*getLogFlag)
		cmd := exec.Command("git", "log", "--color=always", cleanName)
		cmd.Stdout = os.Stdout
		cmd.Stderr = os.Stderr
		err := cmd.Run()
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error getting log for %s: %v\n", cleanName, err)
			os.Exit(1)
		}
		os.Exit(0)
	}

	if *helpFlag {
		usage, _ := localizer.Localize(&i18n.LocalizeConfig{MessageID: "HelpUsage"})
		description, _ := localizer.Localize(&i18n.LocalizeConfig{MessageID: "HelpDescription"})
		help, _ := localizer.Localize(&i18n.LocalizeConfig{MessageID: "HelpFlag"})
		langHelp, _ := localizer.Localize(&i18n.LocalizeConfig{MessageID: "HelpLangFlag"})

		fmt.Printf("%s\n\n%s\n\nOptions:\n  -h, --help    %s\n  -lang string  %s\n", usage, description, help, langHelp)
		os.Exit(0)
	}

	// Check if fzf is installed
	if _, err := exec.LookPath("fzf"); err != nil {
		fmt.Println(localizer.MustLocalize(&i18n.LocalizeConfig{MessageID: "FzfNotFound"}))
		fmt.Println(localizer.MustLocalize(&i18n.LocalizeConfig{MessageID: "InstallFzf"}))
		os.Exit(1)
	}

	// Get current branch
	currentBranchCmd := exec.Command("git", "rev-parse", "--abbrev-ref", "HEAD")
	currentBranchOutput, err := currentBranchCmd.CombinedOutput()
	if err != nil {
		msg, _ := localizer.Localize(&i18n.LocalizeConfig{
			MessageID: "ErrorGettingCurrentBranch",
			TemplateData: map[string]interface{}{"Error": err},
		})
		fmt.Println(msg)
		os.Exit(1)
	}
	currentBranch := strings.TrimSpace(string(currentBranchOutput))

	// Get all local branches
	cmd := exec.Command("git", "branch")
	output, err := cmd.CombinedOutput()
	if err != nil {
		msg, _ := localizer.Localize(&i18n.LocalizeConfig{
			MessageID: "ErrorRunningGitBranch",
			TemplateData: map[string]interface{}{"Error": err},
		})
		fmt.Println(msg)
		os.Exit(1)
	}

	allBranches := strings.Split(string(output), "\n")

	// Get merged branches
	mergedCmd := exec.Command("git", "branch", "--merged")
	mergedOutput, err := mergedCmd.CombinedOutput()
	if err != nil {
		// Log error but continue, as this is not critical
		fmt.Fprintf(os.Stderr, "Warning: Could not get merged branches: %v\n", err)
	}
	mergedBranchesMap := make(map[string]bool)
	for _, branch := range strings.Split(string(mergedOutput), "\n") {
		mergedBranchesMap[strings.TrimSpace(strings.TrimPrefix(branch, "* "))] = true
	}

	var fzfItems []string
	for _, branch := range allBranches {
		branch = strings.TrimSpace(strings.TrimPrefix(branch, "* "))
		if branch != "" && branch != currentBranch {
			indicator := localizer.MustLocalize(&i18n.LocalizeConfig{MessageID: "UnmergedIndicator"})
			color := ColorRed
			if mergedBranchesMap[branch] {
				indicator = localizer.MustLocalize(&i18n.LocalizeConfig{MessageID: "MergedIndicator"})
				color = ColorGreen
			}
			fzfItems = append(fzfItems, fmt.Sprintf("%s%s %s%s", color, branch, indicator, ColorReset))
		}
	}

	if len(fzfItems) == 0 {
		msg, _ := localizer.Localize(&i18n.LocalizeConfig{MessageID: "NoBranchesToDelete"})
		fmt.Println(msg)
		os.Exit(0)
	}

	// Prepare fzf command
	// Use os.Args[0] to get the path to the current executable
	executablePath, err := os.Executable()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error getting executable path: %v\n", err)
		os.Exit(1)
	}

	fzfCmd := exec.Command("fzf", "--multi", "--ansi", "--preview", fmt.Sprintf("%s -get-log {}", executablePath))
	fzfCmd.Stderr = os.Stderr // Show fzf errors

	// Pass branches to fzf stdin
	fzfStdin, err := fzfCmd.StdinPipe()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error creating stdin pipe for fzf: %v\n", err)
		os.Exit(1)
	}
	go func() {
		defer fzfStdin.Close()
		for _, item := range fzfItems {
			fmt.Fprintln(fzfStdin, item)
		}
	}()

	// Capture fzf stdout
	var fzfStdout bytes.Buffer
	fzfCmd.Stdout = &fzfStdout

	// Run fzf
	err = fzfCmd.Run()
	if err != nil {
		// fzf returns non-zero exit code if no selection or cancelled
		if exitError, ok := err.(*exec.ExitError); ok && exitError.ExitCode() == 130 {
			// User cancelled (Ctrl+C or Esc)
			fmt.Println(localizer.MustLocalize(&i18n.LocalizeConfig{MessageID: "DeletionCancelled"}))
			os.Exit(0)
		}
		fmt.Fprintf(os.Stderr, "Error running fzf: %v\n", err)
		os.Exit(1)
	}

	selectedBranchesStr := strings.TrimSpace(fzfStdout.String())
	if selectedBranchesStr == "" {
		msg, _ := localizer.Localize(&i18n.LocalizeConfig{MessageID: "NoBranchesSelected"})
		fmt.Println(msg)
		os.Exit(0)
	}

	// Clean selected branch names by removing indicators and color codes
	var branchesToDelete []string
	for _, selectedItem := range strings.Split(selectedBranchesStr, "\n") {
		branchesToDelete = append(branchesToDelete, cleanBranchName(selectedItem))
	}

	// Get details for selected branches
	var details []BranchDetail
	for _, branchName := range branchesToDelete {
		detail, err := getBranchDetail(branchName)
		if err != nil {
			msg, _ := localizer.Localize(&i18n.LocalizeConfig{
				MessageID: "ErrorGettingBranchDetails",
				TemplateData: map[string]interface{}{"Branch": branchName, "Error": err},
			})
			fmt.Println(msg)
			continue
		}
		details = append(details, detail)
	}

	if len(details) == 0 {
		msg, _ := localizer.Localize(&i18n.LocalizeConfig{MessageID: "NoBranchesSelected"})
		fmt.Println(msg)
		os.Exit(0)
	}

	// Display confirmation
	confirmMsg, _ := localizer.Localize(&i18n.LocalizeConfig{MessageID: "ConfirmDeletion"})
	fmt.Printf("\n%s\n", confirmMsg)

	branchHeader, _ := localizer.Localize(&i18n.LocalizeConfig{MessageID: "Branch"})
	hashHeader, _ := localizer.Localize(&i18n.LocalizeConfig{MessageID: "Hash"})
	authorHeader, _ := localizer.Localize(&i18n.LocalizeConfig{MessageID: "Author"})
	dateHeader, _ := localizer.Localize(&i18n.LocalizeConfig{MessageID: "Date"})
	messageHeader, _ := localizer.Localize(&i18n.LocalizeConfig{MessageID: "Message"})

	fmt.Printf("%-20s %-8s %-20s %-25s %s\n", branchHeader, hashHeader, authorHeader, dateHeader, messageHeader)
	fmt.Println(strings.Repeat("-", 90))

	for _, d := range details {
		fmt.Printf("%-20s %-8.8s %-20s %-25s %s\n", d.Name, d.Hash, d.Author, d.Date, d.Message)
	}
	fmt.Println(strings.Repeat("-", 90))

	// Use survey.Confirm for final confirmation
	confirmPrompt := &survey.Confirm{
		Message: "Proceed with deletion?",
		Default: false,
	}
	var confirm bool
	survey.AskOne(confirmPrompt, &confirm)

	if !confirm {
		cancelMsg, _ := localizer.Localize(&i18n.LocalizeConfig{MessageID: "DeletionCancelled"})
		fmt.Println(cancelMsg)
		os.Exit(0)
	}

	// Proceed with deletion
	for _, branch := range branchesToDelete {
		deleteCmd := exec.Command("git", "branch", "-d", branch)
		deleteOutput, err := deleteCmd.CombinedOutput()
		if err != nil {
			msg, _ := localizer.Localize(&i18n.LocalizeConfig{
				MessageID: "ErrorDeletingBranch",
				TemplateData: map[string]interface{}{"Branch": branch, "Error": err},
			})
			fmt.Println(msg)
			fmt.Println(string(deleteOutput))
		} else {
			msg, _ := localizer.Localize(&i18n.LocalizeConfig{
				MessageID: "BranchDeletedSuccessfully",
				TemplateData: map[string]interface{}{"Branch": branch},
			})
			fmt.Println(msg)
			fmt.Println(string(deleteOutput))
		}
	}
}

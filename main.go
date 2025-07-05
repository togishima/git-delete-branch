

package main

import (
	"embed"
	"encoding/json"
	"flag"
	"fmt"
	"os"
	"os/exec"
	"strings"

	"github.com/AlecAivazis/survey/v2"
	"github.com/nicksnyder/go-i18n/v2/i18n"
	"golang.org/x/text/language"
)

//go:embed locales/*.json
var localeFS embed.FS

type BranchDetail struct {
	Name    string
	Hash    string
	Author  string
	Date    string
	Message string
}

func getBranchDetail(branchName string) (BranchDetail, error) {
	cmd := exec.Command("git", "log", "-1", "--pretty=format:%H%n%an%n%ad%n%s", branchName)
	output, err := cmd.CombinedOutput()
	if err != nil {
		return BranchDetail{}, fmt.Errorf("git log failed: %w\n%s", err, string(output))
	}

	lines := strings.Split(strings.TrimSpace(string(output)), "\n")
	if len(lines) < 4 {
		return BranchDetail{}, fmt.Errorf("unexpected git log output: %s", string(output))
	}

	return BranchDetail{
		Name:    branchName,
		Hash:    lines[0],
		Author:  lines[1],
		Date:    lines[2],
		Message: lines[3],
	},	nil
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

	if *helpFlag {
		usage, _ := localizer.Localize(&i18n.LocalizeConfig{MessageID: "HelpUsage"})
		description, _ := localizer.Localize(&i18n.LocalizeConfig{MessageID: "HelpDescription"})
		help, _ := localizer.Localize(&i18n.LocalizeConfig{MessageID: "HelpFlag"})
		langHelp, _ := localizer.Localize(&i18n.LocalizeConfig{MessageID: "HelpLangFlag"})

		fmt.Printf("%s\n\n%s\n\nOptions:\n  -h, --help    %s\n  -lang string  %s\n", usage, description, help, langHelp)
		os.Exit(0)
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

	branches := strings.Split(string(output), "\n")
	var cleanBranches []string
	for _, branch := range branches {
		branch = strings.TrimSpace(strings.TrimPrefix(branch, "* "))
		if branch != "" && branch != currentBranch {
			cleanBranches = append(cleanBranches, branch)
		}
	}

	if len(cleanBranches) == 0 {
		msg, _ := localizer.Localize(&i18n.LocalizeConfig{MessageID: "NoBranchesToDelete"})
		fmt.Println(msg)
		os.Exit(0)
	}

	var branchesToDelete []string
	promptMsg, _ := localizer.Localize(&i18n.LocalizeConfig{MessageID: "SelectBranchesToDelete"})
	prompt := &survey.MultiSelect{
		Message: promptMsg,
		Options: cleanBranches,
	}
	survey.AskOne(prompt, &branchesToDelete)

	if len(branchesToDelete) == 0 {
		msg, _ := localizer.Localize(&i18n.LocalizeConfig{MessageID: "NoBranchesSelected"})
		fmt.Println(msg)
		os.Exit(0)
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

	confirm := false
	promptConfirm := &survey.Confirm{
		Message: "Proceed with deletion?",
		Default: false,
	}
	survey.AskOne(promptConfirm, &confirm)

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

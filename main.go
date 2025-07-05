
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

func main() {
	bundle := i18n.NewBundle(language.English)
	bundle.RegisterUnmarshalFunc("json", json.Unmarshal)
	bundle.LoadMessageFileFS(localeFS, "locales/en.json")
	bundle.LoadMessageFileFS(localeFS, "locales/ja.json")

	lang := os.Getenv("LANG")
	if !strings.HasPrefix(lang, "ja") {
		lang = "en"
	}
	localizer := i18n.NewLocalizer(bundle, lang)

	helpFlag := flag.Bool("h", false, "Show help")
	flag.BoolVar(helpFlag, "help", false, "Show help")
	flag.Parse()

	if *helpFlag {
		usage, _ := localizer.Localize(&i18n.LocalizeConfig{MessageID: "HelpUsage"})
		description, _ := localizer.Localize(&i18n.LocalizeConfig{MessageID: "HelpDescription"})
		help, _ := localizer.Localize(&i18n.LocalizeConfig{MessageID: "HelpFlag"})

		fmt.Printf("%s\n\n%s\n\nOptions:\n  -h, --help    %s\n", usage, description, help)
		os.Exit(0)
	}

	// ... (rest of the code is the same)

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

	for _, branch := range branches {
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

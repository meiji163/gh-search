package main

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"os/exec"
	"strconv"
	"strings"

	"github.com/AlecAivazis/survey/v2"
	"github.com/cli/safeexec"
	"github.com/fatih/color"
	"github.com/spf13/cobra"
)

var blue = color.New(color.FgBlue)
var yellow = color.New(color.FgYellow)

type SearchOptions struct {
	Query       string
	Interactive bool
	SortBy      string
	SearchIn    string
	Limit       int
	Topic       string
}

func rootCmd() *cobra.Command {
	opts := &SearchOptions{}
	cmd := &cobra.Command{
		Use:   "search <repository>",
		Short: "search repositories",
		Long: `Search for GitHub repositories.

Search through names, descriptions, or readme's, 
sort by repository stats, and filter by topic. 
`,
		Example:       `gh search cli --topic=hacktober`,
		Args:          cobra.ExactArgs(1),
		SilenceErrors: true,
		RunE: func(cmd *cobra.Command, args []string) error {
			opts.Query = args[0]

			searchIn := strings.ToLower(opts.SearchIn)
			if searchIn != "name" && searchIn != "description" && searchIn != "readme" {
				return errors.New(`--in argument must be "name", "description", or "readme"`)
			}

			if cmd.Flags().Changed("sort") || cmd.Flags().Changed("s") {
				sortBy := strings.ToLower(opts.SortBy)
				if sortBy != "stars" && sortBy != "forks" && sortBy != "issues" {
					return errors.New(`--sort argument must be "bestmatch", "stars", "forks", or "issues"`)
				}
			}

			if opts.Limit <= 0 {
				return errors.New("invalid limit")
			}

			return runSearch(opts)
		},
	}

	cmd.Flags().StringVarP(&opts.Topic, "topic", "t", "", `Specify a topic`)
	cmd.Flags().StringVarP(&opts.SearchIn, "in", "i", "name", `Search in "name", "description", or "readme"`)
	cmd.Flags().StringVarP(&opts.SortBy, "sort", "s", "", `Sort by "stars", "forks", or "issues"`)
	cmd.Flags().IntVarP(&opts.Limit, "limit", "L", 30, "Max number of search results")
	return cmd
}

type Repository struct {
	Name        string `json:"full_name"`
	Description string
	Stars       int    `json:"stargazers_count"`
	URL         string `json:"html_url"`
	Language    string
}

func runSearch(opts *SearchOptions) error {
	results, err := searchRepos(opts)
	if len(results) == 0 {
		fmt.Printf(`No results found for "%s"%s`, opts.Query, "\n")
	}

	var repos []string
	for i, repo := range results {
		repos = append(repos, prettyPrint(i+1, &repo))
	}

	numResults := len(repos)

	selector := &survey.Select{
		Message:  fmt.Sprintf("%d Results\n", numResults),
		Options:  repos,
		PageSize: 6,
	}

	var selection string
	err = survey.AskOne(selector, &selection,
		survey.WithIcons(func(icons *survey.IconSet) {
			icons.Question.Text = "::"
			icons.Question.Format = "yellow+hb"
		}),
	)
	if err != nil {
		return nil
	}

	n, err := strconv.Atoi(strings.Split(selection, " ")[0])
	if err != nil {
		return err
	}
	selectedRepo := results[n-1]
	fmt.Printf(`%[1]sFor more info, run "gh repo view %s" or view on the web at %s%[1]s`,
		"\n", selectedRepo.Name, color.GreenString(selectedRepo.URL))

	return nil
}

func searchRepos(opts *SearchOptions) ([]Repository, error) {
	var result []Repository
	query := fmt.Sprintf("q=%s", opts.Query)

	if opts.Topic != "" {
		query += fmt.Sprintf("+topic:%s", opts.Topic)
	}
	if opts.SearchIn != "name" {
		query += fmt.Sprintf("+in:%s", opts.SearchIn)
	}

	args := []string{
		"api", "-X", "GET", "https://api.github.com/search/repositories",
		"--jq", ".items",
		"--cache=5m",
		"-f", query,
		"-f", fmt.Sprintf("per_page=%d", opts.Limit)}

	if opts.SortBy != "" {
		args = append(args, "-f")
		args = append(args, fmt.Sprintf("sort=%s", strings.ToLower(opts.SortBy)))
	}

	stdOut, _, err := gh(args...)
	if err != nil {
		return result, err
	}

	err = json.Unmarshal(stdOut.Bytes(), &result)
	if err != nil {
		return result, err
	}
	return result, nil
}

func prettyPrint(i int, repo *Repository) string {
	out := fmt.Sprintf("%d %s\n", i, color.GreenString(repo.Name))

	dscript := repo.Description
	if len(dscript) > 100 {
		dscript = dscript[0:97]
		dscript += "..."
	}
	out += fmt.Sprintf("\t%s\n", dscript)

	if repo.Language != "" {
		out += fmt.Sprintf("\tLanguage: %s\n", blue.Sprintf(repo.Language))
	}

	if repo.Stars >= 1000 {
		out += yellow.Sprintf("\t★ %.1fk", float32(repo.Stars)/1000.0)
	} else {
		out += yellow.Sprintf("\t★ %d", repo.Stars)
	}
	return out
}

// call gh and write output to buffer
func gh(args ...string) (stdOut, errOut bytes.Buffer, err error) {
	ghBin, err := safeexec.LookPath("gh")
	if err != nil {
		err = fmt.Errorf("gh not found. error: %w", err)
		return
	}

	cmd := exec.Command(ghBin, args...)
	cmd.Stderr = &errOut
	cmd.Stdout = &stdOut

	err = cmd.Run()
	if err != nil {
		err = fmt.Errorf("failed to run gh. error: %w, stderr: %s", err, errOut.String())
		return
	}

	return
}

func main() {
	cmd := rootCmd()
	if err := cmd.Execute(); err != nil {
		fmt.Fprintf(os.Stderr, "%s\n", err)
		os.Exit(1)
	}
}

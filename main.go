package main

import (
	"errors"
	"fmt"
	"os"
	"strconv"
	"strings"

	"github.com/AlecAivazis/survey/v2"
	"github.com/cli/go-gh"
	"github.com/spf13/cobra"
)

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
		Use:   "gh search <repository>",
		Short: "search repositories",
		Long: `Search for GitHub repositories.

Search through names, descriptions, or readme's, 
sort by repository stats, and filter by topic.`,
		Example: `# cli repos with hacktoberfest topic
$ gh search cli --topic=hacktoberfest

# 10 most starred cli repos
$ gh search cli --sort=stars --limit=10`,
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

func runSearch(opts *SearchOptions) error {
	results, total, err := searchRepos(opts)
	if err != nil {
		return err
	}

	if total == 0 {
		fmt.Printf(`No results found for "%s"%s`, opts.Query, "\n")
	}

	var repoStrs []string
	for i, repo := range results {
		repoStrs = append(repoStrs, prettyPrint(i+1, &repo))
	}

	numResults := len(repoStrs)

	selector := &survey.Select{
		Message:  fmt.Sprintf("%d/%d Results\n", numResults, total),
		Options:  repoStrs,
		PageSize: 10,
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
	args := []string{"repo", "view", selectedRepo.NameWithOwner}
	stdOut, _, err := gh.Exec(args...)
	if err != nil {
		return err
	}
	fmt.Print(stdOut.String())

	return nil
}

func prettyPrint(i int, repo *Repository) string {
	out := fmt.Sprintf("%d %s\n", i, repo.NameWithOwner)

	dscript := repo.Description
	if len(dscript) > 100 {
		dscript = dscript[0:97]
		dscript += "..."
	}
	out += fmt.Sprintf("\t%s\n", dscript)

	lang := repo.PrimaryLanguage.Name

	if lang != "" {
		out += fmt.Sprintf("\tLanguage: %s\n", lang)
	}

	if repo.StargazerCount >= 1000 {
		out += fmt.Sprintf("\t★ %.1fk", float32(repo.StargazerCount)/1000.0)
	} else {
		out += fmt.Sprintf("\t★ %d", repo.StargazerCount)
	}
	return out
}

func main() {
	cmd := rootCmd()
	if err := cmd.Execute(); err != nil {
		fmt.Fprintf(os.Stderr, "%s\n", err)
		os.Exit(1)
	}
}

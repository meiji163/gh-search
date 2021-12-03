# GitHub CLI Repo Search 

[gh](https://github.com/cli/cli) extension to search for repositories on GitHub.

Requires gh version 2.1.0 or later

## Install
```
gh extension install meiji163/gh-search
```

## Usage
```
Usage:
  gh search <query> [flags]

Examples:
# cli repos with hacktoberfest topic
$ gh search cli --topic=hacktoberfest

# custom search with GitHub syntax
$ gh search -q="org:cli created:>2019-01-01"

Flags:
  -h, --help           help for gh
  -i, --in string      Search in "name", "description", or "readme"
  -l, --lang string    Search by programming language
  -L, --limit int      Max number of search results (default 50)
  -q, --query string   Query in GitHub syntax
  -t, --topic string   Specify a topic
```

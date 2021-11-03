# GitHub CLI Repo Search 

[gh](https://github.com/cli/cli) extension to search for repositories on GitHub.

## Install
```
gh extension install meiji163/gh-search
```

## Usage
```
Usage:
  search <repository> [flags]

Examples:
# cli repos with hacktoberfest topic
$ gh search cli --topic=hacktoberfest

# 10 most starred cli repos
$ gh search cli --sort stars --limit 10

Flags:
  -h, --help           help for search
  -i, --in string      Search in "name", "description", or "readme" (default "name")
  -L, --limit int      Max number of search results (default 30)
  -s, --sort string    Sort by "stars", "forks", or "issues"
  -t, --topic string   Specify a topic
```


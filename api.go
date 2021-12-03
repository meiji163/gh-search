package main

import (
	"fmt"

	"github.com/cli/go-gh"
)

type Repository struct {
	NameWithOwner   string
	Description     string
	StargazerCount  int
	PrimaryLanguage struct{ Name string }
}

func prepareQuery(opts *SearchOptions) string {
	query := opts.Query
	if opts.Topic != "" {
		query += fmt.Sprintf(" topic:%s", opts.Topic)
	}
	if opts.SearchIn != "" {
		query += fmt.Sprintf(" in:%s", opts.SearchIn)
	}
	if opts.Language != "" {
		query += fmt.Sprintf(" language:%s", opts.Language)
	}
	return query
}

func searchRepos(opts *SearchOptions) ([]Repository, int, error) {
	searchQuery := opts.Query

	gqlQuery := `query GetRepos($limit: Int, $query: String!){
	search(query: $query, first: $limit, type: REPOSITORY) {
		repositoryCount
		nodes {
				... on Repository {
					nameWithOwner,
					stargazerCount,
					description,
					primaryLanguage { name } 
				}
			}
		}
	}`

	variables := map[string]interface{}{
		"limit": opts.Limit,
		"query": searchQuery,
	}

	type responseData struct {
		Search struct {
			RepositoryCount int
			Nodes           []Repository
		}
	}

	var response responseData

	client, err := gh.GQLClient(nil)
	if err = client.Do(gqlQuery, variables, &response); err != nil {
		return response.Search.Nodes, response.Search.RepositoryCount, err
	}

	return response.Search.Nodes, response.Search.RepositoryCount, nil
}

package main

import (
	"context"
	"errors"
	"flag"
	"fmt"
	"log"
	"net/http"
	"os"
	"sort"
	"strings"

	"github.com/abustany/saboteur"
	"github.com/shurcooL/githubv4"
)

func main() {
	flag.Usage = func() {
		fmt.Fprintf(flag.CommandLine.Output(), `Usage: saboteur [OPTIONS]

Merges GitHub PRs according to a list of rules.

Command line options:
`)
		flag.PrintDefaults()
	}
	configFile := flag.String("config", "saboteur.yml", "Path to the config file")
	flag.Parse()

	ctx := context.Background()

	config, err := saboteur.LoadConfigFromFile(*configFile)
	if err != nil {
		log.Fatalf("error loading %s: %s", *configFile, err)
	}

	httpTransport, gitAuthHeaderSource, err := saboteur.SetupAuth(config.Auth)
	if err != nil {
		log.Fatalf("error setting up authentication: %s", err)
	}

	httpClient := &http.Client{Transport: httpTransport}

	githubToken := os.Getenv("GITHUB_TOKEN")
	if githubToken == "" {
		log.Fatal("GITHUB_TOKEN environment variable is not defined")
	}

	client := githubv4.NewClient(httpClient)

	repoList := mapKeys(config.Repositories)
	sort.Strings(repoList)

	for _, repoName := range repoList {
		owner, repo, err := parseRepoName(repoName)
		if err != nil {
			log.Printf("error parsing repo name %q :%s", repoName, err)
			continue
		}

		log.Printf("Listing mergeable MRs in repo %s...", repoName)

		mrs, err := saboteur.ListMergeableMRs(ctx, client, owner, repo, config.Repositories[repoName].Rules)
		if err != nil {
			log.Fatalf("error listing mergeable MRs: %s", err)
		}

		for _, mr := range mrs {
			log.Printf("PR #%d: will merge %s into %s", mr.Number, mr.Head, mr.BaseRef)
			if err := saboteur.Merge(ctx, gitAuthHeaderSource, owner, repo, mr); err != nil {
				log.Printf("Error merging PR #%d: %s", mr.Number, err)
			}
		}
	}
}

func mapKeys[T comparable, U any](m map[T]U) []T {
	res := make([]T, 0, len(m))

	for k := range m {
		res = append(res, k)
	}

	return res
}

func parseRepoName(name string) (owner string, repo string, err error) {
	idx := strings.IndexByte(name, '/')
	if idx == -1 {
		return "", "", errors.New("missing / separator")
	}

	return name[:idx], name[1+idx:], nil
}

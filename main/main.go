package main

import (
	"context"
	"flag"
	"fmt"
	"os"
	"path"
	"sync"

	"github.com/google/go-github/v32/github"
	"golang.org/x/oauth2"
)

var configPath string
var removed int = 0

func main() {
	parseCmdFlags()
	cfg, err := GetConfig(configPath)
	if err != nil {
		fmt.Printf("Could not parse configuration yaml: %v\n", err)
		os.Exit(1)
	}

	client := getGithubClient(cfg)

	var wg sync.WaitGroup

	issues, _, _ := client.Issues.ListByRepo(context.Background(), cfg.Repo.Owner, cfg.Repo.Name, nil)
	wg.Add(len(issues))

	for _, issue := range issues {
		go deleteIssueReactions(&wg, client, cfg, *issue.Number)
	}
	wg.Wait()
	fmt.Printf("Successfully removed %d issue reactions from repo %v\n", removed, cfg.Repo)
}

func deleteIssueReactions(wg *sync.WaitGroup, client *github.Client, cfg *Config, issueNumber int) {
	defer wg.Done()
	ctx := context.Background()
	reactions, _, _ := client.Reactions.ListIssueReactions(ctx, cfg.Repo.Owner, cfg.Repo.Name, issueNumber, nil)
	for _, reaction := range reactions {
		if *reaction.User.Login == cfg.Auth.Login {
			_, err := client.Reactions.DeleteIssueReaction(ctx, cfg.Repo.Owner, cfg.Repo.Name, issueNumber, *reaction.ID)
			if err != nil {
				fmt.Println("[ERR] Could not delete reaction ", reaction)
				fmt.Println(err)
			} else {
				removed++
			}
		}
	}
}

func getDefaultConfigPath() string {
	home, _ := os.UserHomeDir()
	return path.Join(home, ".config", "github-cli.yml")
}

func parseCmdFlags() {
	flag.StringVar(&configPath, "c", getDefaultConfigPath(), "A path to the YAML configuration file (default: "+getDefaultConfigPath()+")")
	flag.Parse()
}

func getGithubClient(cfg *Config) *github.Client {
	ctx := context.Background()
	ts := oauth2.StaticTokenSource(
		&oauth2.Token{AccessToken: cfg.Auth.Token},
	)
	tc := oauth2.NewClient(ctx, ts)
	return github.NewClient(tc)
}

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
var mutex = &sync.Mutex{}

func main() {
	parseCmdFlags()
	cfg, err := GetConfig(configPath)
	handleError(err, "[ERR] Could not parse configuration yaml\n", true)

	client := getGithubClient(cfg)

	var wg sync.WaitGroup

	issues, _, err := client.Issues.ListByRepo(context.Background(), cfg.Repo.Owner, cfg.Repo.Name, nil)
	handleError(err, "[ERR] Could not connect to GitHub API, check API token?\n", true)

	wg.Add(len(issues))

	for _, issue := range issues {
		go deleteIssueReactions(&wg, client, cfg, *issue.Number)
	}
	wg.Wait()
	fmt.Printf("Removed %d issue reactions from repo %v\n", removed, cfg.Repo)
}

func deleteIssueReactions(wg *sync.WaitGroup, client *github.Client, cfg *Config, issueNumber int) {
	defer wg.Done()
	ctx := context.Background()
	reactions, _, err := client.Reactions.ListIssueReactions(ctx, cfg.Repo.Owner, cfg.Repo.Name, issueNumber, nil)
	handleError(err, fmt.Sprintf("[WARN] Could not list reactions for issue %d", issueNumber), false)
	for _, reaction := range reactions {
		if *reaction.User.Login == cfg.Auth.Login {
			_, err := client.Reactions.DeleteIssueReaction(ctx, cfg.Repo.Owner, cfg.Repo.Name, issueNumber, *reaction.ID)
			if err != nil {
				fmt.Fprintf(os.Stderr, "[WARN] Could not delete reaction %v\nerror:\n%v", reaction, err)
			} else {
				mutex.Lock()
				defer mutex.Unlock()
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
	flag.StringVar(&configPath, "c", getDefaultConfigPath(), "A path to the YAML configuration file")
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

func handleError(err error, msg string, fatal bool) {
	if err != nil {
		fmt.Fprintf(os.Stderr, msg)
		fmt.Fprintf(os.Stderr, "%v\n", err)
		if fatal {
			os.Exit(1)
		}
	}
}

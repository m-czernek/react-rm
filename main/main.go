package main

import (
	"context"
	"flag"
	"fmt"
	"os"
	"path"
	"sync"

	"github.com/google/go-github/v33/github"
	"golang.org/x/oauth2"
)

var configPath string
var listMode bool
var removed int = 0
var mutex = &sync.Mutex{}
var userReactionNumberMap = make(map[string]int)

func main() {
	parseCmdFlags()
	cfg, err := GetConfig(configPath)
	handleError(err, "[ERR] Could not parse configuration yaml\n", true)

	client := getGithubClient(cfg)

	var wg sync.WaitGroup

	issues, resp, err := client.Issues.ListByRepo(context.Background(), cfg.Repo.Owner, cfg.Repo.Name, nil)
	handleError(err, "[ERR] Could not connect to GitHub API, check API token?\n", true)
	// add paginated results if any
	for resp.NextPage != 0 {
		opts := &github.IssueListByRepoOptions{ListOptions: github.ListOptions{Page: resp.NextPage}}
		nextIssues, nextResp, nextErr := client.Issues.ListByRepo(context.Background(), cfg.Repo.Owner, cfg.Repo.Name, opts)
		handleError(nextErr, "[ERR] Could not connect to GitHub API, check rate limits?\n", true)
		issues = append(issues, nextIssues...)
		resp = nextResp
	}

	wg.Add(len(issues))
	for _, issue := range issues {
		go deleteIssueReactions(&wg, client, cfg, *issue.Number)
	}
	wg.Wait()
	fmt.Printf("Removed %d issue reactions from repo %v\n", removed, cfg.Repo)

	if listMode {
		fmt.Printf("\n")
		for user, numOfReactions := range userReactionNumberMap {
			if numOfReactions > 3 {
				fmt.Printf("User %s has %d reactions in the repo\n", user, numOfReactions)
			}
		}
	}
}

func deleteIssueReactions(wg *sync.WaitGroup, client *github.Client, cfg *Config, issueNumber int) {
	defer wg.Done()
	ctx := context.Background()
	reactions, _, err := client.Reactions.ListIssueReactions(ctx, cfg.Repo.Owner, cfg.Repo.Name, issueNumber, nil)
	handleError(err, fmt.Sprintf("[WARN] Could not list reactions for issue %d", issueNumber), false)
	for _, reaction := range reactions {
		mutex.Lock()
		userReactionNumberMap[*reaction.User.Login] += 1
		mutex.Unlock()
		if *reaction.User.Login == cfg.Auth.Login {
			_, err := client.Reactions.DeleteIssueReaction(ctx, cfg.Repo.Owner, cfg.Repo.Name, issueNumber, *reaction.ID)
			if err != nil {
				fmt.Fprintf(os.Stderr, "[WARN] Could not delete reaction %v\nerror:\n%v", reaction, err)
			} else {
				mutex.Lock()
				removed++
				mutex.Unlock()
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
	flag.BoolVar(&listMode, "l", false, "List people with more than 3 votes")
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

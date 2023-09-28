package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"sort"

	"github.com/dotenv-org/godotenvvault"
	"github.com/google/go-github/v55/github"
)

func main() {
	err := godotenvvault.Load()
	if err != nil {
		log.Fatal("Error loading .env file")
	}

	github_username := os.Getenv("GITHUB_USERNAME")
	if github_username == "" {
		log.Fatal("GITHUB_USERNAME is not set")
	}

	fmt.Println("Github Repositories for " + github_username + " :")
	repos := getGithubClient(github_username)
	for _, repo := range repos {
		fmt.Println(repo)
	}
}

func getGithubClient(username string) []string {
	githubToken := os.Getenv("GITHUB_TOKEN")
	if githubToken == "" {
		log.Fatal("GITHUB_TOKEN is not set")
	}
	client := github.NewClient(nil).WithAuthToken(githubToken)

	opt := &github.RepositoryListOptions{Type: "public"}
	orgs, _, err := client.Repositories.List(context.Background(), username, opt)
	if err != nil {
		log.Fatal(err)
	}

	sort.SliceStable(orgs, func(i int, j int) bool {
		return orgs[i].GetUpdatedAt().Time.After(orgs[j].GetUpdatedAt().Time)
	})

	var repos []string
	for _, org := range orgs {
		repos = append(repos, *org.Name)
	}

	return repos
}
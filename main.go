package main

import (
	"context"
	"encoding/csv"
	"fmt"
	"log"
	"os"
	"sort"
	"strconv"

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
	repos := getClientRepositories(github_username)
	for _, repo := range repos {
		fmt.Println(*repo.Name)
	}

	fmt.Println("Storing in CSV file...")
	storeInCSV(repos)
}

func getClientRepositories(username string) []*github.Repository {
	githubToken := os.Getenv("GITHUB_TOKEN")
	if githubToken == "" {
		log.Fatal("GITHUB_TOKEN is not set")
	}
	client := github.NewClient(nil).WithAuthToken(githubToken)

	var allRepos []*github.Repository
	opt := &github.RepositoryListOptions{Type: "public", ListOptions: github.ListOptions{PerPage: 100}}
	for {
		repos, response, err := client.Repositories.List(context.Background(), username, opt)
		if err != nil {
			log.Fatal(err)
		}
		allRepos = append(allRepos, repos...)
		if response.NextPage == 0 {
			break
		}
		opt.Page = response.NextPage
	}

	sort.SliceStable(allRepos, func(i int, j int) bool {
		return allRepos[i].GetUpdatedAt().Time.After(allRepos[j].GetUpdatedAt().Time)
	})

	return allRepos
}

func storeInCSV(repos []*github.Repository) {
	file, err := os.Create("repos.csv")
	if err != nil {
		log.Fatal("Fail creating CSV file", err)
	}
	defer file.Close()

	write := csv.NewWriter(file)
	defer write.Flush()

	for _, repo := range repos {
		if err := write.Write([]string{
			strconv.FormatInt(repo.GetID(), 10),
			repo.GetName(),
			repo.GetFullName(),
			strconv.FormatBool(repo.GetPrivate()),
			repo.GetOwner().GetLogin(),
			repo.GetHTMLURL(),
			repo.GetCreatedAt().String(),
			repo.GetUpdatedAt().String(),
			repo.GetPushedAt().String(),
			repo.GetDescription(),
		}); err != nil {
			log.Fatal(err)
		}
	}
	if err := write.Error(); err != nil {
		log.Fatal(err)
	} else {
		fmt.Println("CSV file created successfully")
	}
}

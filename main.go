package main

import (
	"context"
	"encoding/csv"
	"fmt"
	"log"
	"os"
	"os/exec"
	"sort"
	"strconv"
	"strings"

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
		fmt.Println(repo.GetName())
	}

	fmt.Println("\nStoring in CSV file...")
	storeInCSV(repos)

	fmt.Println("\nCloning repositories...")
	clonseRepositories(repos)

	// fmt.Println("\nDetecting branch of latest commit...")
	// for _, repo := range repos {
	// 	latestBranch := detectBranchOfLatestCommit(repo)
	// }

	fmt.Println("\nPulling latest branch...")
	for _, repo := range repos {
		pullLatestBranch(repo)
	}

	fmt.Println("\nFetching all branches...")
	for _, repo := range repos {
		fetchAllBranches(repo)
	}
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

func clonseRepositories(repos []*github.Repository) {
	path := "./repos/"
	if _, err := os.Stat(path); os.IsNotExist(err) {
        os.Mkdir(path, 0755)
    }

	for _, repo := range repos {
		repoURL := repo.GetCloneURL()
		
		cmd := exec.Command("git", "clone", repoURL)
		cmd.Dir = path

		if err := cmd.Run(); err != nil {
			fmt.Println("Error cloning repository: " + repo.GetName() + " : " + repo.GetCloneURL())
		} else {
			fmt.Println("Repository cloned successfully: " + repo.GetName())
		}
	}
}

func detectBranchOfLatestCommit(repo *github.Repository) string {
	path := "./repos/" + repo.GetName()
	cmd := exec.Command("git", "for-each-ref", "--sort=-committerdate", "--count=1", "--format='%(refname:short)'", "refs/heads/")
	cmd.Dir = path

	output, err := cmd.Output()
	if err != nil {
		fmt.Println("Error detecting branch of latest commit: " + repo.GetName())
		return ""
	}
	
	branch := strings.TrimSpace(string(output))
	fmt.Println("Branch of latest commit for " + repo.GetName() + ": " + branch)
	return branch
}

func pullLatestBranch(repo *github.Repository) {
	path := "./repos/" + repo.GetName()
	cmd := exec.Command("git", "pull")
	cmd.Dir = path

	if err := cmd.Run(); err != nil {
		fmt.Println("Error pulling latest branch: " + repo.GetName())
	} else {
		fmt.Println("Latest branch pulled successfully: " + repo.GetName())
	}
}

func fetchAllBranches(repo *github.Repository) {
	path := "./repos/" + repo.GetName()
	cmd := exec.Command("git", "fetch", "--all")
	cmd.Dir = path

	if err := cmd.Run(); err != nil {
		fmt.Println("Error fetching all branches: " + repo.GetName())
	} else {
		fmt.Println("All branches fetched successfully: " + repo.GetName())
	}
}

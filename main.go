package main

import (
	"changelog-version/utils"
	"fmt"
	"os"
)

const Version string = "1.0.1"

func init() {
	vars := []string{
		"GITHUB_REPOSITORY",
		"GITHUB_TOKEN",
		"GITHUB_ACTOR",
		"VERSION",
	}

	for i, v := range vars {
		if os.Getenv(v) == "" {
			utils.Fatal("missing required environmental variable: %s", vars[i])
		}
	}
}

func main() {
	fmt.Println("- initializing action")
	a, err := new()
	if err != nil {
		utils.Fatal("initialization error: %s", err)
	}

	fmt.Println("- updating changelog file")
	if err := a.updateChangelog(); err != nil {
		utils.Fatal("error updating changelog file: %s", err)
	}

	fmt.Println("- committing changes")
	commit, err := a.commit()
	if err != nil {
		utils.Fatal("error committing changes: %s", err)
	}

	fmt.Println("- creating tags")
	if err := a.tag(commit); err != nil {
		utils.Fatal("error creating tags: %s", err)
	}

	fmt.Println("- pushing changes")
	if err := a.push(); err != nil {
		utils.Fatal("error pushing changes: %s", err)
	}

	o := map[string]string{
		"hash": commit.String(),
		"tag":  a.version,
	}
	for k, v := range o {
		fmt.Printf("- creating output: %s=%s\n", k, v)
		if err := utils.Output(k, v); err != nil {
			utils.Fatal("error defining output (%s=%s): %s", k, v, err)
		}
	}
	fmt.Println("- success")
}

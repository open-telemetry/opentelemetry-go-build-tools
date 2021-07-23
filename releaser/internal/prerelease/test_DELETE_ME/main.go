
package main

import (
	"fmt"
	"github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/plumbing"

	tools "go.opentelemetry.io/build-tools"
)

func main() {
	root, _ := tools.FindRepoRoot()

	repo, _ := git.PlainOpen(root)

	tags, _ := repo.Tags()

	var errors []error

	tags.ForEach(func (ref *plumbing.Reference) error {
		obj, _ := repo.TagObject(ref.Hash())

		if obj.Name != "" {
			errors = append(errors, fmt.Errorf("hello"))
		}

		return nil
	})

	fmt.Println(errors)
}
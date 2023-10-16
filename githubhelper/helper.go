package githubhelper

import (
	"context"
	"fmt"

	"github.com/shurcooL/githubv4"
	"golang.org/x/oauth2"
)

type GithubIssue struct {
	ID    string
	Title string
	Body  string
}

func NewClient(ctx context.Context, gitHubToken string) (*githubv4.Client, error) {
	if gitHubToken == "" {
		return nil, fmt.Errorf("GitHub token is not set")
	}

	src := oauth2.StaticTokenSource(
		&oauth2.Token{AccessToken: gitHubToken},
	)
	httpClient := oauth2.NewClient(ctx, src)
	client := githubv4.NewClient(httpClient)
	return client, nil
}

func GetIssue(ctx context.Context, client *githubv4.Client, repoOwner string, repoName string, issueNumber int) (*GithubIssue, error) {
	var query struct {
		Repository struct {
			Issue struct {
				ID    string
				Title string
				Body  string
			} `graphql:"issue(number: $issue_number)"`
		} `graphql:"repository(owner: $repo_owner, name: $repo_name)"`
	}
	variables := map[string]interface{}{
		"repo_owner":   githubv4.String(repoOwner),
		"repo_name":    githubv4.String(repoName),
		"issue_number": githubv4.Int(issueNumber),
	}

	err := client.Query(ctx, &query, variables)
	if err != nil {
		return nil, err
	}
	return &GithubIssue{
		ID:    query.Repository.Issue.ID,
		Title: query.Repository.Issue.Title,
		Body:  query.Repository.Issue.Body,
	}, nil
}

func AddComment(ctx context.Context, client *githubv4.Client, issueID string, comment string) error {
	var mutation struct {
		AddComment struct {
			ClientMutationID string
		} `graphql:"addComment(input: $input)"`
	}
	variables := githubv4.AddCommentInput{
		SubjectID: githubv4.ID(issueID),
		Body:      githubv4.String(comment),
	}

	err := client.Mutate(ctx, &mutation, variables, nil)
	if err != nil {
		return err
	}
	return nil
}

func GetLabel(ctx context.Context, client *githubv4.Client, repoOwner string, repoName string, labelName string) (string, error) {
	var query struct {
		Repository struct {
			Label struct {
				ID string
			} `graphql:"label(name: $label_name)"`
		} `graphql:"repository(owner: $repo_owner, name: $repo_name)"`
	}
	variables := map[string]interface{}{
		"repo_owner": githubv4.String(repoOwner),
		"repo_name":  githubv4.String(repoName),
		"label_name": githubv4.String(labelName),
	}

	err := client.Query(ctx, &query, variables)
	if err != nil {
		return "", err
	}
	return query.Repository.Label.ID, nil
}

func SetLabel(ctx context.Context, client *githubv4.Client, issueID string, labelID string) error {
	var mutation struct {
		AddLabelsToLabelable struct {
			ClientMutationID string
		} `graphql:"addLabelsToLabelable(input: $input)"`
	}
	variables := githubv4.AddLabelsToLabelableInput{
		LabelableID: githubv4.ID(issueID),
		LabelIDs:    []githubv4.ID{githubv4.ID(labelID)},
	}

	err := client.Mutate(ctx, &mutation, variables, nil)
	if err != nil {
		return err
	}
	return nil
}

func RemoveLabel(ctx context.Context, client *githubv4.Client, issueID string, labelID string) error {
	var mutation struct {
		RemoveLabelsFromLabelable struct {
			ClientMutationID string
		} `graphql:"removeLabelsFromLabelable(input: $input)"`
	}
	variables := githubv4.RemoveLabelsFromLabelableInput{
		LabelableID: githubv4.ID(issueID),
		LabelIDs:    []githubv4.ID{githubv4.ID(labelID)},
	}

	err := client.Mutate(ctx, &mutation, variables, nil)
	if err != nil {
		return err
	}
	return nil
}

package githubhelper

import (
	"context"
	"fmt"
	"time"

	"github.com/shurcooL/githubv4"
	"golang.org/x/oauth2"
)

type GithubIssue struct {
	ID         string
	Number     int
	Title      string
	Body       string
	URL        string
	DateFields map[string]time.Time
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
				ID     string
				Number int
				Title  string
				Body   string
				URL    string
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
		ID:         query.Repository.Issue.ID,
		Number:     query.Repository.Issue.Number,
		Title:      query.Repository.Issue.Title,
		Body:       query.Repository.Issue.Body,
		URL:        query.Repository.Issue.URL,
		DateFields: make(map[string]time.Time),
	}, nil
}

func GetProjectID(ctx context.Context, client *githubv4.Client, repoOwner string, projectNumber int) (githubv4.ID, error) {
	var query struct {
		User struct {
			Project struct {
				ID githubv4.ID
			} `graphql:"projectV2(number: $project_number)"`
		} `graphql:"user(login: $repo_owner)"`
	}
	variables := map[string]interface{}{
		"project_number": githubv4.Int(projectNumber),
		"repo_owner":     githubv4.String(repoOwner),
	}
	err := client.Query(ctx, &query, variables)
	if err != nil {
		return "", err
	}
	return query.User.Project.ID, nil
}

func GetIssuesWithProjectDateFieldValue(ctx context.Context, client *githubv4.Client, projectID githubv4.ID, fieldName string) ([]GithubIssue, error) {
	var query struct {
		ProjectNode struct {
			Project struct {
				Items struct {
					Nodes []struct {
						ID          githubv4.ID
						ContentNode struct {
							Issue struct {
								Number int
								Title  string
								URL    string
							} `graphql:"... on Issue"`
						} `graphql:"content"`
						FieldValueUnion struct {
							FieldValue struct {
								Date string
							} `graphql:"... on ProjectV2ItemFieldDateValue"`
						} `graphql:"fieldValueByName(name: $field_name)"`
					} `graphql:"nodes"`
					PageInfo struct {
						EndCursor   githubv4.String
						HasNextPage bool
					}
				} `graphql:"items(first: 100, after: $end_cursor)"`
			} `graphql:"... on ProjectV2"`
		} `graphql:"node(id: $project_id)"`
	}
	variables := map[string]interface{}{
		"project_id": githubv4.ID(projectID),
		"field_name": githubv4.String(fieldName),
	}

	var issues []GithubIssue
	hasNextPage := true
	endCursor := ""

	for hasNextPage {
		variables["end_cursor"] = githubv4.String(endCursor)
		err := client.Query(ctx, &query, variables)
		if err != nil {
			return nil, err
		}

		hasNextPage = query.ProjectNode.Project.Items.PageInfo.HasNextPage
		endCursor = string(query.ProjectNode.Project.Items.PageInfo.EndCursor)

		for _, node := range query.ProjectNode.Project.Items.Nodes {
			if node.FieldValueUnion.FieldValue.Date != "" {
				date, err := time.Parse("2006-01-02", node.FieldValueUnion.FieldValue.Date)
				if err != nil {
					return nil, err
				}
				issues = append(issues, GithubIssue{
					Number: node.ContentNode.Issue.Number,
					Title:  node.ContentNode.Issue.Title,
					URL:    node.ContentNode.Issue.URL,
					DateFields: map[string]time.Time{
						fieldName: date,
					},
				})
			}
		}
	}
	return issues, nil
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

package saboteur

import (
	"context"
	"fmt"
	"log"

	"github.com/shurcooL/githubv4"
)

type PR struct {
	Number  int
	Title   string
	BaseRef struct {
		Prefix string
		Name   string
		Target struct{ Oid string }
	}
	Labels  struct{ Nodes []struct{ Name string } } `graphql:"labels(first: 100)"`
	HeadRef struct {
		Target struct {
			Oid    string
			Commit struct {
				CheckSuites struct {
					Nodes []struct {
						Status      githubv4.CheckStatusState
						Conclusion  githubv4.CheckConclusionState
						WorkflowRun struct {
							Workflow struct{ Name string }
						}
						CheckRuns struct {
							Nodes []struct {
								Name string
							}
						} `graphql:"checkRuns(first: 10)"`
					}
				} `graphql:"checkSuites(first: 10)"`
			} `graphql:"... on Commit"`
		}
	}
	Commits struct {
		Nodes []struct {
			Commit struct {
				Parents struct {
					Nodes []struct {
						Oid string
					}
				} `graphql:"parents(first:100)"`
			}
		}
	} `graphql:"commits(first: 1)"`
}

type MergeableMR struct {
	// Number and title just for display/debug purposes
	Number int
	Title  string

	BaseRef  string // eg. refs/heads/master
	BaseHead string // SHA of BaseRef
	Head     string // Git commit SHA
}

func ListMergeableMRs(ctx context.Context, client *githubv4.Client, owner, repo string, rules []Rule) ([]MergeableMR, error) {
	var query struct {
		Repository struct {
			PullRequests struct {
				Nodes []PR
			} `graphql:"pullRequests(states: [OPEN], first: 10)"`
		} `graphql:"repository(owner: $owner, name: $name)"`
	}

	if err := client.Query(ctx, &query, map[string]interface{}{"owner": githubv4.String(owner), "name": githubv4.String(repo)}); err != nil {
		return nil, fmt.Errorf("error running query: %w", err)
	}

	var res []MergeableMR

	for _, pr := range query.Repository.PullRequests.Nodes {
		log.Printf("Checking pull request #%d: %s", pr.Number, pr.Title)

		for _, rule := range rules {
			m := prMatchesRule(pr, rule)
			if !m.Result {
				log.Printf("  %s, skipping", m.Reason)
				continue
			}

			log.Printf("  PR can be merged!")
			res = append(res, MergeableMR{
				Number:   pr.Number,
				Title:    pr.Title,
				BaseRef:  pr.BaseRef.Prefix + pr.BaseRef.Name,
				BaseHead: pr.BaseRef.Target.Oid,
				Head:     pr.HeadRef.Target.Oid,
			})
			break
		}
	}

	return res, nil
}

type MatchResult struct {
	Result bool
	Reason string
}

func prMatchesRule(pr PR, rule Rule) MatchResult {
	if rule.TargetBranch != "" && pr.BaseRef.Prefix+pr.BaseRef.Name != rule.TargetBranch {

		return MatchResult{Result: false, Reason: fmt.Sprintf("  PR doesn't have base ref %s", rule.TargetBranch)}
	}

	if len(pr.Commits.Nodes) == 0 {
		return MatchResult{Result: false, Reason: "PR has no commits"}
	}

	if !gitObjectsContain(pr.Commits.Nodes[0].Commit.Parents.Nodes, pr.BaseRef.Target.Oid) {
		return MatchResult{Result: false, Reason: "PR needs to be rebased on top of target branch"}
	}

	if len(rule.SuccessfulChecks) > 0 {
		for _, check := range rule.SuccessfulChecks {
			if check.Name == "" {
				return MatchResult{Result: false, Reason: "Check name missing"}
			}

			if !prHasSuccessfulCheck(pr, check) {
				return MatchResult{Result: false, Reason: fmt.Sprintf("PR didn't pass check %s", check)}
			}
		}
	}

	if len(rule.Labels) > 0 {
		for _, label := range rule.Labels {
			if !prHasLabel(pr, label) {
				return MatchResult{Result: false, Reason: fmt.Sprintf("PR doesn't have label %q", label)}
			}
		}
	}

	return MatchResult{Result: true}
}

func prHasSuccessfulCheck(pr PR, check Check) bool {
	for _, suite := range pr.HeadRef.Target.Commit.CheckSuites.Nodes {
		if suite.Status != githubv4.CheckStatusStateCompleted || suite.Conclusion != githubv4.CheckConclusionStateSuccess {
			continue
		}

		if check.WorkflowName != "" && suite.WorkflowRun.Workflow.Name != check.WorkflowName {
			continue
		}

		for _, run := range suite.CheckRuns.Nodes {
			if run.Name == check.Name {
				return true
			}
		}
	}

	return false
}

func prHasLabel(pr PR, label string) bool {
	for _, x := range pr.Labels.Nodes {
		if x.Name == label {
			return true
		}
	}

	return false
}

func gitObjectsContain(objects []struct{ Oid string }, oid string) bool {
	for _, obj := range objects {
		if obj.Oid == oid {
			return true
		}
	}

	return false
}

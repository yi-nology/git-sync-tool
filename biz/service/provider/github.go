package provider

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"
)

type githubProvider struct {
	baseURL string
	token   string
	client  *http.Client
}

func NewGitHubProvider(baseURL, token string) *githubProvider {
	if baseURL == "" {
		baseURL = "https://api.github.com"
	}
	return &githubProvider{
		baseURL: strings.TrimRight(baseURL, "/"),
		token:   token,
		client:  &http.Client{Timeout: 30 * time.Second},
	}
}

func (g *githubProvider) Platform() Platform { return PlatformGitHub }

func (g *githubProvider) TestConnection(ctx context.Context) (*TestConnectionResult, error) {
	var user struct {
		Login string `json:"login"`
	}
	if err := g.doRequest(ctx, "GET", "/user", nil, &user); err != nil {
		return &TestConnectionResult{Connected: false, Message: err.Error()}, nil
	}
	return &TestConnectionResult{Connected: true, Platform: string(g.Platform()), UserName: user.Login}, nil
}

func (g *githubProvider) ListRepos(ctx context.Context, opts ListRepoOptions) ([]*PlatformRepo, error) {
	path := "/user/repos"
	if opts.Owner != "" {
		path = fmt.Sprintf("/users/%s/repos", opts.Owner)
	}
	if opts.Page == 0 {
		opts.Page = 1
	}
	if opts.PerPage == 0 {
		opts.PerPage = 20
	}
	path = fmt.Sprintf("%s?page=%d&per_page=%d", path, opts.Page, opts.PerPage)
	var repos []struct {
		ID            int    `json:"id"`
		FullName      string `json:"full_name"`
		Name          string `json:"name"`
		Description   string `json:"description"`
		CloneURL      string `json:"clone_url"`
		SSHURL        string `json:"ssh_url"`
		DefaultBranch string `json:"default_branch"`
		Private       bool   `json:"private"`
	}
	if err := g.doRequest(ctx, "GET", path, nil, &repos); err != nil {
		return nil, err
	}
	result := make([]*PlatformRepo, 0, len(repos))
	for _, r := range repos {
		parts := strings.SplitN(r.FullName, "/", 2)
		owner := ""
		if len(parts) == 2 {
			owner = parts[0]
		}
		result = append(result, &PlatformRepo{
			ID: int64(r.ID), FullName: r.FullName, Name: r.Name, Owner: owner,
			Description: r.Description, CloneURL: r.CloneURL, SSHURL: r.SSHURL,
			DefaultBranch: r.DefaultBranch, Private: r.Private, Platform: g.Platform(),
		})
	}
	return result, nil
}

func (g *githubProvider) GetRepo(ctx context.Context, owner, repo string) (*PlatformRepo, error) {
	var r struct {
		ID            int    `json:"id"`
		FullName      string `json:"full_name"`
		Name          string `json:"name"`
		Description   string `json:"description"`
		CloneURL      string `json:"clone_url"`
		SSHURL        string `json:"ssh_url"`
		DefaultBranch string `json:"default_branch"`
		Private       bool   `json:"private"`
	}
	if err := g.doRequest(ctx, "GET", fmt.Sprintf("/repos/%s/%s", owner, repo), nil, &r); err != nil {
		return nil, err
	}
	parts := strings.SplitN(r.FullName, "/", 2)
	ownerR := ""
	if len(parts) == 2 {
		ownerR = parts[0]
	}
	return &PlatformRepo{
		ID: int64(r.ID), FullName: r.FullName, Name: r.Name, Owner: ownerR,
		Description: r.Description, CloneURL: r.CloneURL, SSHURL: r.SSHURL,
		DefaultBranch: r.DefaultBranch, Private: r.Private, Platform: g.Platform(),
	}, nil
}

func (g *githubProvider) CreateCR(ctx context.Context, opts CreateCROptions) (*ChangeRequest, error) {
	body := map[string]interface{}{
		"title": opts.Title, "body": opts.Description,
		"head": opts.SourceBranch, "base": opts.TargetBranch,
	}
	var pr githubPR
	if err := g.doRequest(ctx, "POST", fmt.Sprintf("/repos/%s/%s/pulls", opts.Owner, opts.Repo), body, &pr); err != nil {
		return nil, err
	}
	return pr.toCR(), nil
}

func (g *githubProvider) GetCR(ctx context.Context, owner, repo string, number int) (*ChangeRequest, error) {
	var pr githubPR
	if err := g.doRequest(ctx, "GET", fmt.Sprintf("/repos/%s/%s/pulls/%d", owner, repo, number), nil, &pr); err != nil {
		return nil, err
	}
	return pr.toCR(), nil
}

func (g *githubProvider) ListCRs(ctx context.Context, opts ListCROptions) ([]*ChangeRequest, int, error) {
	if opts.Page == 0 {
		opts.Page = 1
	}
	if opts.PerPage == 0 {
		opts.PerPage = 20
	}
	path := fmt.Sprintf("/repos/%s/%s/pulls?page=%d&per_page=%d", opts.Owner, opts.Repo, opts.Page, opts.PerPage)
	if opts.State != "" {
		path += "&state=" + mapGHStateToGitHub(opts.State)
	}
	var prs []githubPR
	if err := g.doRequest(ctx, "GET", path, nil, &prs); err != nil {
		return nil, 0, err
	}
	crs := make([]*ChangeRequest, 0, len(prs))
	for i := range prs {
		crs = append(crs, prs[i].toCR())
	}
	return crs, len(crs), nil
}

func (g *githubProvider) MergeCR(ctx context.Context, owner, repo string, number int, opts MergeCROptions) (*ChangeRequest, error) {
	body := map[string]interface{}{}
	if opts.MergeCommitMessage != "" {
		body["commit_message"] = opts.MergeCommitMessage
	}
	if opts.Squash {
		body["merge_method"] = "squash"
	}
	var result struct {
		Merged bool `json:"merged"`
	}
	if err := g.doRequest(ctx, "PUT", fmt.Sprintf("/repos/%s/%s/pulls/%d/merge", owner, repo, number), body, &result); err != nil {
		return nil, err
	}
	return g.GetCR(ctx, owner, repo, number)
}

func (g *githubProvider) CloseCR(ctx context.Context, owner, repo string, number int) (*ChangeRequest, error) {
	body := map[string]interface{}{"state": "closed"}
	var pr githubPR
	if err := g.doRequest(ctx, "PATCH", fmt.Sprintf("/repos/%s/%s/pulls/%d", owner, repo, number), body, &pr); err != nil {
		return nil, err
	}
	return pr.toCR(), nil
}

func (g *githubProvider) CreateWebhook(ctx context.Context, opts CreateWebhookOptions) (*PlatformWebhook, error) {
	events := opts.Events
	if len(events) == 0 {
		events = []string{"push", "pull_request"}
	}
	body := map[string]interface{}{
		"name": "web", "url": opts.URL, "secret": opts.Secret,
		"events": events, "active": true,
	}
	var wh struct {
		ID  int    `json:"id"`
		URL string `json:"url"`
	}
	if err := g.doRequest(ctx, "POST", fmt.Sprintf("/repos/%s/%s/hooks", opts.Owner, opts.Repo), body, &wh); err != nil {
		return nil, err
	}
	return &PlatformWebhook{ID: int64(wh.ID), URL: wh.URL}, nil
}

func (g *githubProvider) DeleteWebhook(ctx context.Context, owner, repo string, webhookID int64) error {
	return g.doRequest(ctx, "DELETE", fmt.Sprintf("/repos/%s/%s/hooks/%d", owner, repo, webhookID), nil, nil)
}

func (g *githubProvider) ListWebhooks(ctx context.Context, owner, repo string) ([]*PlatformWebhook, error) {
	var whs []struct {
		ID  int    `json:"id"`
		URL string `json:"url"`
	}
	if err := g.doRequest(ctx, "GET", fmt.Sprintf("/repos/%s/%s/hooks", owner, repo), nil, &whs); err != nil {
		return nil, err
	}
	result := make([]*PlatformWebhook, 0, len(whs))
	for _, wh := range whs {
		result = append(result, &PlatformWebhook{ID: int64(wh.ID), URL: wh.URL})
	}
	return result, nil
}

func (g *githubProvider) ParseWebhookEvent(r *http.Request, secret string) (*NormalizedEvent, error) {
	if err := g.ValidateWebhookSignature(r, secret); err != nil {
		return nil, err
	}
	body, _ := io.ReadAll(r.Body)
	r.Body = io.NopCloser(bytes.NewReader(body))

	eventType := r.Header.Get("X-GitHub-Event")

	var pl struct {
		Action string `json:"action"`
		Sender struct {
			ID    int    `json:"id"`
			Login string `json:"login"`
		} `json:"sender"`
		Repository struct {
			FullName string `json:"full_name"`
		} `json:"repository"`
		Number      int `json:"number"`
		PullRequest *struct {
			ID     int    `json:"id"`
			Number int    `json:"number"`
			Title  string `json:"title"`
			Body   string `json:"body"`
			State  string `json:"state"`
			Head   struct {
				Ref string `json:"ref"`
			} `json:"head"`
			Base struct {
				Ref string `json:"ref"`
			} `json:"base"`
			Mergeable *bool  `json:"mergeable"`
			Merged    bool   `json:"merged"`
			HTMLURL   string `json:"html_url"`
			User      struct {
				ID    int    `json:"id"`
				Login string `json:"login"`
			} `json:"user"`
			CreatedAt time.Time `json:"created_at"`
			UpdatedAt time.Time `json:"updated_at"`
		} `json:"pull_request"`
		Ref string `json:"ref"`
	}
	if err := json.Unmarshal(body, &pl); err != nil {
		return nil, err
	}

	parts := strings.SplitN(pl.Repository.FullName, "/", 2)
	er := &EventRepo{FullName: pl.Repository.FullName}
	if len(parts) == 2 {
		er.Owner = parts[0]
		er.Name = parts[1]
	}
	actor := &CRUser{ID: int64(pl.Sender.ID), Username: pl.Sender.Login}

	event := &NormalizedEvent{
		ID:     fmt.Sprintf("gh-%d-%d", time.Now().UnixNano(), pl.Number),
		Source: g.Platform(), Timestamp: time.Now(), Actor: actor, Repo: er,
	}

	switch eventType {
	case "pull_request":
		action := pl.Action
		if action == "closed" && pl.PullRequest != nil && pl.PullRequest.Merged {
			action = "merged"
		}
		event.Type = "cr." + action
		if pl.PullRequest != nil {
			mergeStatus := "unknown"
			if pl.PullRequest.Mergeable != nil {
				if *pl.PullRequest.Mergeable {
					mergeStatus = "mergeable"
				} else {
					mergeStatus = "conflicting"
				}
			}
			event.CR = &ChangeRequest{
				ID: int64(pl.PullRequest.Number), Number: pl.PullRequest.Number,
				Title: pl.PullRequest.Title, Description: pl.PullRequest.Body,
				State:        mapGHState(pl.PullRequest.State, pl.PullRequest.Merged),
				SourceBranch: pl.PullRequest.Head.Ref, TargetBranch: pl.PullRequest.Base.Ref,
				MergeStatus: mergeStatus, WebURL: pl.PullRequest.HTMLURL,
				Author:    &CRUser{ID: int64(pl.PullRequest.User.ID), Username: pl.PullRequest.User.Login},
				CreatedAt: pl.PullRequest.CreatedAt, UpdatedAt: pl.PullRequest.UpdatedAt,
			}
		}
	case "push":
		event.Type = "push"
		event.Branch = strings.TrimPrefix(pl.Ref, "refs/heads/")
	case "create":
		event.Type = "branch.created"
		event.Branch = pl.Ref
	case "delete":
		event.Type = "branch.deleted"
		event.Branch = pl.Ref
	}
	return event, nil
}

func (g *githubProvider) ValidateWebhookSignature(r *http.Request, secret string) error {
	if secret == "" {
		return nil
	}
	sig := r.Header.Get("X-Hub-Signature-256")
	if sig == "" {
		return fmt.Errorf("missing X-Hub-Signature-256 header")
	}
	return nil
}

func (g *githubProvider) doRequest(ctx context.Context, method, path string, body interface{}, result interface{}) error {
	var reqBody io.Reader
	if body != nil {
		b, err := json.Marshal(body)
		if err != nil {
			return err
		}
		reqBody = bytes.NewReader(b)
	}
	req, err := http.NewRequestWithContext(ctx, method, g.baseURL+path, reqBody)
	if err != nil {
		return err
	}
	req.Header.Set("Authorization", "Bearer "+g.token)
	req.Header.Set("Accept", "application/vnd.github.v3+json")
	req.Header.Set("Content-Type", "application/json")
	resp, err := g.client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	respBody, _ := io.ReadAll(resp.Body)
	if resp.StatusCode >= 400 {
		return fmt.Errorf("GitHub API %s %s returned %d: %s", method, path, resp.StatusCode, string(respBody))
	}
	if result != nil && resp.StatusCode != http.StatusNoContent {
		return json.Unmarshal(respBody, result)
	}
	return nil
}

type githubPR struct {
	ID     int    `json:"id"`
	Number int    `json:"number"`
	Title  string `json:"title"`
	Body   string `json:"body"`
	State  string `json:"state"`
	Head   struct {
		Ref string `json:"ref"`
	} `json:"head"`
	Base struct {
		Ref string `json:"ref"`
	} `json:"base"`
	User struct {
		ID    int    `json:"id"`
		Login string `json:"login"`
	} `json:"user"`
	HTMLURL   string    `json:"html_url"`
	Mergeable *bool     `json:"mergeable"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

func (pr *githubPR) toCR() *ChangeRequest {
	state := CRStateOpened
	if pr.State == "closed" {
		state = CRStateClosed
	}
	mergeStatus := "unknown"
	if pr.Mergeable != nil {
		if *pr.Mergeable {
			mergeStatus = "mergeable"
		} else {
			mergeStatus = "conflicting"
		}
	}
	return &ChangeRequest{
		ID: int64(pr.Number), Number: pr.Number, Title: pr.Title, Description: pr.Body,
		State: state, SourceBranch: pr.Head.Ref, TargetBranch: pr.Base.Ref,
		Author:      &CRUser{ID: int64(pr.User.ID), Username: pr.User.Login},
		MergeStatus: mergeStatus, WebURL: pr.HTMLURL,
		CreatedAt: pr.CreatedAt, UpdatedAt: pr.UpdatedAt,
	}
}

func mapGHState(state string, merged bool) CRState {
	if state == "closed" && merged {
		return CRStateMerged
	}
	if state == "closed" && !merged {
		return CRStateClosed
	}
	return CRStateOpened
}

func mapGHStateToGitHub(state CRState) string {
	switch state {
	case CRStateOpened:
		return "open"
	case CRStateClosed:
		return "closed"
	case CRStateMerged:
		return "closed"
	default:
		return "all"
	}
}

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

type giteaProvider struct {
	baseURL string
	token   string
	client  *http.Client
}

func NewGiteaProvider(baseURL, token string) *giteaProvider {
	if baseURL == "" {
		baseURL = "https://gitea.com/api/v1"
	}
	return &giteaProvider{
		baseURL: strings.TrimRight(baseURL, "/"),
		token:   token,
		client:  &http.Client{Timeout: 30 * time.Second},
	}
}

func (g *giteaProvider) Platform() Platform { return PlatformGitea }

func (g *giteaProvider) TestConnection(ctx context.Context) (*TestConnectionResult, error) {
	var user struct {
		Login string `json:"login"`
	}
	if err := g.doRequest(ctx, "GET", "/user", nil, &user); err != nil {
		return &TestConnectionResult{Connected: false, Message: err.Error()}, nil
	}
	return &TestConnectionResult{Connected: true, Platform: string(g.Platform()), UserName: user.Login}, nil
}

func (g *giteaProvider) ListRepos(ctx context.Context, opts ListRepoOptions) ([]*PlatformRepo, error) {
	path := "/repos/search"
	if opts.Page == 0 {
		opts.Page = 1
	}
	if opts.PerPage == 0 {
		opts.PerPage = 20
	}
	path = fmt.Sprintf("%s?page=%d&limit=%d", path, opts.Page, opts.PerPage)
	var result struct {
		Data []struct {
			ID            int    `json:"id"`
			FullName      string `json:"full_name"`
			Name          string `json:"name"`
			Description   string `json:"description"`
			CloneURL      string `json:"clone_url"`
			SSHURL        string `json:"ssh_url"`
			DefaultBranch string `json:"default_branch"`
			Private       bool   `json:"private"`
		} `json:"data"`
	}
	if err := g.doRequest(ctx, "GET", path, nil, &result); err != nil {
		return nil, err
	}
	repos := make([]*PlatformRepo, 0, len(result.Data))
	for _, r := range result.Data {
		parts := strings.SplitN(r.FullName, "/", 2)
		owner := ""
		if len(parts) == 2 {
			owner = parts[0]
		}
		repos = append(repos, &PlatformRepo{
			ID: int64(r.ID), FullName: r.FullName, Name: r.Name, Owner: owner,
			Description: r.Description, CloneURL: r.CloneURL, SSHURL: r.SSHURL,
			DefaultBranch: r.DefaultBranch, Private: r.Private, Platform: g.Platform(),
		})
	}
	return repos, nil
}

func (g *giteaProvider) GetRepo(ctx context.Context, owner, repo string) (*PlatformRepo, error) {
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

func (g *giteaProvider) CreateCR(ctx context.Context, opts CreateCROptions) (*ChangeRequest, error) {
	body := map[string]interface{}{
		"title": opts.Title, "body": opts.Description,
		"head": opts.SourceBranch, "base": opts.TargetBranch,
	}
	var pr giteaPR
	if err := g.doRequest(ctx, "POST", fmt.Sprintf("/repos/%s/%s/pulls", opts.Owner, opts.Repo), body, &pr); err != nil {
		return nil, err
	}
	return pr.toCR(), nil
}

func (g *giteaProvider) GetCR(ctx context.Context, owner, repo string, number int) (*ChangeRequest, error) {
	var pr giteaPR
	if err := g.doRequest(ctx, "GET", fmt.Sprintf("/repos/%s/%s/pulls/%d", owner, repo, number), nil, &pr); err != nil {
		return nil, err
	}
	return pr.toCR(), nil
}

func (g *giteaProvider) ListCRs(ctx context.Context, opts ListCROptions) ([]*ChangeRequest, int, error) {
	if opts.Page == 0 {
		opts.Page = 1
	}
	if opts.PerPage == 0 {
		opts.PerPage = 20
	}
	path := fmt.Sprintf("/repos/%s/%s/pulls?page=%d&limit=%d", opts.Owner, opts.Repo, opts.Page, opts.PerPage)
	if opts.State != "" {
		path += "&state=" + string(opts.State)
	}
	var prs []giteaPR
	if err := g.doRequest(ctx, "GET", path, nil, &prs); err != nil {
		return nil, 0, err
	}
	crs := make([]*ChangeRequest, 0, len(prs))
	for i := range prs {
		crs = append(crs, prs[i].toCR())
	}
	return crs, len(crs), nil
}

func (g *giteaProvider) MergeCR(ctx context.Context, owner, repo string, number int, opts MergeCROptions) (*ChangeRequest, error) {
	body := map[string]interface{}{
		"Do": "merge",
	}
	if opts.MergeCommitMessage != "" {
		body["MergeTitleField"] = opts.MergeCommitMessage
	}
	if opts.Squash {
		body["Do"] = "squash"
	}
	if err := g.doRequest(ctx, "POST", fmt.Sprintf("/repos/%s/%s/pulls/%d/merge", owner, repo, number), body, nil); err != nil {
		return nil, err
	}
	return g.GetCR(ctx, owner, repo, number)
}

func (g *giteaProvider) CloseCR(ctx context.Context, owner, repo string, number int) (*ChangeRequest, error) {
	body := map[string]interface{}{"state": "closed"}
	var pr giteaPR
	if err := g.doRequest(ctx, "PATCH", fmt.Sprintf("/repos/%s/%s/pulls/%d", owner, repo, number), body, &pr); err != nil {
		return nil, err
	}
	return pr.toCR(), nil
}

func (g *giteaProvider) CreateWebhook(ctx context.Context, opts CreateWebhookOptions) (*PlatformWebhook, error) {
	events := opts.Events
	if len(events) == 0 {
		events = []string{"push", "pull_request"}
	}
	body := map[string]interface{}{
		"type": "gitea", "url": opts.URL, "secret": opts.Secret,
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

func (g *giteaProvider) DeleteWebhook(ctx context.Context, owner, repo string, webhookID int64) error {
	return g.doRequest(ctx, "DELETE", fmt.Sprintf("/repos/%s/%s/hooks/%d", owner, repo, webhookID), nil, nil)
}

func (g *giteaProvider) ListWebhooks(ctx context.Context, owner, repo string) ([]*PlatformWebhook, error) {
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

func (g *giteaProvider) ParseWebhookEvent(r *http.Request, secret string) (*NormalizedEvent, error) {
	if err := g.ValidateWebhookSignature(r, secret); err != nil {
		return nil, err
	}
	body, _ := io.ReadAll(r.Body)
	r.Body = io.NopCloser(bytes.NewReader(body))

	eventType := r.Header.Get("X-Gitea-Event")
	var pl struct {
		Action string `json:"action"`
		Sender struct {
			ID    int    `json:"id"`
			Login string `json:"login"`
		} `json:"sender"`
		Repository struct {
			FullName string `json:"full_name"`
		} `json:"repository"`
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
			Merged  bool   `json:"merged"`
			HTMLURL string `json:"html_url"`
			User    struct {
				ID    int    `json:"id"`
				Login string `json:"login"`
			} `json:"user"`
			CreatedAt time.Time `json:"created_at"`
			UpdatedAt time.Time `json:"updated_at"`
		} `json:"pull_request"`
		Number int    `json:"number"`
		Ref    string `json:"ref"`
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
		ID:     fmt.Sprintf("gt-%d-%d", time.Now().UnixNano(), pl.Number),
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
			event.CR = &ChangeRequest{
				ID: int64(pl.PullRequest.Number), Number: pl.PullRequest.Number,
				Title: pl.PullRequest.Title, Description: pl.PullRequest.Body,
				State:        mapGiteaState(pl.PullRequest.State, pl.PullRequest.Merged),
				SourceBranch: pl.PullRequest.Head.Ref, TargetBranch: pl.PullRequest.Base.Ref,
				WebURL:    pl.PullRequest.HTMLURL,
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

func (g *giteaProvider) ValidateWebhookSignature(r *http.Request, secret string) error {
	if secret == "" {
		return nil
	}
	sig := r.Header.Get("X-Gitea-Signature")
	if sig == "" {
		return fmt.Errorf("missing X-Gitea-Signature header")
	}
	return nil
}

func (g *giteaProvider) doRequest(ctx context.Context, method, path string, body interface{}, result interface{}) error {
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
	req.Header.Set("Authorization", "token "+g.token)
	req.Header.Set("Content-Type", "application/json")
	resp, err := g.client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	respBody, _ := io.ReadAll(resp.Body)
	if resp.StatusCode >= 400 {
		return fmt.Errorf("Gitea API %s %s returned %d: %s", method, path, resp.StatusCode, string(respBody))
	}
	if result != nil && resp.StatusCode != http.StatusNoContent {
		return json.Unmarshal(respBody, result)
	}
	return nil
}

type giteaPR struct {
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
	Merged    bool      `json:"merged"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

func (pr *giteaPR) toCR() *ChangeRequest {
	return &ChangeRequest{
		ID: int64(pr.Number), Number: pr.Number, Title: pr.Title, Description: pr.Body,
		State:        mapGiteaState(pr.State, pr.Merged),
		SourceBranch: pr.Head.Ref, TargetBranch: pr.Base.Ref,
		Author: &CRUser{ID: int64(pr.User.ID), Username: pr.User.Login},
		WebURL: pr.HTMLURL, CreatedAt: pr.CreatedAt, UpdatedAt: pr.UpdatedAt,
	}
}

func mapGiteaState(state string, merged bool) CRState {
	if merged {
		return CRStateMerged
	}
	if state == "closed" {
		return CRStateClosed
	}
	return CRStateOpened
}

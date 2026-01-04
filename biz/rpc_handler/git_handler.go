package rpc_handler

import (
	"context"
	"fmt"

	"github.com/yi-nology/git-manage-service/biz/dal/db"
	gitSvc "github.com/yi-nology/git-manage-service/biz/service/git"
	"github.com/yi-nology/git-manage-service/biz/kitex_gen/git"
)

// GitServiceImpl implements the last service interface defined in the IDL.
type GitServiceImpl struct{}

// ListRepos implements the GitServiceImpl interface.
func (s *GitServiceImpl) ListRepos(ctx context.Context, req *git.ListReposRequest) (resp *git.ListReposResponse, err error) {
	repos, err := db.NewRepoDAO().FindAll()
	if err != nil {
		return nil, err
	}

	var repoList []*git.Repo
	for _, r := range repos {
		repoList = append(repoList, &git.Repo{
			Id:   int64(r.ID),
			Key:  r.Key,
			Name: r.Name,
			// Description: r.Description,
			RemoteUrl: r.RemoteURL,
			Status:    "active", // Simplified
		})
	}

	// Simple pagination
	start := (req.Page - 1) * req.PageSize
	end := start + req.PageSize
	if start < 0 {
		start = 0
	}
	if end > int32(len(repoList)) {
		end = int32(len(repoList))
	}
	if start > int32(len(repoList)) {
		start = int32(len(repoList))
	}

	return &git.ListReposResponse{
		Repos: repoList[start:end],
		Total: int64(len(repoList)),
	}, nil
}

// GetRepo implements the GitServiceImpl interface.
func (s *GitServiceImpl) GetRepo(ctx context.Context, req *git.GetRepoRequest) (resp *git.GetRepoResponse, err error) {
	r, err := db.NewRepoDAO().FindByKey(req.Key)
	if err != nil {
		return nil, fmt.Errorf("repo not found")
	}

	return &git.GetRepoResponse{
		Repo: &git.Repo{
			Id:   int64(r.ID),
			Key:  r.Key,
			Name: r.Name,
			// Description: r.Description,
			RemoteUrl: r.RemoteURL,
			Status:    "active",
		},
	}, nil
}

// ListBranches implements the GitServiceImpl interface.
func (s *GitServiceImpl) ListBranches(ctx context.Context, req *git.ListBranchesRequest) (resp *git.ListBranchesResponse, err error) {
	r, err := db.NewRepoDAO().FindByKey(req.RepoKey)
	if err != nil {
		return nil, fmt.Errorf("repo not found")
	}

	svc := gitSvc.NewGitService()
	branches, err := svc.ListBranchesWithInfo(r.Path)
	if err != nil {
		return nil, err
	}

	var branchList []*git.Branch
	for _, b := range branches {
		branchList = append(branchList, &git.Branch{
			Name:          b.Name,
			CommitHash:    b.Hash,
			CommitMessage: b.Message,
			Author:        b.Author,
			Date:          b.Date.String(),
			IsCurrent:     b.IsCurrent,
		})
	}

	return &git.ListBranchesResponse{
		Branches: branchList,
	}, nil
}

// CreateBranch implements the GitServiceImpl interface.
func (s *GitServiceImpl) CreateBranch(ctx context.Context, req *git.CreateBranchRequest) (resp *git.CreateBranchResponse, err error) {
	r, err := db.NewRepoDAO().FindByKey(req.RepoKey)
	if err != nil {
		return &git.CreateBranchResponse{Success: false, Message: "repo not found"}, nil
	}

	svc := gitSvc.NewGitService()
	if err := svc.CreateBranch(r.Path, req.BranchName, req.Ref); err != nil {
		return &git.CreateBranchResponse{Success: false, Message: err.Error()}, nil
	}

	return &git.CreateBranchResponse{Success: true, Message: "success"}, nil
}

// DeleteBranch implements the GitServiceImpl interface.
func (s *GitServiceImpl) DeleteBranch(ctx context.Context, req *git.DeleteBranchRequest) (resp *git.DeleteBranchResponse, err error) {
	r, err := db.NewRepoDAO().FindByKey(req.RepoKey)
	if err != nil {
		return &git.DeleteBranchResponse{Success: false, Message: "repo not found"}, nil
	}

	svc := gitSvc.NewGitService()
	if err := svc.DeleteBranch(r.Path, req.BranchName, req.Force); err != nil {
		return &git.DeleteBranchResponse{Success: false, Message: err.Error()}, nil
	}

	return &git.DeleteBranchResponse{Success: true, Message: "success"}, nil
}

package main

import (
	"context"
	git "github.com/yi-nology/git-manage-service/biz/kitex_gen/git"
)

// GitServiceImpl implements the last service interface defined in the IDL.
type GitServiceImpl struct{}

// ListRepos implements the GitServiceImpl interface.
func (s *GitServiceImpl) ListRepos(ctx context.Context, req *git.ListReposRequest) (resp *git.ListReposResponse, err error) {
	// TODO: Your code here...
	return
}

// GetRepo implements the GitServiceImpl interface.
func (s *GitServiceImpl) GetRepo(ctx context.Context, req *git.GetRepoRequest) (resp *git.GetRepoResponse, err error) {
	// TODO: Your code here...
	return
}

// ListBranches implements the GitServiceImpl interface.
func (s *GitServiceImpl) ListBranches(ctx context.Context, req *git.ListBranchesRequest) (resp *git.ListBranchesResponse, err error) {
	// TODO: Your code here...
	return
}

// CreateBranch implements the GitServiceImpl interface.
func (s *GitServiceImpl) CreateBranch(ctx context.Context, req *git.CreateBranchRequest) (resp *git.CreateBranchResponse, err error) {
	// TODO: Your code here...
	return
}

// DeleteBranch implements the GitServiceImpl interface.
func (s *GitServiceImpl) DeleteBranch(ctx context.Context, req *git.DeleteBranchRequest) (resp *git.DeleteBranchResponse, err error) {
	// TODO: Your code here...
	return
}

package branch

import (
	"strings"
	"sync"

	"github.com/sirupsen/logrus"
	"github.com/yi-nology/git-manage-service/biz/model/domain"
	"github.com/yi-nology/git-manage-service/biz/service/git"
	"github.com/yi-nology/git-manage-service/pkg/constants"
	"github.com/yi-nology/git-manage-service/pkg/logger"
)

// BranchService 分支业务服务
type BranchService struct {
	gitSvc *git.GitService
}

var (
	instance *BranchService
	once     sync.Once
)

// NewBranchService 创建分支服务实例
func NewBranchService() *BranchService {
	once.Do(func() {
		instance = &BranchService{
			gitSvc: git.NewGitService(),
		}
	})
	return instance
}

// ListOptions 列表查询选项
type ListOptions struct {
	Keyword    string // 关键词过滤
	BranchType string // 分支类型: local, remote, all
	Page       int    // 页码
	PageSize   int    // 每页大小
	WithSync   bool   // 是否获取同步状态
}

// ListResult 列表查询结果
type ListResult struct {
	Total int                 `json:"total"`
	List  []domain.BranchInfo `json:"list"`
}

// ListBranches 列出分支（带过滤和分页）
func (s *BranchService) ListBranches(repoPath string, opts ListOptions) (*ListResult, error) {
	logger.Debug("Listing branches", logrus.Fields{
		"repo_path": repoPath,
		"keyword":   opts.Keyword,
		"type":      opts.BranchType,
		"page":      opts.Page,
		"page_size": opts.PageSize,
	})

	// 获取所有分支
	branches, err := s.gitSvc.ListBranchesWithInfo(repoPath)
	if err != nil {
		logger.ErrorWithErr("Failed to list branches", err, logrus.Fields{
			"repo_path": repoPath,
		})
		return nil, err
	}

	// 类型过滤
	if opts.BranchType != "" && opts.BranchType != constants.BranchTypeAll {
		branches = s.filterByType(branches, opts.BranchType)
	}

	// 关键词过滤
	if opts.Keyword != "" {
		branches = s.filterByKeyword(branches, opts.Keyword)
	}

	total := len(branches)

	// 分页
	page, pageSize := s.normalizePageParams(opts.Page, opts.PageSize)
	paged := s.paginate(branches, page, pageSize)

	// 获取同步状态
	if opts.WithSync {
		s.enrichSyncStatus(repoPath, paged)
	}

	logger.Info("Branches listed successfully", logrus.Fields{
		"repo_path": repoPath,
		"total":     total,
		"returned":  len(paged),
	})

	return &ListResult{
		Total: total,
		List:  paged,
	}, nil
}

// filterByType 按类型过滤分支
func (s *BranchService) filterByType(branches []domain.BranchInfo, branchType string) []domain.BranchInfo {
	var filtered []domain.BranchInfo
	for _, b := range branches {
		if b.Type == branchType {
			filtered = append(filtered, b)
		}
	}
	return filtered
}

// filterByKeyword 按关键词过滤分支
func (s *BranchService) filterByKeyword(branches []domain.BranchInfo, keyword string) []domain.BranchInfo {
	var filtered []domain.BranchInfo
	keyword = strings.ToLower(keyword)
	for _, b := range branches {
		if strings.Contains(strings.ToLower(b.Name), keyword) ||
			strings.Contains(strings.ToLower(b.Author), keyword) {
			filtered = append(filtered, b)
		}
	}
	return filtered
}

// normalizePageParams 标准化分页参数
func (s *BranchService) normalizePageParams(page, pageSize int) (int, int) {
	if page < 1 {
		page = constants.DefaultPage
	}
	if pageSize < 1 {
		pageSize = constants.DefaultPageSize
	}
	if pageSize > constants.MaxPageSize {
		pageSize = constants.MaxPageSize
	}
	return page, pageSize
}

// paginate 执行分页
func (s *BranchService) paginate(branches []domain.BranchInfo, page, pageSize int) []domain.BranchInfo {
	start := (page - 1) * pageSize
	end := start + pageSize

	if start > len(branches) {
		start = len(branches)
	}
	if end > len(branches) {
		end = len(branches)
	}

	return branches[start:end]
}

// enrichSyncStatus 填充同步状态信息
func (s *BranchService) enrichSyncStatus(repoPath string, branches []domain.BranchInfo) {
	for i := range branches {
		b := &branches[i]
		if b.Upstream != "" {
			ahead, behind, err := s.gitSvc.GetBranchSyncStatus(repoPath, b.Name, b.Upstream)
			if err == nil {
				b.Ahead = ahead
				b.Behind = behind
			}
		}
	}
}

// CreateBranch 创建分支
func (s *BranchService) CreateBranch(repoPath, name, baseRef string) error {
	logger.Info("Creating branch", logrus.Fields{
		"repo_path": repoPath,
		"name":      name,
		"base_ref":  baseRef,
	})

	err := s.gitSvc.CreateBranch(repoPath, name, baseRef)
	if err != nil {
		logger.ErrorWithErr("Failed to create branch", err, logrus.Fields{
			"repo_path": repoPath,
			"name":      name,
		})
		return err
	}

	logger.Info("Branch created successfully", logrus.Fields{
		"repo_path": repoPath,
		"name":      name,
	})
	return nil
}

// DeleteBranch 删除分支
func (s *BranchService) DeleteBranch(repoPath, name string, force bool) error {
	logger.Info("Deleting branch", logrus.Fields{
		"repo_path": repoPath,
		"name":      name,
		"force":     force,
	})

	err := s.gitSvc.DeleteBranch(repoPath, name, force)
	if err != nil {
		logger.ErrorWithErr("Failed to delete branch", err, logrus.Fields{
			"repo_path": repoPath,
			"name":      name,
		})
		return err
	}

	logger.Info("Branch deleted successfully", logrus.Fields{
		"repo_path": repoPath,
		"name":      name,
	})
	return nil
}

// GetGitService 获取底层 Git 服务（用于复杂操作）
func (s *BranchService) GetGitService() *git.GitService {
	return s.gitSvc
}

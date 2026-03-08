package git

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
)

// PatchInfo patch 文件信息
type PatchInfo struct {
	Name    string `json:"name"`
	Path    string `json:"path"`
	Size    int64  `json:"size"`
	ModTime string `json:"mod_time"`
}

// GeneratePatch 生成 patch 内容
// base 和 target 可以是 commit hash、分支名、tag 等
func (s *GitService) GeneratePatch(path, base, target string) (string, error) {
	// 使用 git diff 生成标准 patch 格式
	cmd := exec.Command("git", "diff", base+".."+target)
	cmd.Dir = path
	output, err := cmd.CombinedOutput()
	if err != nil {
		return "", fmt.Errorf("failed to generate patch: %v. Output: %s", err, string(output))
	}
	return string(output), nil
}

// GeneratePatchForCommits 为指定的 commit 列表生成 patch
func (s *GitService) GeneratePatchForCommits(path string, commits []string) (string, error) {
	if len(commits) == 0 {
		return "", fmt.Errorf("no commits specified")
	}

	var patches []string
	for _, commit := range commits {
		cmd := exec.Command("git", "format-patch", "-1", "--stdout", commit)
		cmd.Dir = path
		output, err := cmd.CombinedOutput()
		if err != nil {
			return "", fmt.Errorf("failed to generate patch for commit %s: %v", commit, err)
		}
		patches = append(patches, string(output))
	}

	return strings.Join(patches, "\n"), nil
}

// SavePatch 保存 patch 到指定路径
// commitMessage: 如果不为空，保存后自动提交到 git
func (s *GitService) SavePatch(repoPath, patchContent, patchName, customPath string, commitMessage string) (string, error) {
	var savePath string
	var patchesDir string

	if customPath != "" {
		// 使用用户指定的路径（可以是绝对路径或相对路径）
		if filepath.IsAbs(customPath) {
			patchesDir = customPath
		} else {
			// 相对路径，相对于仓库根目录
			patchesDir = filepath.Join(repoPath, customPath)
		}
	} else {
		// 默认保存在仓库的 patches 目录
		patchesDir = filepath.Join(repoPath, "patches")
	}

	// 确保目录存在
	if err := os.MkdirAll(patchesDir, 0755); err != nil {
		return "", fmt.Errorf("failed to create directory: %v", err)
	}

	savePath = filepath.Join(patchesDir, patchName)

	// 确保 .patch 后缀
	if !strings.HasSuffix(savePath, ".patch") {
		savePath += ".patch"
	}

	// 写入文件
	if err := os.WriteFile(savePath, []byte(patchContent), 0644); err != nil {
		return "", fmt.Errorf("failed to write patch file: %v", err)
	}

	// 如果指定了提交消息，自动提交
	if commitMessage != "" {
		// 获取相对路径（相对于仓库根目录）
		relPath, err := filepath.Rel(repoPath, savePath)
		if err != nil {
			relPath = savePath // 如果获取失败，使用绝对路径
		}

		// git add
		cmd := exec.Command("git", "add", relPath)
		cmd.Dir = repoPath
		if output, err := cmd.CombinedOutput(); err != nil {
			return "", fmt.Errorf("failed to stage patch file: %v. Output: %s", err, string(output))
		}

		// git commit
		cmd = exec.Command("git", "commit", "-m", commitMessage)
		cmd.Dir = repoPath
		if output, err := cmd.CombinedOutput(); err != nil {
			// 如果 commit 失败，可能是没有改动（文件已存在且内容相同），不算错误
			if !strings.Contains(string(output), "nothing to commit") {
				return "", fmt.Errorf("failed to commit patch: %v. Output: %s", err, string(output))
			}
		}
	}

	return savePath, nil
}

// ListPatches 列出仓库中的所有 patch 文件
func (s *GitService) ListPatches(path string) ([]PatchInfo, error) {
	patchesDir := filepath.Join(path, "patches")

	// 检查目录是否存在
	if _, err := os.Stat(patchesDir); os.IsNotExist(err) {
		// 也检查 .git/patches（兼容旧位置）
		patchesDir = filepath.Join(path, ".git", "patches")
		if _, err := os.Stat(patchesDir); os.IsNotExist(err) {
			return []PatchInfo{}, nil
		}
	}

	entries, err := os.ReadDir(patchesDir)
	if err != nil {
		return nil, fmt.Errorf("failed to read patches directory: %v", err)
	}

	var patches []PatchInfo
	for _, entry := range entries {
		if entry.IsDir() {
			continue
		}
		if !strings.HasSuffix(entry.Name(), ".patch") {
			continue
		}

		info, err := entry.Info()
		if err != nil {
			continue
		}

		patches = append(patches, PatchInfo{
			Name:    entry.Name(),
			Path:    filepath.Join(patchesDir, entry.Name()),
			Size:    info.Size(),
			ModTime: info.ModTime().Format("2006-01-02 15:04:05"),
		})
	}

	return patches, nil
}

// GetPatchContent 读取 patch 文件内容
func (s *GitService) GetPatchContent(path string) (string, error) {
	content, err := os.ReadFile(path)
	if err != nil {
		return "", fmt.Errorf("failed to read patch file: %v", err)
	}
	return string(content), nil
}

// DeletePatch 删除 patch 文件
func (s *GitService) DeletePatch(path string) error {
	if err := os.Remove(path); err != nil {
		return fmt.Errorf("failed to delete patch file: %v", err)
	}
	return nil
}

// ApplyPatch 应用 patch 到仓库
// signOff: 是否添加 Signed-off-by
// commitMessage: 应用后自动提交的消息（为空则不自动提交）
func (s *GitService) ApplyPatch(repoPath, patchPath string, signOff bool, commitMessage string) error {
	// 使用 git apply 应用 patch
	cmd := exec.Command("git", "apply", patchPath)
	cmd.Dir = repoPath
	output, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("failed to apply patch: %v. Output: %s", err, string(output))
	}

	// 如果指定了提交消息，自动提交
	if commitMessage != "" {
		// 添加所有更改
		if err := s.AddAll(repoPath); err != nil {
			return fmt.Errorf("failed to stage changes: %v", err)
		}

		// 构建提交命令
		args := []string{"commit", "-m", commitMessage}
		if signOff {
			args = append(args, "--signoff")
		}

		cmd := exec.Command("git", args...)
		cmd.Dir = repoPath
		output, err := cmd.CombinedOutput()
		if err != nil {
			return fmt.Errorf("failed to commit: %v. Output: %s", err, string(output))
		}
	}

	return nil
}

// ApplyPatchFromContent 从内容应用 patch
func (s *GitService) ApplyPatchFromContent(repoPath, patchContent string, signOff bool, commitMessage string) error {
	// 创建临时 patch 文件
	tmpFile, err := os.CreateTemp("", "patch-*.patch")
	if err != nil {
		return fmt.Errorf("failed to create temp file: %v", err)
	}
	defer os.Remove(tmpFile.Name())

	if _, err := tmpFile.WriteString(patchContent); err != nil {
		tmpFile.Close()
		return fmt.Errorf("failed to write patch content: %v", err)
	}
	tmpFile.Close()

	return s.ApplyPatch(repoPath, tmpFile.Name(), signOff, commitMessage)
}

// CheckApplyDryRun 检查 patch 是否可以应用（dry-run）
func (s *GitService) CheckApplyDryRun(repoPath, patchPath string) error {
	cmd := exec.Command("git", "apply", "--check", patchPath)
	cmd.Dir = repoPath
	output, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("patch cannot be applied: %v. Output: %s", err, string(output))
	}
	return nil
}

// GetPatchStats 获取 patch 的统计信息
func (s *GitService) GetPatchStats(repoPath, patchPath string) (map[string]interface{}, error) {
	// 使用 git apply --stat 查看统计
	cmd := exec.Command("git", "apply", "--stat", patchPath)
	cmd.Dir = repoPath
	output, err := cmd.CombinedOutput()
	if err != nil {
		return nil, fmt.Errorf("failed to get patch stats: %v", err)
	}

	stats := map[string]interface{}{
		"stat":   string(output),
		"can_apply": true,
	}

	// 检查是否可以应用
	if err := s.CheckApplyDryRun(repoPath, patchPath); err != nil {
		stats["can_apply"] = false
		stats["error"] = err.Error()
	}

	return stats, nil
}

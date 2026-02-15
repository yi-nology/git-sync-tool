package git

import (
	"fmt"
	"strings"

	"github.com/sirupsen/logrus"
	"github.com/yi-nology/git-manage-service/pkg/logger"
)

// NextVersionInfo 下一版本信息
type NextVersionInfo struct {
	Current   string `json:"current"`
	NextMajor string `json:"next_major"`
	NextMinor string `json:"next_minor"`
	NextPatch string `json:"next_patch"`
}

// GetDescribe 获取 git describe 输出
func (s *GitService) GetDescribe(path string) (string, error) {
	logger.Debug("Getting git describe", logrus.Fields{"path": path})
	return s.RunCommand(path, "describe", "--tags", "--always", "--long")
}

// GetLatestVersion 获取最新版本标签
func (s *GitService) GetLatestVersion(path string) (string, error) {
	logger.Debug("Getting latest version", logrus.Fields{"path": path})

	out, err := s.RunCommand(path, "describe", "--tags", "--abbrev=0")
	if err != nil {
		logger.Debug("No tags found", logrus.Fields{"path": path, "error": err.Error()})
		return "", err
	}
	return out, nil
}

// GetNextVersions 计算下一个版本号
func (s *GitService) GetNextVersions(path string) (*NextVersionInfo, error) {
	logger.Debug("Getting next versions", logrus.Fields{"path": path})

	// 获取当前最新版本
	latest, err := s.GetLatestVersion(path)
	if err != nil || latest == "" {
		latest = "v0.0.0"
	}

	// 解析版本号
	version := latest
	hasV := false
	if strings.HasPrefix(version, "v") {
		hasV = true
		version = version[1:]
	}

	parts := strings.Split(version, ".")
	major, minor, patch := 0, 0, 0

	if len(parts) >= 1 {
		fmt.Sscanf(parts[0], "%d", &major)
	}
	if len(parts) >= 2 {
		fmt.Sscanf(parts[1], "%d", &minor)
	}
	if len(parts) >= 3 {
		fmt.Sscanf(parts[2], "%d", &patch)
	}

	// 计算下一版本
	nextMajor := fmt.Sprintf("%d.0.0", major+1)
	nextMinor := fmt.Sprintf("%d.%d.0", major, minor+1)
	nextPatch := fmt.Sprintf("%d.%d.%d", major, minor, patch+1)

	if hasV {
		nextMajor = "v" + nextMajor
		nextMinor = "v" + nextMinor
		nextPatch = "v" + nextPatch
	}

	result := &NextVersionInfo{
		Current:   latest,
		NextMajor: nextMajor,
		NextMinor: nextMinor,
		NextPatch: nextPatch,
	}

	logger.Debug("Next versions calculated", logrus.Fields{
		"current":    result.Current,
		"next_major": result.NextMajor,
		"next_minor": result.NextMinor,
		"next_patch": result.NextPatch,
	})

	return result, nil
}

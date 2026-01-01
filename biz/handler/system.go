package handler

import (
	"context"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"github.com/cloudwego/hertz/pkg/app"
	"github.com/cloudwego/hertz/pkg/protocol/consts"
)

type ListDirsReq struct {
	Path   string `query:"path"`
	Search string `query:"search"`
}

type DirItem struct {
	Name string `json:"name"`
	Path string `json:"path"`
}

type ListDirsResp struct {
	Parent  string    `json:"parent"`
	Current string    `json:"current"`
	Dirs    []DirItem `json:"dirs"`
}

// @Summary List directories for file browser
// @Tags System
// @Param path query string false "Current path"
// @Param search query string false "Search term"
// @Produce json
// @Success 200 {object} ListDirsResp
// @Router /api/system/dirs [get]
func ListDirs(ctx context.Context, c *app.RequestContext) {
	var req ListDirsReq
	if err := c.BindAndValidate(&req); err != nil {
		c.JSON(consts.StatusBadRequest, map[string]string{"error": err.Error()})
		return
	}

	currentPath := req.Path
	if currentPath == "" {
		// Default to user home or root
		home, err := os.UserHomeDir()
		if err != nil {
			currentPath = "/"
		} else {
			currentPath = home
		}
	}

	entries, err := os.ReadDir(currentPath)
	if err != nil {
		c.JSON(consts.StatusInternalServerError, map[string]string{"error": err.Error()})
		return
	}

	var dirs []DirItem
	for _, entry := range entries {
		if entry.IsDir() && !strings.HasPrefix(entry.Name(), ".") {
			if req.Search != "" && !strings.Contains(strings.ToLower(entry.Name()), strings.ToLower(req.Search)) {
				continue
			}
			dirs = append(dirs, DirItem{
				Name: entry.Name(),
				Path: filepath.Join(currentPath, entry.Name()),
			})
		}
	}

	// Sort by name
	sort.Slice(dirs, func(i, j int) bool {
		return dirs[i].Name < dirs[j].Name
	})

	parent := filepath.Dir(currentPath)
	if currentPath == "/" {
		parent = ""
	}

	c.JSON(consts.StatusOK, ListDirsResp{
		Parent:  parent,
		Current: currentPath,
		Dirs:    dirs,
	})
}

type SSHKey struct {
	Name string `json:"name"`
	Path string `json:"path"`
}

// @Summary List available SSH keys
// @Description List public SSH keys available in the user's home .ssh directory.
// @Tags System
// @Produce json
// @Success 200 {array} SSHKey
// @Failure 500 {object} map[string]string "Internal Server Error"
// @Router /api/system/ssh-keys [get]
func ListSSHKeys(ctx context.Context, c *app.RequestContext) {
	home, err := os.UserHomeDir()
	if err != nil {
		c.JSON(consts.StatusInternalServerError, map[string]string{"error": "cannot find home dir"})
		return
	}

	sshDir := filepath.Join(home, ".ssh")
	entries, err := os.ReadDir(sshDir)
	if err != nil {
		// If .ssh dir doesn't exist, return empty
		c.JSON(consts.StatusOK, []SSHKey{})
		return
	}

	var keys []SSHKey
	for _, entry := range entries {
		if !entry.IsDir() {
			keys = append(keys, SSHKey{
				Name: entry.Name(),
				Path: filepath.Join(sshDir, entry.Name()),
			})
		}
	}

	c.JSON(consts.StatusOK, keys)
}

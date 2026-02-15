// biz/service/git/git_file.go - Git文件浏览服务

package git

import (
	"encoding/base64"
	"io"
	"path/filepath"
	"strings"
	"unicode/utf8"

	"github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/plumbing"
	"github.com/go-git/go-git/v5/plumbing/filemode"
	"github.com/go-git/go-git/v5/plumbing/object"
)

// TreeEntry 目录树条目
type TreeEntry struct {
	Name string `json:"name"`
	Path string `json:"path"`
	Type string `json:"type"` // "file" 或 "dir"
	Size int64  `json:"size"`
	Mode string `json:"mode"`
	Hash string `json:"hash"`
}

// BlobContent 文件内容
type BlobContent struct {
	Content  string `json:"content"`
	Encoding string `json:"encoding"` // "utf-8" 或 "base64"
	Size     int64  `json:"size"`
	IsBinary bool   `json:"is_binary"`
	MimeType string `json:"mime_type"`
}

// FileCommit 文件提交记录
type FileCommit struct {
	Hash      string `json:"hash"`
	ShortHash string `json:"short_hash"`
	Message   string `json:"message"`
	Author    string `json:"author"`
	Date      string `json:"date"`
}

// GetTree 获取目录树
func (s *GitService) GetTree(repoPath, ref, dirPath string, recursive bool) ([]TreeEntry, error) {
	r, err := s.openRepo(repoPath)
	if err != nil {
		return nil, err
	}

	// 解析ref到commit
	commit, err := s.resolveCommit(r, ref)
	if err != nil {
		return nil, err
	}

	// 获取tree
	tree, err := commit.Tree()
	if err != nil {
		return nil, err
	}

	// 如果指定了路径，获取子树
	if dirPath != "" && dirPath != "/" {
		dirPath = strings.TrimPrefix(dirPath, "/")
		tree, err = tree.Tree(dirPath)
		if err != nil {
			return nil, err
		}
	}

	var entries []TreeEntry

	if recursive {
		// 递归获取所有文件
		err = tree.Files().ForEach(func(f *object.File) error {
			entries = append(entries, TreeEntry{
				Name: filepath.Base(f.Name),
				Path: f.Name,
				Type: "file",
				Size: f.Size,
				Mode: f.Mode.String(),
				Hash: f.Hash.String(),
			})
			return nil
		})
	} else {
		// 只获取当前目录
		for _, entry := range tree.Entries {
			entryType := "file"
			if entry.Mode == filemode.Dir {
				entryType = "dir"
			}

			var size int64 = 0
			if entry.Mode.IsFile() {
				// 获取文件大小
				blob, err := r.BlobObject(entry.Hash)
				if err == nil {
					size = blob.Size
				}
			}

			path := entry.Name
			if dirPath != "" {
				path = filepath.Join(dirPath, entry.Name)
			}

			entries = append(entries, TreeEntry{
				Name: entry.Name,
				Path: path,
				Type: entryType,
				Size: size,
				Mode: entry.Mode.String(),
				Hash: entry.Hash.String(),
			})
		}
	}

	return entries, err
}

// GetBlob 获取文件内容
func (s *GitService) GetBlob(repoPath, ref, filePath string) (*BlobContent, error) {
	r, err := s.openRepo(repoPath)
	if err != nil {
		return nil, err
	}

	// 解析ref到commit
	commit, err := s.resolveCommit(r, ref)
	if err != nil {
		return nil, err
	}

	// 获取文件
	filePath = strings.TrimPrefix(filePath, "/")
	file, err := commit.File(filePath)
	if err != nil {
		return nil, err
	}

	// 读取内容
	reader, err := file.Reader()
	if err != nil {
		return nil, err
	}
	defer reader.Close()

	content, err := io.ReadAll(reader)
	if err != nil {
		return nil, err
	}

	// 判断是否是二进制文件
	isBinary := !utf8.Valid(content) || containsNullByte(content)

	result := &BlobContent{
		Size:     file.Size,
		IsBinary: isBinary,
		MimeType: getMimeType(filePath),
	}

	if isBinary {
		result.Content = base64.StdEncoding.EncodeToString(content)
		result.Encoding = "base64"
	} else {
		result.Content = string(content)
		result.Encoding = "utf-8"
	}

	return result, nil
}

// GetFileHistory 获取文件的提交历史
func (s *GitService) GetFileHistory(repoPath, ref, filePath string, limit int) ([]FileCommit, error) {
	r, err := s.openRepo(repoPath)
	if err != nil {
		return nil, err
	}

	// 解析ref
	hash, err := r.ResolveRevision(plumbing.Revision(ref))
	if err != nil {
		return nil, err
	}

	// 获取日志
	filePath = strings.TrimPrefix(filePath, "/")
	iter, err := r.Log(&git.LogOptions{
		From:     *hash,
		FileName: &filePath,
	})
	if err != nil {
		return nil, err
	}

	if limit <= 0 {
		limit = 50
	}

	var commits []FileCommit
	count := 0
	err = iter.ForEach(func(c *object.Commit) error {
		if count >= limit {
			return io.EOF
		}
		commits = append(commits, FileCommit{
			Hash:      c.Hash.String(),
			ShortHash: c.Hash.String()[:7],
			Message:   strings.Split(c.Message, "\n")[0],
			Author:    c.Author.Name,
			Date:      c.Author.When.Format("2006-01-02 15:04:05"),
		})
		count++
		return nil
	})

	if err != nil && err != io.EOF {
		return nil, err
	}

	return commits, nil
}

// containsNullByte 检查是否包含空字节
func containsNullByte(data []byte) bool {
	for _, b := range data {
		if b == 0 {
			return true
		}
	}
	return false
}

// getMimeType 根据文件扩展名获取MIME类型
func getMimeType(filename string) string {
	ext := strings.ToLower(filepath.Ext(filename))
	mimeTypes := map[string]string{
		".go":   "text/x-go",
		".js":   "application/javascript",
		".ts":   "application/typescript",
		".py":   "text/x-python",
		".java": "text/x-java",
		".c":    "text/x-c",
		".cpp":  "text/x-c++",
		".h":    "text/x-c",
		".rs":   "text/x-rust",
		".rb":   "text/x-ruby",
		".php":  "text/x-php",
		".sh":   "text/x-shellscript",
		".md":   "text/markdown",
		".json": "application/json",
		".xml":  "application/xml",
		".yaml": "text/yaml",
		".yml":  "text/yaml",
		".html": "text/html",
		".css":  "text/css",
		".sql":  "text/x-sql",
		".txt":  "text/plain",
		".png":  "image/png",
		".jpg":  "image/jpeg",
		".jpeg": "image/jpeg",
		".gif":  "image/gif",
		".svg":  "image/svg+xml",
		".pdf":  "application/pdf",
		".zip":  "application/zip",
	}
	if mime, ok := mimeTypes[ext]; ok {
		return mime
	}
	return "application/octet-stream"
}

package api

type DirItem struct {
	Name string `json:"name"`
	Path string `json:"path"`
}

type ListDirsResp struct {
	Parent  string    `json:"parent"`
	Current string    `json:"current"`
	Dirs    []DirItem `json:"dirs"`
}

type SSHKey struct {
	Name string `json:"name"`
	Path string `json:"path"`
}

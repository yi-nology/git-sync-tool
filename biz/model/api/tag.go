package api

type CreateTagReq struct {
	TagName    string `json:"tag_name" binding:"required"`
	Ref        string `json:"ref" binding:"required"` // Branch name or Commit Hash
	Message    string `json:"message"`
	PushRemote string `json:"push_remote"` // Optional: Remote name to push to
}

type PushTagReq struct {
	TagName string `json:"tag_name" binding:"required"`
	Remote  string `json:"remote" binding:"required"`
}

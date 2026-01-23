package hz

import (
	"github.com/cloudwego/hertz/pkg/app/server"
	"github.com/yi-nology/git-manage-service/biz/router"
)

// GeneratedRegister registers all routes
func GeneratedRegister(h *server.Hertz) {
	router.GeneratedRegister(h)
}

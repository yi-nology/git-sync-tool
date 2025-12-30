package handler

import (
	"context"
	"github.com/yi-nology/git-sync-tool/biz/config"
	"net/http"

	"github.com/cloudwego/hertz/pkg/app"
)

type ConfigReq struct {
	DebugMode bool `json:"debug_mode"`
}

// @Summary Get global configuration
// @Tags Config
// @Produce json
// @Success 200 {object} map[string]interface{}
// @Router /api/config [get]
func GetConfig(ctx context.Context, c *app.RequestContext) {
	c.JSON(http.StatusOK, map[string]interface{}{
		"debug_mode": config.DebugMode,
	})
}

// @Summary Update global configuration
// @Tags Config
// @Accept json
// @Produce json
// @Param request body ConfigReq true "Config info"
// @Success 200 {object} map[string]interface{}
// @Router /api/config [post]
func UpdateConfig(ctx context.Context, c *app.RequestContext) {
	var req ConfigReq
	if err := c.BindAndValidate(&req); err != nil {
		c.JSON(http.StatusBadRequest, map[string]string{"error": err.Error()})
		return
	}

	config.DebugMode = req.DebugMode
	c.JSON(http.StatusOK, map[string]interface{}{
		"debug_mode": config.DebugMode,
	})
}

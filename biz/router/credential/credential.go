package credential

import (
	"github.com/cloudwego/hertz/pkg/app/server"
	credentialHandler "github.com/yi-nology/git-manage-service/biz/handler/credential"
)

// Register register credential routes
func Register(r *server.Hertz) {
	root := r.Group("/")
	{
		_api := root.Group("/api")
		{
			_v1 := _api.Group("/v1")
			{
				_credentials := _v1.Group("/credentials")
				_credentials.GET("/", credentialHandler.List)
				_credentials.POST("/", credentialHandler.Create)
				_credentials.POST("/match", credentialHandler.Match)
				_credentials.GET("/:id", credentialHandler.Get)
				_credentials.PUT("/:id", credentialHandler.Update)
				_credentials.DELETE("/:id", credentialHandler.Delete)
				_credentials.POST("/:id/test", credentialHandler.Test)
			}
		}
	}
}

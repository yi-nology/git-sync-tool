package sshkey

import (
	"github.com/cloudwego/hertz/pkg/app/server"
	sshkeyHandler "github.com/yi-nology/git-manage-service/biz/handler/sshkey"
)

// Register register routes based on the IDL 'api.${HTTP Method}' annotation.
func Register(r *server.Hertz) {

	root := r.Group("/", rootMw()...)
	{
		_api := root.Group("/api", _apiMw()...)
		{
			_v1 := _api.Group("/v1", _v1Mw()...)
			{
				_system := _v1.Group("/system", _systemMw()...)
				{
					_dbsshkeys := _system.Group("/db-ssh-keys", _dbsshkeysMw()...)
					_dbsshkeys.GET("/", append(_listdbsshkeysMw(), sshkeyHandler.ListDBSSHKeys)...)
					_dbsshkeys.POST("/", append(_createdbsshkeyMw(), sshkeyHandler.CreateDBSSHKey)...)
					_dbsshkeys.GET("/:id", append(_getdbsshkeyMw(), sshkeyHandler.GetDBSSHKey)...)
					_dbsshkeys.PUT("/:id", append(_updatedbsshkeyMw(), sshkeyHandler.UpdateDBSSHKey)...)
					_dbsshkeys.DELETE("/:id", append(_deletedbsshkeyMw(), sshkeyHandler.DeleteDBSSHKey)...)
					_dbsshkeys.POST("/:id/test", append(_testdbsshkeyMw(), sshkeyHandler.TestDBSSHKey)...)
				}
			}
		}
	}
}

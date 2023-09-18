package web_api

import (
	. "github.com/Bunny3th/easy-workflow/workflow/web_api/router"
	"github.com/gin-gonic/gin"
)

/*开启工作流引擎WebApi	参数说明:
engine: *gin.Engine gin框架(github.com/gin-gonic/gin)
GinMode可选项: debug | release
addr 监听地址与端口,如"localhost:8080"
*/
func StartWebApi(engine *gin.Engine, ApiBaseUrl string, ShowSwaggerDoc bool,SwaggerUrl string, addr string) {
	e := NewRouter(engine, ApiBaseUrl, ShowSwaggerDoc,SwaggerUrl)
	e.Run(addr)

	//如何使用swagger生成文档
	//一般在main包所在目录执行 swag init
	//但本项目中，swagger命令需要在web_api目录中加上-d参数执行，如下
	//swag init --generalInfo Launcher.go --instanceName easyworkflow -d ./,../model,../database
}

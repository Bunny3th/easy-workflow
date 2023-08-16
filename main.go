package main

import (
	"easy-workflow/api/router"
	. "easy-workflow/workflow/config"

)

// @title easy-workflow工作流引擎API
// @version 1.0.0
// @description 演示说明文档
// @contact.name go-swagger帮助文档
// @contact.url https://github.com/swaggo/swag/blob/master/README_zh-CN.md
// @host localhost:8180
// @BasePath /
func main() {
	//示例:开启一个子协程，一般可用在定时任务
	//go func() {
	//	for {
    //      do something
	//	}
	//}()
	router := router.NewRouter()
	router.Run(Gin.Port)

	//这是使用证书走https协议的示例。但不建议在程序中使用，应通过nginx发布时定义
	//router.RunTLS(":8080","D:\\nginx-1.21.6\\conf\\website_config\\cert\\dyuanzi.com.pem",
	//	"D:\\nginx-1.21.6\\conf\\website_config\\cert\\dyuanzi.com.key")

	//这里演示如何编译成linux下可执行文件 -o 参数指定生成文件名
	//set GOARCH=amd64
	//set GOOS=linux
	//go build -o easy_linux main.go

	//如何使用swagger生成文档
	//一般在main包所在目录执行 swag init
	//但本项目中，swagger命令需要加上-d参数执行，如下
	//swag init -d ./,../pkg/workflow/model

}

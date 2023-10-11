package router

import (
	 "github.com/Bunny3th/easy-workflow/workflow/web_api/docs" // 导入swagger文档用的
	. "github.com/Bunny3th/easy-workflow/workflow/web_api/service"
	"github.com/gin-gonic/gin"
	"github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"
)

func NewRouter(engine *gin.Engine, ApiBasePath string, ShowSwaggerDoc bool, SwaggerUrl string) *gin.Engine {
	//注意，由于我们执行swag init的时候指定了InstanceName，所以这里也必须传入InstanceName
	if ShowSwaggerDoc {
		engine.GET(SwaggerUrl, ginSwagger.WrapHandler(swaggerFiles.Handler, func(c *ginSwagger.Config) {
			c.InstanceName = "easyworkflow"
		}))
	}
	//swagger信息设置
	docs.SwaggerInfoeasyworkflow.BasePath=ApiBasePath
	docs.SwaggerInfoeasyworkflow.Title="Easy WorkFlow接口说明"
	docs.SwaggerInfoeasyworkflow.Description="https://github.com/Bunny3th/easy-workflow"

	router := engine.Group(ApiBasePath)

	router.POST("/def/save", ProcDef_Save)
	router.GET("/def/list", ProcDef_ListBySource)
	router.GET("/def/get", ProcDef_GetProcDefByID)

	router.POST("/inst/start", ProcInst_Start)
	router.GET("/inst/start/by", ProcInst_StartByUser)
	router.POST("/inst/revoke", ProcInst_Revoke)
	router.GET("/inst/task_history", ProcInst_TaskHistory)

	router.POST("/task/pass", Task_Pass)
	router.POST("/task/pass/directly", Task_Pass_DirectlyToWhoRejectedMe)
	router.POST("/task/reject", Task_Reject)
	router.POST("/task/reject/free", Task_FreeRejectToUpstreamNode)
	router.POST("/task/transfer",Task_Transfer)
	router.GET("/task/todo", Task_ToDoList)
	router.GET("/task/finished", Task_FinishedList)
	router.GET("/task/upstream", Task_UpstreamNodeList)
	router.GET("/task/action", Task_WhatCanIDo)
	router.GET("/task/info", Task_Info)

	return engine
}

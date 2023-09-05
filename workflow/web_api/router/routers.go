package router

import (
	_ "github.com/Bunny3th/easy-workflow/workflow/web_api/docs" // 导入swagger文档用的
	. "github.com/Bunny3th/easy-workflow/workflow/web_api/service"
	"github.com/gin-gonic/gin"
	"github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"
)

func NewRouter(engine *gin.Engine,BaseUrl string,ShowSwaggerDoc bool) *gin.Engine {
	router:=engine.Group(BaseUrl)

	if ShowSwaggerDoc {
		router.GET("/process/swagger/*any", ginSwagger.WrapHandler(swaggerFiles.Handler))
	}

	router.POST("/process/def/save",ProcDef_Save)
	router.GET("/process/def/list",ProcDef_ListBySource)
	router.GET("/process/def/get",ProcDef_GetProcDefByID)

	router.POST("/process/inst/start",ProcInst_Start)
	router.POST("/process/inst/revoke",ProcInst_Revoke)
	router.GET("/process/inst/task_history",ProcInst_TaskHistory)

	router.POST("/process/task/pass",Task_Pass)
	router.POST("/process/task/pass/directly",Task_Pass_DirectlyToWhoRejectedMe)

	router.POST("/process/task/reject",Task_Reject)
	router.POST("/process/task/reject/free",Task_FreeRejectToUpstreamNode)
	router.GET("/process/task/todo",Task_ToDoList)
	router.GET("/process/task/finished",Task_FinishedList)
	router.GET("/process/task/upstream",Task_UpstreamNodeList)

	return engine
}

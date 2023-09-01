package router

import (
	_ "easy-workflow/example/docs" // 导入swagger文档用的
	. "easy-workflow/example/service"
	"github.com/gin-gonic/gin"
	"github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"
)

const GinMode string="debug"

func NewRouter() *gin.Engine {
	gin.SetMode(GinMode)
	r := gin.New()

	//这里定义中间件
	r.Use(gin.Logger())      //gin的默认log，默认输出是os.Stdout，即屏幕
	r.Use(gin.Recovery())    //从任何panic中恢复，并在出现panic时返回http 500


	//只有在debug模式下才开启swagger
	if GinMode == "debug" {
		r.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerFiles.Handler))
	}


	r.POST("/process/def/save",ProcDef_Save)
	r.GET("/process/def/list",ProcDef_ListBySource)
    r.GET("/process/def/get",ProcDef_GetProcDefByID)

	r.POST("/process/inst/start",ProcInst_Start)
	r.POST("/process/inst/revoke",ProcInst_Revoke)
	r.GET("/process/inst/task_history",ProcInst_TaskHistory)

	r.POST("/process/task/pass",Task_Pass)
	r.POST("/process/task/pass/directly",Task_Pass_DirectlyToWhoRejectedMe)

	r.POST("/process/task/reject",Task_Reject)
	r.POST("/process/task/reject/free",Task_FreeRejectToUpstreamNode)
	r.GET("/process/task/todo",Task_ToDoList)
	r.GET("/process/task/finished",Task_FinishedList)
	r.GET("/process/task/upstream",Task_UpstreamNodeList)


	return r
}

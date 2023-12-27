package service

import (
	. "github.com/Bunny3th/easy-workflow/workflow/engine"
	"github.com/gin-gonic/gin"
	"net/http"
	"strconv"
	"strings"
)

/*
swagger注解描述 https://github.com/swaggo/swag/blob/master/README_zh-CN.md
@Summary	摘要
@Produce	API 可以产生的 MIME 类型的列表，MIME 类型你可以简单的理解为响应类型，例如：json、xml、html 等等,详细如下：
        ---Alias-------------------------MIME Type------------------------------
           json	                         application/json
           x-www-form-urlencoded	     application/x-www-form-urlencoded
           xml	                         text/xml
           plain	                     text/plain
           html	                         text/html
           mpfd	                         multipart/form-data
           json-api	                     application/vnd.api+json
           json-stream	                 application/x-json-stream
           octet-stream	                 application/octet-stream
           png	                         image/png
           jpeg	                         image/jpeg
           gif	                         image/gif
@Param	参数格式，从左到右分别为：参数名、入参类型、数据类型、是否必填、注释、example(示例)
        -入参类型有以下几种：path query header cookie  body formData
        -数据类型有 string int uint uint32 uint64 float32 bool 以及用户自定义类型(struct)
@Success	响应成功，从左到右分别为：状态码、参数类型、数据类型、注释
@Failure	响应失败，从左到右分别为：状态码、参数类型、数据类型、注释
@Router	路由，从左到右分别为：路由地址，HTTP 方法
*/

// @Summary      流程定义保存/升级
// @Description
// @Tags         流程定义
// @Produce      json
// @Param        Resource  formData string  true  "流程定义资源(json)" example(json字符串)
// @Param        CreateUserID  formData string  true  "创建者ID" example(0001)
// @Success      200  {object}  int 流程ID
// @Failure      400  {object}  string 报错信息
// @Router       /def/save [post]
func ProcDef_Save(c *gin.Context) {
	Resource := c.PostForm("Resource")
	CreateUserID := c.PostForm("CreateUserID")

	if ProcID, err := ProcessSave(Resource, CreateUserID); err == nil {
		c.JSON(http.StatusOK, ProcID)
	} else {
		c.JSON(400, err.Error())
	}
}

// @Summary      获取特定source下所有流程
// @Description  引擎可能被多个系统、组件等使用，source表示从哪个来源创建的流程
// @Tags         流程定义
// @Produce      json
// @Param        source  query string  true  "来源" example(办公系统)
// @Success      200  {object}  []database.ProcDef 流程定义列表
// @Failure      400  {object}  string 报错信息
// @Router       /def/list [get]
func ProcDef_ListBySource(c *gin.Context) {
	source := c.Query("source")
	if procDef, err := GetProcessList(source); err == nil {
		c.JSON(200, procDef)
	} else {
		c.JSON(400, err.Error())
	}
}

// @Summary      获取流程定义
// @Description  返回的是Node数组，流程是由N个Node组成的
// @Tags         流程定义
// @Produce      json
// @Param        id  query string  true  "流程ID" example(1)
// @Success      200  {object}  model.Process "流程定义"
// @Failure      400  {object}  string 报错信息
// @Router       /def/get [get]
func ProcDef_GetProcDefByID(c *gin.Context) {
	id, err := strconv.Atoi(c.Query("id"))
	if err != nil {
		c.JSON(400, err.Error())
		return
	}

	if nodes, err := GetProcessDefine(id); err == nil {
		c.JSON(200, nodes)
	} else {
		c.JSON(400, err.Error())
	}
}

// @Summary      开始流程
// @Description  注意，VariablesJson格式是key-value对象集合:[{"Key":"starter","Value":"U0001"}]
// @Tags         流程实例
// @Produce      json
// @Param        ProcessID  formData string  true  "流程ID" example(1)
// @Param        BusinessID  formData string  true  "业务ID" example("订单001")
// @Param        Comment  formData string  false  "评论意见" example("家中有事请假三天,请领导批准")
// @Param        VariablesJson  formData string  false  "变量(Json)" example([{"Key":"starter","Value":"U0001"}])
// @Success      200  {object}  int 流程实例ID
// @Failure      400  {object}  string 报错信息
// @Router       /inst/start [post]
func ProcInst_Start(c *gin.Context) {
	ProcessID, err := strconv.Atoi(c.PostForm("ProcessID"))
	if err != nil {
		c.JSON(400, err.Error())
		return
	}
	BusinessID := c.PostForm("BusinessID")
	Comment := c.PostForm("Comment")
	VariablesJson := c.PostForm("VariablesJson")

	if id, err := InstanceStart(ProcessID, BusinessID, Comment, VariablesJson); err == nil {
		c.JSON(200, id)
	} else {
		c.JSON(400, err.Error())
	}
}

// @Summary      撤销流程
// @Description  注意，Force 是否强制撤销，若为false,则只有流程回到发起人这里才能撤销
// @Tags         流程实例
// @Produce      json
// @Param        InstanceID  formData int  true  "流程实例ID" example(1)
// @Param        RevokeUserID  formData string  true  "撤销发起用户ID" example("U001")
// @Param        Force  formData bool  true  "是否强制撤销" example("false")
// @Success      200  {object}  string "ok"
// @Failure      400  {object}  string 报错信息
// @Router       /inst/revoke [post]
func ProcInst_Revoke(c *gin.Context) {
	InstanceID, err := strconv.Atoi(c.PostForm("InstanceID"))
	if err != nil {
		c.JSON(400, err.Error())
		return
	}

	RevokeUserID := c.PostForm("RevokeUserID")

	Force, err := strconv.ParseBool(c.PostForm("Force"))
	if err != nil {
		c.JSON(400, err.Error())
		return
	}

	if err := InstanceRevoke(InstanceID, Force, RevokeUserID); err == nil {
		c.JSON(200, "ok")
	} else {
		c.JSON(400, err.Error())
	}
}

// @Summary      流程实例中任务执行记录
// @Description
// @Tags         流程实例
// @Produce      json
// @Param        instid  query int  true  "流程实例ID" example(1)
// @Success      200  {object}  []model.Task "任务列表"
// @Failure      400  {object}  string 报错信息
// @Router       /inst/task_history [get]
func ProcInst_TaskHistory(c *gin.Context) {
	InstanceID, err := strconv.Atoi(c.Query("instid"))
	if err != nil {
		c.JSON(400, err.Error())
		return
	}
	if tasklist, err := GetInstanceTaskHistory(InstanceID); err == nil {
		c.JSON(200, tasklist)
	} else {
		c.JSON(400, err.Error())
	}
}

// @Summary      获取起始人为特定用户的流程实例
// @Description
// @Tags         流程实例
// @Produce      json
// @Param        userid  query string  true  "用户ID 传入空则获取所有用户的流程实例" example("U001")
// @Param        procname  query string  false  "指定流程名称，非必填" example("请假")
// @Param        idx  query int  true  "分页用,开始index" example(0)
// @Param        rows  query int  true  "分页用,最大返回行数" example(0)
// @Success      200  {object}  []model.Instance "流程实例列表"
// @Failure      400  {object}  string 报错信息
// @Router       /inst/start/by [get]
func ProcInst_StartByUser(c *gin.Context) {
	UserID := c.Query("userid")
	ProcessName := c.Query("procname")
	StartIndex, err := strconv.Atoi(c.Query("idx"))
	if err != nil {
		c.JSON(400, err.Error())
		return
	}
	MaxRow, err := strconv.Atoi(c.Query("rows"))
	if err != nil {
		c.JSON(400, err.Error())
		return
	}

	if insts, err := GetInstanceStartByUser(UserID, ProcessName, StartIndex, MaxRow); err == nil {
		c.JSON(200, insts)
	} else {
		c.JSON(400, err.Error())
	}
}

// @Summary      任务通过
// @Description  任务通过后根据流程定义，进入下一个节点进行处理
// @Tags         任务
// @Produce      json
// @Param        TaskID  formData int  true  "任务ID" example(1)
// @Param        Comment  formData string  false  "评论意见" example("同意请假")
// @Param        VariablesJson  formData string  false  "变量(Json)" example("{"User":"001"}")
// @Success      200  {object}  string "ok"
// @Failure      400  {object}  string 报错信息
// @Router       /task/pass [post]
func Task_Pass(c *gin.Context) {
	TaskID, err := strconv.Atoi(c.PostForm("TaskID"))
	if err != nil {
		c.JSON(400, err.Error())
		return
	}
	Comment := c.PostForm("Comment")
	VariableJson := c.PostForm("VariableJson")

	if err := TaskPass(TaskID, Comment, VariableJson, false); err == nil {
		c.JSON(200, "ok")
	} else {
		c.JSON(400, err.Error())
	}
}

// @Summary      任务通过后流程直接返回到上一个驳回我的节点
// @Description  此功能只有在非会签节点时才能使用
// @Tags         任务
// @Produce      json
// @Param        TaskID  formData int  true  "任务ID" example(1)
// @Param        Comment  formData string  false  "评论意见" example("同意请假")
// @Param        VariablesJson  formData string  false  "变量(Json)" example("{"User":"001"}")
// @Success      200  {object}  string "ok"
// @Failure      400  {object}  string 报错信息
// @Router       /task/pass/directly [post]
func Task_Pass_DirectlyToWhoRejectedMe(c *gin.Context) {
	TaskID, err := strconv.Atoi(c.PostForm("TaskID"))
	if err != nil {
		c.JSON(400, err.Error())
		return
	}
	Comment := c.PostForm("Comment")
	VariableJson := c.PostForm("VariableJson")

	if err := TaskPass(TaskID, Comment, VariableJson, true); err == nil {
		c.JSON(200, "ok")
	} else {
		c.JSON(400, err.Error())
	}
}

// @Summary      任务驳回
// @Description
// @Tags         任务
// @Produce      json
// @Param        TaskID  formData int  true  "任务ID" example(1)
// @Param        Comment  formData string  false  "评论意见" example("不同意")
// @Param        VariablesJson  formData string  false  "变量(Json)" example("{"User":"001"}")
// @Success      200  {object}  string "ok"
// @Failure      400  {object}  string 报错信息
// @Router       /task/reject [post]
func Task_Reject(c *gin.Context) {
	TaskID, err := strconv.Atoi(c.PostForm("TaskID"))
	if err != nil {
		c.JSON(400, err.Error())
		return
	}
	Comment := c.PostForm("Comment")
	VariableJson := c.PostForm("VariableJson")

	if err := TaskReject(TaskID, Comment, VariableJson); err == nil {
		c.JSON(200, "ok")
	} else {
		c.JSON(400, err.Error())
	}
}

// @Summary      将任务转交给他人处理
// @Description
// @Tags         任务
// @Produce      json
// @Param        TaskID  formData int  true  "任务ID" example(1)
// @Param        Users  formData string  true  "用户,多个用户使用逗号分隔" example("U001,U002,U003")
// @Success      200  {object}  string "ok"
// @Failure      400  {object}  string 报错信息
// @Router       /task/transfer [post]
func Task_Transfer(c *gin.Context) {
	TaskID, err := strconv.Atoi(c.PostForm("TaskID"))
	if err != nil {
		c.JSON(400, err.Error())
		return
	}
	Users := strings.Split(c.PostForm("Users"), ",")

	if err := TaskTransfer(TaskID, Users); err == nil {
		c.JSON(200, "ok")
	} else {
		c.JSON(400, err.Error())
	}
}

// @Summary      获取待办任务
// @Description  返回的是任务数组
// @Tags         任务
// @Produce      json
// @Param        userid  query string  true  "用户ID 传入空则获取所有用户的待办任务" example("U001")
// @Param        procname  query string  false  "指定流程名称，非必填" example("请假")
// @Param        asc  query bool  true  "是否按照任务生成时间升序排列" example(true)
// @Param        idx  query int  true  "分页用,开始index" example(0)
// @Param        rows  query int  true  "分页用,最大返回行数" example(0)
// @Success      200  {object}  []model.Task 任务数组
// @Failure      400  {object}  string 报错信息
// @Router       /task/todo [get]
func Task_ToDoList(c *gin.Context) {
	UserID := c.Query("userid")
	ProcessName := c.Query("procname")
	SortByAsc, err := strconv.ParseBool(c.Query("asc"))
	if err != nil {
		c.JSON(400, err.Error())
		return
	}
	StartIndex, err := strconv.Atoi(c.Query("idx"))
	if err != nil {
		c.JSON(400, err.Error())
		return
	}
	MaxRow, err := strconv.Atoi(c.Query("rows"))
	if err != nil {
		c.JSON(400, err.Error())
		return
	}

	if tasks, err := GetTaskToDoList(UserID, ProcessName, SortByAsc, StartIndex, MaxRow); err == nil {
		c.JSON(200, tasks)
	} else {
		c.JSON(400, err.Error())
	}
}

// @Summary      获取已办任务
// @Description  返回的是任务数组
// @Tags         任务
// @Produce      json
// @Param        userid  query string  true  "用户ID 传入空则获取所有用户的已完成任务(此时IgnoreStartByMe参数强制为False)" example("U001")
// @Param        procname  query string  false  "指定流程名称，非必填" example("请假")
// @Param        ignorestartbyme  query bool  true  "忽略由我开启流程,而生成处理人是我自己的任务" example("true")
// @Param        asc  query bool  true  "是否按照任务完成时间升序排列" example(true)
// @Param        idx  query int  true  "分页用,开始index" example(0)
// @Param        rows  query int  true  "分页用,最大返回行数" example(0)
// @Success      200  {object}  []model.Task 任务数组
// @Failure      400  {object}  string 报错信息
// @Router       /task/finished [get]
func Task_FinishedList(c *gin.Context) {
	UserID := c.Query("userid")
	ProcessName := c.Query("procname")
	IgnoreStartByMe, err := strconv.ParseBool(c.Query("ignorestartbyme"))
	if err != nil {
		c.JSON(400, err.Error())
		return
	}
	SortByAsc, err := strconv.ParseBool(c.Query("asc"))
	if err != nil {
		c.JSON(400, err.Error())
		return
	}
	StartIndex, err := strconv.Atoi(c.Query("idx"))
	if err != nil {
		c.JSON(400, err.Error())
		return
	}
	MaxRow, err := strconv.Atoi(c.Query("rows"))
	if err != nil {
		c.JSON(400, err.Error())
		return
	}
	if tasks, err := GetTaskFinishedList(UserID, ProcessName, IgnoreStartByMe,SortByAsc, StartIndex, MaxRow); err == nil {
		c.JSON(200, tasks)
	} else {
		c.JSON(400, err.Error())
	}
}

// @Summary      获取本任务所在节点的所有上游节点
// @Description  此功能为自由驳回使用
// @Tags         任务
// @Produce      json
// @Param        taskid  query int  true  "任务ID" example("8")
// @Success      200  {object}  []model.Node 节点任务数组
// @Failure      400  {object}  string 报错信息
// @Router       /task/upstream [get]
func Task_UpstreamNodeList(c *gin.Context) {
	TaskID, err := strconv.Atoi(c.Query("taskid"))
	if err != nil {
		c.JSON(400, err.Error())
		return
	}

	if nodes, err := TaskUpstreamNodeList(TaskID); err == nil {
		c.JSON(200, nodes)
	} else {
		c.JSON(400, err.Error())
	}
}

// @Summary      自由任务驳回
// @Description  驳回到上游任意一个节点
// @Tags         任务
// @Produce      json
// @Param        TaskID  formData int  true  "任务ID" example(1)
// @Param        Comment  formData string  false  "评论意见" example("不同意")
// @Param        VariablesJson  formData string  false  "变量(Json)" example("{"User":"001"}")
// @Param        RejectToNodeID  formData string  false  "驳回到哪个节点" example("流程开始节点")
// @Success      200  {object}  string "ok"
// @Failure      400  {object}  string 报错信息
// @Router       /task/reject/free [post]
func Task_FreeRejectToUpstreamNode(c *gin.Context) {
	TaskID, err := strconv.Atoi(c.PostForm("TaskID"))
	if err != nil {
		c.JSON(400, err.Error())
		return
	}

	Comment := c.PostForm("Comment")
	VariableJson := c.PostForm("VariableJson")
	RejectToNodeID := c.PostForm("RejectToNodeID")

	if err := TaskFreeRejectToUpstreamNode(TaskID, RejectToNodeID, Comment, VariableJson); err == nil {
		c.JSON(200, "ok")
	} else {
		c.JSON(400, err.Error())
	}
}

// @Summary      当前任务可以执行哪些操作
// @Description  前端无法提前知道当前任务可以做哪些操作，此方法目的是解决这个困扰
// @Tags         任务
// @Produce      json
// @Param        taskid  query int  true  "任务ID" example(1)
// @Success      200  {object}  model.TaskAction "可执行任务"
// @Failure      400  {object}  string 报错信息
// @Router       /task/action [get]
func Task_WhatCanIDo(c *gin.Context) {
	TaskID, err := strconv.Atoi(c.Query("taskid"))
	if err != nil {
		c.JSON(400, err.Error())
		return
	}
	if action, err := WhatCanIDo(TaskID); err == nil {
		c.JSON(200, action)
	} else {
		c.JSON(400, err.Error())
	}
}

// @Summary      任务信息
// @Description
// @Tags         任务
// @Produce      json
// @Param        taskid  query int  true  "任务ID" example(1)
// @Success      200  {object}  model.Task "任务信息"
// @Failure      400  {object}  string 报错信息
// @Router       /task/info [get]
func Task_Info(c *gin.Context) {
	TaskID, err := strconv.Atoi(c.Query("taskid"))
	if err != nil {
		c.JSON(400, err.Error())
		return
	}
	if taskInfo, err := GetTaskInfo(TaskID); err == nil {
		c.JSON(200, taskInfo)
	} else {
		c.JSON(400, err.Error())
	}
}

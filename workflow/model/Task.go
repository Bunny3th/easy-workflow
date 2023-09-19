package model

import . "github.com/Bunny3th/easy-workflow/workflow/database"

type Task struct {
	TaskID             int       `gorm:"column:id"`                    //任务ID
	BusinessID         string    `gorm:"column:business_id"`           //业务ID
	Starter            string    `gorm:"column:starter"`               //流程起始人
	ProcID             int       `gorm:"column:proc_id"`               //流程ID
	ProcName           string    `gorm:"column:name"`                  //流程名称
	ProcInstID         int       `gorm:"column:proc_inst_id"`          //流程实例ID
	NodeID             string    `gorm:"column:node_id"`               //节点ID
	NodeName           string    `gorm:"column:node_name"`             //节点名称
	PrevNodeID         string    `gorm:"column:prev_node_id"`          //上一节点ID
	IsCosigned         int       `gorm:"column:is_cosigned"`           //0:任意一人通过即可 1:会签
	BatchCode          string    `gorm:"column:batch_code"`            //批次码.节点会被驳回，一个节点可能产生多批task,用此码做分别
	UserID             string    `gorm:"column:user_id"`               //分配用户ID
	Status             int       `gorm:"column:status"`                //任务状态:0:初始 1:通过 2:驳回
	IsFinished         int       `gorm:"column:is_finished"`           // 0:任务未完成 1:处理完成
	Comment            string    `gorm:"column:comment"`               //评论意见
	ProcInstCreateTime *LocalTime `gorm:"column:proc_inst_create_time;"` //流程实例创建时间
	CreateTime         *LocalTime `gorm:"column:create_time;"`           //任务创建时间
	FinishedTime       *LocalTime `gorm:"column:finished_time;"`         // 处理任务时间
}

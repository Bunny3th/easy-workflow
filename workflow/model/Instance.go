package model

import . "github.com/Bunny3th/easy-workflow/workflow/database"

type Instance struct {
	ProcInstID    int       `gorm:"column:id;"`             //流程实例ID
	ProcID        int       `gorm:"column:proc_id"`         //流程ID
	ProcName      string    `gorm:"column:name"`            //流程名称
	ProcVersion   int       `gorm:"column:proc_version"`    //流程版本号
	BusinessID    string    `gorm:"column:business_id"`     //业务ID
	Starter       string    `gorm:"column:starter"`         //流程发起人用户ID
	CurrentNodeID string    `gorm:"column:current_node_id"` //当前进行节点ID
	CreateTime    LocalTime `gorm:"column:create_time"`     //创建时间
	Status        int       `gorm:"column:status"`          //0:未完成(审批中) 1:已完成(通过) 2:撤销
}

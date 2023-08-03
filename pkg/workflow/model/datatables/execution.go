package datatables


type Execution struct {
	ProcID      int    `gorm:"column:proc_id"`      //流程ID
	ProcVersion int    `gorm:"column:proc_version"` //流程版本号
	NodeID      string `gorm:"column:node_id"`      //节点ID
	NodeName    string `gorm:"column:node_name"`    //节点名称
	PrevNodeID  string `gorm:"column:prev_node_id"` //上级节点ID
	NodeType    int    `gorm:"column:node_type"`    //流程类型 0:根节点 1:任务节点 2:网关节点 3:结束节点
	Gateway     string `gorm:"column:gateway"`      //网关定义(只有在nodetype为2时才会有)
	IsCosigned  int    `gorm:"column:is_cosigned"`  //是否会签
	PreEvents   string `gorm:"column:pre_events"`   //前置事件
	ExitEvents  string `gorm:"column:exit_events"`  //退出事件
	CreateTime  string `gorm:"column:create_time"`  //创建时间
}





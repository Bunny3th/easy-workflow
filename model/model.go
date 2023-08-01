package modelfff

type NodeType int

const (
	Root    NodeType = 0 //根节点
	General NodeType = 1 //通常节点
	GateWay NodeType = 2 //参考activiti的网关.目前只实现了排他网关
	End     NodeType = 3 //结束节点

)





type Node struct {
	NodeID      string   //节点名称
	NodeName    string   //节点名字
	NodeType    NodeType //节点类型
	PrevNodeID  string   //上级节点(任何节点只能有一个上级节点)
	NextNodeIDs []string //下级节点(节点可以有N个直接下级节点)  是否需要下级节点？感觉可以不用。需要，因为生成task时候需要知道下级是谁，或者可以通过计算得知
	UserIDs     []string //节点处理人数组
	//Role        []string //节点处理角色数组。注意，一旦使用角色，则该节点默认不能会签。因为系统无法预先知道角色中存在多少用户。除非通过事件修改.暂时不用
	GateWays    []GW     //网关。只有在节点类型为GateWay的情况下此字段才会有值
	Comment     string   //备注
	Cosigned    bool     //是否会签  会签的情况下，需要所有人通过才能进行下一节点，只要有一人反对，则退回上一节点
	PreEvent    []string //前置事件
	ExitEvent   []string //退出事件
	//Variables   map[string]string //传入的变量  这个不应该在node中定义，应该是在传入参数中获取
}

//run time variable
type Variables struct {
	BusinessID string //关联业务ID
	NodeUserMapping map[string][]string
}


//节点数组可以存在全局map中，key就是proc_id_version
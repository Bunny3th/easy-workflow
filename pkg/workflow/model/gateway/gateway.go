package gateway

//目前只实现排他网关
type ExclusiveGateway struct {
	Conditions []Condition
}

//条件
type Condition struct {
	Expression string   //条件表达式
	NodeID     string   //满足条件后转跳到哪个节点
}


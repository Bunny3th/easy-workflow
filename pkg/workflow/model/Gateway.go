package model

//混合网关,等于activiti中排他、并行网关、包含网关的混合体
type HybridGateway struct {
	Conditions         []Condition //条件判断节点
	InevitableNodes    []string    //必然执行的节点
	WaitForAllPrevNode int         //0:等于包含网关，只要上级节点有一个完成，就可以往下走   1:等于并行网关，必须要上级节点全部完成才能往下走
}

//条件
type Condition struct {
	Expression string //条件表达式
	NodeID     string //满足条件后转跳到哪个节点
}

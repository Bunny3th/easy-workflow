package gateway

//目前只实现排他网关
type Gateway struct {
	Key               string
	RegularExpression string
	NodeIDs           []string
	
}

type Condition struct {
	Keys       []string //变量名
	Expression string   //条件表达式
	NodeID     string   //满足条件后转跳到哪个节点
}

func (g *Gateway) NextNodeIDs() (NodeIDs []string) {
	return g.NodeIDs
}

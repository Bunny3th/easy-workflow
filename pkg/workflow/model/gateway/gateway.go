package gateway


//目前只实现排他网关
type Gateway struct {
	Key               string
	RegularExpression string
	NodeIDs           []string
}

func (g *Gateway) NextNodeIDs() (NodeIDs []string) {
	return g.NodeIDs
}

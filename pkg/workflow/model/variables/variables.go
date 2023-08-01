package variables

type Variables struct {
	BusinessID      string //关联业务ID
	NodeUserMapping map[string][]string
	GateWayMapping  map[string][]string
}

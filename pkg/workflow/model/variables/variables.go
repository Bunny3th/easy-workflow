package variables

type Variables struct {
	BusinessIDMapping map[string]string //关联业务ID
	NodeUserMapping   map[string][]string
	GateWayMapping    map[string][]string
}

package model

type Process struct {
	ProcessName  string   //流程名
	Source       string   //来源(引擎可能被多个系统、组件等使用，这里记下从哪个来源创建的流程
	RevokeEvents []string //流程撤销事件.在流程实例撤销时触发
	Nodes        []Node   //节点
}

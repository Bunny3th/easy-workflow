package event

import (
	. "github.com/Bunny3th/easy-workflow/workflow/engine"
	. "github.com/Bunny3th/easy-workflow/workflow/model"
	"log"
)

//这里创建了一个角色-用户的人员库，用来模拟数据库中存储的角色-用户对应关系
var RoleUser = make(map[string][]string)

func init() {
	//初始化人事数据
	RoleUser["主管"] = []string{"张经理"}
	RoleUser["人事经理"] = []string{"人事老刘"}
	RoleUser["老板"] = []string{"李老板", "老板娘"}
	RoleUser["副总"] = []string{"赵总", "钱总", "孙总"}
}

//示例事件
type MyEvent struct{}

func (e *MyEvent) MyEvent_End(ProcessInstanceID int, CurrentNode *Node, PrevNode Node) error {
	//可以做一些处理，比如通知流程开始人，节点到了哪个步骤
	processName, err := GetProcessNameByInstanceID(ProcessInstanceID)
	if err != nil {
		return err
	}
	log.Printf("--------流程[%s]节点[%s]结束-------", processName, CurrentNode.NodeName)
	return nil
}

func (e *MyEvent) MyEvent_Notify(ProcessInstanceID int, CurrentNode *Node, PrevNode Node) error {
	processName, err := GetProcessNameByInstanceID(ProcessInstanceID)
	if err != nil {
		return err
	}
	log.Printf("--------流程[%s]节点[%s]，通知节点中对应人员--------", processName, CurrentNode.NodeName)
	if CurrentNode.NodeType == EndNode {
		log.Printf("============================== 流程[%s]结束 ==============================", processName)
		variables, err := ResolveVariables(ProcessInstanceID, []string{"$starter"})
		if err != nil {
			return err
		}
		log.Printf("通知流程创建人%s,流程[%s]已完成", variables["$starter"], processName)

	} else {
		for _, user := range CurrentNode.UserIDs {
			log.Printf("通知用户[%s],抓紧去处理", user)
		}
	}
	return nil
}

//解析角色
func (e *MyEvent) MyEvent_ResolveRoles(ProcessInstanceID int, CurrentNode *Node, PrevNode Node) error {
	processName, err := GetProcessNameByInstanceID(ProcessInstanceID)
	if err != nil {
		return err
	}
	log.Printf("--------流程[%s]节点[%s],开始解析角色--------", processName, CurrentNode.NodeName)
	//把用户库中对应角色的用户全部放到CurrentNode.UserIDs中去
	for _, role := range CurrentNode.Roles {
		if users, ok := RoleUser[role]; ok {
			CurrentNode.UserIDs = append(CurrentNode.UserIDs, users...)
		}
	}
	return nil
}

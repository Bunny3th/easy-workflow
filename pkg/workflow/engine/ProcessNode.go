package engine

import (
	. "easy-workflow/pkg/workflow/model/node"
	"errors"
	"github.com/pywee/lit"
	"log"
	"regexp"
)

//处理节点,如：生成task、进行条件判断、处理结束节点等，返回下一个节点
func ProcessNode(ProcessInstanceID int, node Node) Node {
	//这里处理前置事件
	//do something

	//如果处理的是开始节点
	//思考，这里不需要处理开始节点，因为开始节点是特殊的，一个流程只有一个开始节点
	//if node.NodeType == Root {
	//
	//	//生成一条task
	//
	//	//查找下级,并递归运行ProcessNode进行处理
	//
	//}

	if node.NodeType == GateWay {
		//表达式分析,找到下一个节点

		//进行处理
	}

	if node.NodeType == Task {

	}

	if node.NodeType == End {

	}

	exprs := []byte("a=100>90;print(a)")
	_, err := lit.NewExpr(exprs)
	if err != nil {
		log.Println(err)
	}

	//匹配字母数字以及下划线
	reg := regexp.MustCompile(`[$]\w+`)
	match := reg.FindAllString("$Ssb123=$123", -1)

	//match,_:=regexp.Match("^[$][A-Za-z0-9]+$",[]byte("$abd01=c"))
	log.Printf("正则表达式:%+v", match)

	//这里处理退出事件
	//do something

	return node
}

//开始节点处理
func StartNodeHandle(ProcessInstanceID int, node Node, Comment string,VariableJson string) (Node, error) {
	if node.NodeType != Root {
		return node, errors.New("不是开始节点，无法处理节点:" + node.NodeName)
	}

	//开始节点只需要一个处理人
	//思考：如果开始节点有多个处理人，则可能进入会签状态，可能造成流程都无法正常开始
	//所以，开始节点只能有一个处理人，且默认就是非会签节点
	if len(node.UserIDs) < 1 {
		return node, errors.New("未指定处理人，无法处理节点:" + node.NodeName)
	}

	//匹配节点用户变量
	User := node.UserIDs[0]
	if IsVariable(User) {
		value, ok, err := SetVariable(ProcessInstanceID, User)
		if err != nil {
			return node, err
		}
		if !ok {
			return node, errors.New("无法匹配变量:" + User)
		}
		User=value
	}

	//生成一条Task
	taskids,err :=CreateTask(ProcessInstanceID, node.NodeID,"", []string{User})
	if err != nil {
		return node, err
	}

	//完成task,并获取下一步NodeID
	NextNodeID,err:=TaskPass(taskids[0],Comment,VariableJson)
	if err != nil {
		return node, err
	}

	log.Println("下一个处理节点:",NextNodeID)

	//判断节点是否可结束


	//执行下一个节点




	return node, nil

}

//思考：应该在task pass 或者 reject的时候，就马上判断下一步是什么

//判断任务节点是否可流转,若已结束，则返回该task所在节点的下一个节点(通过)或上一节点(驳回)


package engine

import (
	."easy-workflow/pkg/workflow/model/node"
	"github.com/pywee/lit"
	"log"
	"regexp"
)

//处理节点,如：生成task、进行条件判断、处理结束节点等，返回下一个节点
func ProcessNode(ProcessInstanceID int, node Node) Node {
	//这里处理前置事件
	//do something

	//如果处理的是开始节点
	if node.NodeType==Root{
		//生成一条task，状态为通过的task

		//查找下级,并进行处理

	}

	if node.NodeType==GateWay{
		//表达式分析,找到下一个节点



		//进行处理
	}

	exprs:=[]byte("a=100>90;print(a)")
	_, err := lit.NewExpr(exprs)
	if err!=nil{
		log.Println(err)
	}

	reg:=regexp.MustCompile(`^\w[^=]+$`)
	match:=reg.FindAllString("Ssb123=123",-1)

	//match,_:=regexp.Match("^[$][A-Za-z0-9]+$",[]byte("$abd01=c"))
	log.Printf("正则表达式:%+v",match)


	//这里处理退出事件
	//do something

	return node
}

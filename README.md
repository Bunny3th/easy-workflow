# easy-workflow
已完成大量内测工作,正式版已发布

## easy-workflow是什么  
这是一个用纯GO语言开发的简单易用工作流引擎,可以集成到GO项目中，也可以单独作为Web Api Server运行

## 它有什么功能
除了基本的流程处理，作者经过中国式流程设计的洗礼，所以它包含以下增强功能：  
1、支持自定义事件。目前支持4种事件，分别是:a、节点开始 b、节点结束 c、任务结束 d、流程撤销;    
2、支持会签。所谓会签，即节点由多人审核，全部通过才算通过，任意一人驳回即算驳回;   
3、混合网关。约等于activiti中排他、并行网关、包含网关的混合体，更好支持复杂流程;       
4、支持各种飞线跳转审批:如自由驳回功能，可以任意驳回到上游任何节点;也直接提交到上一个驳回我的节点;  
5、可自定义计划任务。  
以上种种，只为完成更多中国式需求。   
### 用一个故事讲述功能点  
故事背景:假设某流程包含A、B、C、D、E 共5个任务节点，E节点由老板审核。而老板比较任性(向来如此)。    
#### 1、自由驳回
正常情况下,老板要把流程打回到创建人A的方式应该是:驳回到D，D再驳回到C...最终打回到流程提交人A。    
但是老板有钱任性，想要直接打回到A。此时，可以使用“自由驳回”功能，满足老板直接驳回，不留情面的需求。     
#### 2、直接提交到上一个驳回我的节点
A被领导E直接驳回，于是按照领导指示做修改后重新提交，传统情况下。流程需要重新流过B、C、D几个主管。    
一次两次还好，奈何老板各种不满意各种驳回，每次都要主管们再审核一遍。这样不仅效率低，还影响主管领导对员工的情绪。      
老板发话：芝麻绿豆大的事，B、C、D不用再参合了，小A你改完直接提我这边吧！  
此时使用"直接提交到上一个驳回我的节点"，A直接提交到上次驳回他的E，皆大欢喜。  
#### 3、混合网关  
国内java开发者常用的activiti引擎中包含以下几种网关:    
1、ParallelGeteway:并行网关，网关中所有节点完成才可流向网关下一个节点;  
2、ExclusiveGateway:排他网关，多用作条件判断，比如满足条件A，则流向节点1；满足条件B，则流向节点2，只能两者选一;   
3、InclusiveGateway:包含网关，并行与排他的混合。  
**网关的本质是控制节点的走向、流转逻辑，从而实现更复杂的流程定义。**   
作者不喜欢把事情搞复杂，所以本项目中只有一种网关:混合网关,可视为activiti中排他、并行网关、包含网关的混合体。  
**简言之，这个流程引擎考虑国情，力求更好的满足多样化需求**
#### 4、事件  
事件可以用作通知、改变流程行为。举几个例子：  
a、节点开始事件可以用于解析节点中角色。流程引擎并不知道节点中“主管”角色到底是谁，此时必须使用开始事件从业务库中解析“主管”角色的用户ID;    
b、节点事件可用于通知，提醒各任务处理人;  
c、任务完成事件与节点事件的区别在于：一个节点可能有多个任务，节点开始与结束事件在节点生命周期中只会各运行一次。而任务事件在每个任务提交后都会触发。    
d、流程被撤销时也可触发事件。  
在示例代码中，使用任务完成事件改变了会签节点的行为，使原本需要3人全部通过的节点变成只要2人通过。   
#### 5、任务计划  
任务计划可以看作是事件的辅助，示例中演示了如何利用计划任务让特定用户的任务自动完成。     

## 开始使用  
Tips:作者强烈建议阅读example中代码示例。注释比代码多一向是作者的风格，基本看完example，即可知引擎使用方法。    

### 环境要求  
1、MySQL 8.0以上版本(支持CTE)   
2、需要1.18以上Go版本(支持泛型)  

### 下载  
方法一:  
在go.mod文件中添加  github.com/Bunny3th/easy-workflow 版本号    

方法二:    
go get github.com/Bunny3th/easy-workflow@版本号  

### 开启引擎  
```go
import "github.com/Bunny3th/easy-workflow/workflow/engine"

func DBConnConfig() {
	engine.DBConnConfigurator.DBConnectString = "数据库账号:密码@tcp(地址:端口)/数据库名称?charset=utf8mb4&parseTime=True&loc=Local"	
}

func main() {
   //开启工作流引擎
   engine.StartWorkFlow(DBConnConfig, false, nil)
}
```
StartWorkFlow方法：  
```go
func StartWorkFlow(DBConnConfigurator DataBaseConfigurator, ignoreEventError bool, EventStructs ...any)
```
传入参数定义
  + DBConnConfigurator:数据库连接配置器,完整配置func如下:
```go
  func DBConnConfig() {
     DBConnConfigurator.DBConnectString = "连接字符串"        //必须设置         数据库连接字符串
     DBConnConfigurator.MaxIdleConns=100                     //非必设,默认10    空闲连接池中连接的最大数量
     DBConnConfigurator.MaxOpenConns=200                     //非必设,默认100   打开数据库连接的最大数量
     DBConnConfigurator.ConnMaxLifetime=200                  //非必设,默认3600  连接可复用的最大时间（分钟）
     DBConnConfigurator.SlowThreshold=3                      //非必设,默认1     慢SQL阈值(秒)
     DBConnConfigurator.LogLevel=4                           //非必设,默认3     日志级别 1:Silent  2:Error 3:Warn 4:Info
     DBConnConfigurator.IgnoreRecordNotFoundError=false      //非必设,默认true  忽略ErrRecordNotFound（记录未找到）错误
     DBConnConfigurator.Colorful=false                       //非必设,默认true  使用彩色打印
  }
```
  + ignoreEventError：在事件执行时，是否忽略其报错。事件出错可能导致流程无法运行,此选项设置为true，则忽略事件出错，让流程继续  
  + EventStructs：作者使用反射运行事件方法，故需将事件方法“挂”在Struct上传入。若流程定义中无需运行事件，则直接传nil即可。事件代码示例:  
```go
//示例事件
type MyEvent struct{}

//节点结束事件
func (e *MyEvent) MyEvent_End(ProcessInstanceID int, CurrentNode *Node, PrevNode Node) error {
	//示例:在节点结束时打印信息
	processName, err := GetProcessNameByInstanceID(ProcessInstanceID)
	if err != nil {
		return err
	}
	log.Printf("--------流程[%s]节点[%s]结束-------", processName, CurrentNode.NodeName)
	return nil
}
```  
**注意事项**  
1、事件方法接收者必须是指针。如上示方法MyEvent_End,其方法接收者为*MyEvent  
2、StartWorkFlow传入事件Struct时，必须传入指针，如：StartWorkFlow(DBConnConfig,false,&MyEvent{})  
### 更多信息
**请查阅源码根目录下"EasyWorkFlow工作流引擎说明文档.pdf"文档**  

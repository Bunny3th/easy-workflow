package engine

import (
	"errors"
	"fmt"
	"time"
)

//计划任务池
//任务计划池的作用是存储任务运行中的信息，方便排查问题
var ScheduledTaskPool = make(map[string]*ScheduledTask)

//被计划的任务
type ScheduledTask struct {
	StartAt        time.Time     //任务开始时间
	StopAt         time.Time     //任务结束时间
	IntervalSecond int64         //重复执行间隔(秒),最小1秒
	Func           func() error  //需要运行的任务方法
	LastRunTime    time.Time     //上一次运行时间
	LastResult     string        //上一次运行结果
	LastDuration   time.Duration //上一次运行时长
}

//获取任务信息
func GetScheduledTaskList() map[string]*ScheduledTask {
	return ScheduledTaskPool
}

//登记计划任务，任务会被添加进计划任务池，并进入运行状态，可使用GetScheduledTaskList查看任务运行信息
//**注意**:请使用go关键词运行此函数。因为任务运行周期可能很长，若不使用goroutine运行,则可能造成主进程堵塞
//参数说明:
//StartAt 任务开始时间
//StopAt  任务结束时间
//IntervalSecond 重复执行间隔(秒),最小1秒
//Func 需要执行的方法,签名必须是func() error
//
func ScheduleTask(TaskName string, StartAt time.Time, StopAt time.Time, IntervalSecond int64, Func func() error) error {
	if _, ok := ScheduledTaskPool[TaskName]; ok {
		return errors.New("此任务已被加入任务池，无需重复操作")
	}

	//当前时间
	now := time.Now()

	//如果当前时间已经在StopAt之后，则不必执行
	if now.After(StopAt) {
		return errors.New("任务结束时间小于当前时间，任务不会被运行")
	}

	//间隔小于1秒会出错
	if IntervalSecond < 1 {
		return errors.New("重复执行间隔最小1秒")
	}

	//结束时间应在开始时间之后
	if StopAt.Before(StartAt) {
		return errors.New("开始时间应小于结束时间")
	}

	ScheduledTaskPool[TaskName] = &ScheduledTask{
		StartAt:        StartAt,
		StopAt:         StopAt,
		IntervalSecond: IntervalSecond,
		Func:           Func,
	}

	//开始运行任务
	runScheduledTask(TaskName)

	return nil
}

//运行任务计划
func runScheduledTask(TaskName string) {
	//如果计划任务池中没有，则直接退出
	task, ok := ScheduledTaskPool[TaskName]
	if !ok {
		return
	}

	//当前时间
	now := time.Now()

	//如果当前时间已经在StopAt之后，则不必执行
	if now.After(task.StopAt) {
		return
	}

	//等待多久后开始执行任务
	var waitDuration time.Duration

	//如果当前时间已经在StartAt之后，定时器马上开启
	if now.After(task.StartAt) {
		waitDuration = 0
	} else { //否则，设置等待时间
		waitDuration = task.StartAt.Sub(now)
	}

	//定义一个定时器
	timer := time.NewTimer(waitDuration)
	defer timer.Stop()

	//堵塞直到定时器时间到达
	<-timer.C

	//定义一个周期性定时器
	ticker := time.NewTicker(time.Duration(task.IntervalSecond) * time.Second)
	defer ticker.Stop()

	//防止panic导致程序退出
	defer func() {
		if err := recover(); err != nil {
			//记录运行状态
			task.LastResult = fmt.Sprint(err)
			//记录运行耗时
			task.LastDuration = time.Now().Sub(task.LastRunTime)
		}
	}()

	for {
		<-ticker.C
		//如果已经不在计划任务池中，则不需要再运行。
		_, ok := ScheduledTaskPool[TaskName]
		if !ok {
			return
		}

		//如果当前已经在StopAt之后，则退出
		if time.Now().After(task.StopAt) {
			return
		}

		//记录运行时间
		task.LastRunTime = time.Now()

		//运行任务 这里不再使用协程。
		//考虑一个情况:任务间隔2秒，但是A协程运行时间长达10秒，2秒后B又开始跑，这样前面还没结束后面就开始，会造成大量的性能消耗与不可控情况
		err := task.Func()
		//记录运行状态
		if err == nil {
			task.LastResult = "ok"
		} else {
			task.LastResult = err.Error()
		}
		//记录运行耗时
		task.LastDuration = time.Now().Sub(task.LastRunTime)
	}
}

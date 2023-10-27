package engine

import (
	"log"
	"time"
)

//记录任务运行状态
var ScheduleRecorder map[string]string



//执行任务计划 参数说明:
//StartAt 任务开始时间
//StopAt  任务结束时间
//IntervalSecond 重复执行间隔(秒)
//Func 需要执行的方法
func RegisterScheduleTask(StartAt time.Time, StopAt time.Time, IntervalSecond int64, Func func()) {
	//当前时间
	now := time.Now()

	//如果当前时间已经在StopAt之后，则不必执行
	if now.After(StopAt) {
		return
	}

	//间隔不能小于1秒
	if IntervalSecond < 1 {
		return
	}

	//结束时间应在开始时间之后
	if StopAt.Before(StartAt) {
		return
	}

	//等待多久后开始执行任务
	var waitDuration time.Duration

	//如果当前时间已经在StartAt之后，定时器马上开启
	if now.After(StartAt) {
		waitDuration = 0
	} else { //否则，设置等待时间
		waitDuration = StartAt.Sub(now)
	}

	//定义一个定时器
	timer := time.NewTimer(waitDuration)
	defer timer.Stop()

	//堵塞直到定时器时间到达
	<-timer.C

	//定义一个周期性定时器
	ticker := time.NewTicker(time.Duration(IntervalSecond) * time.Second)
	defer ticker.Stop()

	for {
		<-ticker.C
		//如果当前已经在StopAt之后，则退出
		if time.Now().After(StopAt) {
			return
		}

		//运行任务
		go func() {
			//防止panic导致程序退出。任务不应为几次panic而停止运行
			defer func() {
				if err := recover(); err != nil {
					log.Println("计划任务执行中出现异常:", err)
				}else{
					ScheduleRecorder[""]=""
				}
			}()

			//运行任务
			Func()
		}()
	}
}

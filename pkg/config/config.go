package config

import (
	"github.com/spf13/viper"
	"log"
	"reflect"
)

var cfg *viper.Viper

//这里定义一个空结构体，目的是使用反射执行该结构体下的func
type Initializer struct{}

//定义包导入初始化方法
func init() {
	cfg = viper.New()
	cfg.AddConfigPath(".") // 在哪个目录中寻找配置文件 “.” 为程序同一文件夹
	cfg.SetConfigName("config")
	cfg.SetConfigType("yaml")

	if err := cfg.ReadInConfig(); err != nil {
		log.Fatalln(err)
	}

	//这里使用反射，执行所有的配置导入函数
	//注意：要求对应的导入函数名必须与config文件中对应key名字一致
	InitializerValue := reflect.ValueOf(&Initializer{})
	InitializerType := InitializerValue.Type()
	for i := 0; i < InitializerType.NumMethod(); i++ {
			log.Println("start 加载配置项:", InitializerType.Method(i).Name)
			if existKey(InitializerType.Method(i).Name){
				var args = []reflect.Value{
					reflect.ValueOf(&Initializer{}), // 这里要加这个，否则报错
				}
				InitializerType.Method(i).Func.Call(args)
			}
	}
}

//检查配置文件中key是否存在
func existKey(key string) bool{
	if cfg.Get(key) == nil {
		log.Printf("warning:配置文件中无%s对应项\n", key)
		return false
	}
	return true
}

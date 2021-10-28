package main

import (
	"flag"
	"fmt"
	"time"

	"api_metrics/config"
	"api_metrics/module"

	log "github.com/sirupsen/logrus"
)

//多个命令行参数使用 接口定义
type Value interface {
	String() string
	Set(string) error
}

type arrayFlags []string

func (i *arrayFlags) String() string {
	return fmt.Sprint(*i)
}

func (i *arrayFlags) Set(value string) error {
	*i = append(*i, value)
	return nil
}

func main() {
	// 定义命令行参数信息
	var modules arrayFlags
	flag.Var(&modules, "module", "Some config metrics module name")
	flag.Parse()

	// 判断传入的是否为空，否则抛出异常
	if modules == nil {
		log.Fatal("Please enter the parameters [module] !")
	}

	// 定时任务执行
	t := time.Duration(config.ReadConfig("timed_task.second").(int))
	ticker := time.NewTicker(t * time.Second) // 定时任务执行间隔，config.yaml里配置
	defer ticker.Stop()                       // 执行完启动阻塞
	for range ticker.C {
		for _, m := range modules {
			urls := config.ReadConfig("modules." + m + ".url").([]interface{})
			mothod := config.ReadConfig("modules." + m + ".mothod").(string)

			// 连接数据库
			conn := module.Conninflux()
			defer conn.Close()

			//判断mothod方法
			for _, url := range urls {
				if mothod == "GET" {
					trace := module.Get_Trace(url.(string))
					module.Writeinflux(conn, m, mothod, trace)

				} else if mothod == "POST" {
					body := config.ReadConfig("modules." + m + ".body").(string)
					trace := module.Post_Trace(url.(string), body)
					module.Writeinflux(conn, m, mothod, trace)

				} else {
					log.Error("Mothod: ", mothod, " is error, Please check yaml config !")
				}
			}
		}
	}
}

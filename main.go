package main

import (
	"flag"
	"fmt"
	"os"
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
	var modules, depends arrayFlags
	flag.Var(&modules, "module", "Some config metrics module name")
	flag.Var(&depends, "depend", "depend module")
	flag.Parse()

	// 判断传入的是否为空，否则抛出异常
	if modules == nil {
		log.Fatal("Please enter the parameters [module] !")
	}

	ticker := time.NewTicker(time.Duration(config.ReadConfig("timer.second").(int)) * time.Second)
	for range ticker.C { //定时器运行
		go func() { //goruntine 并行执行
			for _, d := range depends { //先循环依赖接口 并 获取接口值
				depend_body := config.ReadConfig("depends." + d + ".body").(string)
				params := config.ReadConfig("depends." + d + ".params").(string)
				default_value := config.ReadConfig("depends." + d + ".default_value").(string)

				for _, m := range modules { //循环基础模块值
					mothod := config.ReadConfig("modules." + m + ".mothod").(string)
					urls := config.ReadConfig("modules." + m + ".url").(map[string]interface{})
					depend := config.ReadConfig("modules." + m + ".depend")

					for name, url := range urls { //循环模块下的url
						url := url.(string)
						if mothod == "GET" { //判断mothod
							//获取trace响应信息
							trace := module.Get_Trace(url)
							//写入数据库
							conn := module.Conninflux()
							module.Writeinflux(conn, name, m, mothod, trace)
							module.Writeinflux_business(conn, name, url)
							conn.Close()

						} else if mothod == "POST" { //判断mothod
							//根据depend判断是否需要变量，不等于nil说明yaml中的modules存在depend，依赖某个接口值，为nil则不需要
							if depend != nil {
								request := module.Request_post(url, depend_body, params)
								//request获取depend依赖接口返回值，如果这个依赖接口返回成功写入变量中，失败则设置一个默认接口值，yaml中有配置
								if request != nil {
									os.Setenv(d, request.(string))
								} else {
									log.Error(depend, " request is fail !")
									os.Setenv(d, default_value)
								}
							}
							//根据yaml获取变量中的body，没有定义${}信息加载默认值
							body := os.ExpandEnv(config.ReadConfig("modules." + m + ".body").(string))
							//获取trace响应信息
							trace := module.Post_Trace(url, body)
							//写入数据库
							conn := module.Conninflux()
							module.Writeinflux(conn, name, m, mothod, trace)
							module.Writeinflux_business(conn, name, url)
							conn.Close()

						} else {
							log.Error("Mothod: ", mothod, " is error, Please check yaml config !")
						}
					}
				}
			}
		}()
	}
}

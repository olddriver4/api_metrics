package main

import (
	"flag"
	"fmt"
	"time"

	"api_metrics/config"
	"api_metrics/module"

	log "github.com/sirupsen/logrus"
)

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
	var modules arrayFlags
	flag.Var(&modules, "module", "Some config metrics module name")
	flag.Parse()
	if modules == nil {
		log.Fatal("Please enter the parameters [module] !")
	}

	t := time.Duration(config.ReadConfig("timed_task.second").(int))
	ticker := time.NewTicker(t * time.Second)
	defer ticker.Stop() // 定时任务执行
	for range ticker.C {
		for _, m := range modules {
			urls := config.ReadConfig("modules." + m + ".url").([]interface{})
			mothod := config.ReadConfig("modules." + m + ".mothod").(string)

			conn := module.Conninflux()
			defer conn.Close()

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

package module

import (
	"time"

	"github.com/go-resty/resty/v2"
	log "github.com/sirupsen/logrus"
	"github.com/thedevsaddam/gojsonq/v2"
)

func Request_post(url string, body string, params string) interface{} {
	client := resty.New() // 创建一个restry客户端
	client.SetCloseConnection(true).SetTimeout(time.Second * 10)

	resp, err := client.R().
		SetHeader("Content-Type", "application/json"). //默认请求头
		SetBody(body).                                 //匹配body
		Post(url)

	if err == nil {
		filterkey := gojsonq.New().FromString(resp.String()).Find(params) //result.hash
		return filterkey
	}

	log.Error("Client mothod error: ", err)
	return nil
}

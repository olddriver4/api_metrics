package module

import (
	"api_metrics/config"
	"strings"
	"time"

	"github.com/go-resty/resty/v2"
	client "github.com/influxdata/influxdb1-client/v2"
	log "github.com/sirupsen/logrus"
)

type Request struct {
	URL            string
	Mothod         string
	Status         int
	Proto          string
	DNSLookup      string
	ConnTime       string
	TCPConnTime    string
	TLSHandshake   string
	ServerTime     string
	ResponseTime   string
	TotalTime      string
	RequestAttempt string
}

func Post_Trace(url string, body string) Request {
	client := resty.New() // 创建一个restry客户端
	client.SetRetryWaitTime(10 * time.Second).SetRetryCount(1).SetCloseConnection(true).SetTimeout(time.Second * 10)

	resp, err := client.R().
		EnableTrace().                                 //开启trace
		SetHeader("Content-Type", "application/json"). //默认请求头
		SetBody(body).                                 //匹配body
		Post(url)
	if err != nil {
		log.Error("Client mothod error: ", err)
	}
	ti := resp.Request.TraceInfo()
	req := &Request{
		URL:          url,
		Status:       resp.StatusCode(),
		Proto:        resp.Proto(),
		DNSLookup:    ti.DNSLookup.String(),
		ConnTime:     ti.ConnTime.String(),
		TCPConnTime:  ti.TCPConnTime.String(),
		TLSHandshake: ti.TLSHandshake.String(),
		ServerTime:   ti.ServerTime.String(),
		ResponseTime: ti.ResponseTime.String(),
		TotalTime:    ti.TotalTime.String(),
	}
	return *req
}

func Get_Trace(url string) Request {
	client := resty.New() // 创建一个restry客户端
	client.SetRetryWaitTime(10 * time.Second).SetRetryCount(1).SetCloseConnection(true).SetTimeout(time.Second * 10)
	resp, err := client.R().EnableTrace().Get(url)
	if err != nil {
		log.Error("Client mothod error: ", err)
	}
	ti := resp.Request.TraceInfo()
	req := &Request{
		URL:          url,
		Status:       resp.StatusCode(),
		Proto:        resp.Proto(),
		DNSLookup:    ti.DNSLookup.String(),
		ConnTime:     ti.ConnTime.String(),
		TCPConnTime:  ti.TCPConnTime.String(),
		TLSHandshake: ti.TLSHandshake.String(),
		ServerTime:   ti.ServerTime.String(),
		ResponseTime: ti.ResponseTime.String(),
		TotalTime:    ti.TotalTime.String(),
	}

	return *req
}

func Conninflux() client.Client {
	cli, err := client.NewHTTPClient(client.HTTPConfig{
		Addr:     config.ReadConfig("influx.url").(string),      //数据库地址
		Username: config.ReadConfig("influx.user").(string),     //数据库用户名
		Password: config.ReadConfig("influx.password").(string), //数据库密码
	})
	if err != nil {
		log.Error("Error creating InfluDB Client: ", err)
	}

	return cli
}

func Writeinflux(cli client.Client, name string, public string, module string, mothod string, trace Request) {
	bp, err := client.NewBatchPoints(client.BatchPointsConfig{
		Database:  config.ReadConfig("influx.db").(string),        //数据库名称
		Precision: config.ReadConfig("influx.precision").(string), //时间精度到毫秒（很重要，不然循环写入会覆盖之前的数据，influxdb是以时间戳为单位）
	})

	if err != nil {
		log.Error("Connection influxdb fail :", err)
	}

	md := strings.ToLower(mothod)
	t := strings.ToLower(module)

	tags := map[string]string{
		"api": t,
	}
	fields := map[string]interface{}{
		"nodeid":       name,
		"url":          trace.URL,
		"publicip":     public,
		"mothod":       md,
		"proto":        trace.Proto,
		"status":       trace.Status,
		"dnslookup":    trace.DNSLookup,
		"conntime":     trace.ConnTime,
		"tcpconntime":  trace.TCPConnTime,
		"tlsHandshake": trace.TLSHandshake,
		"servertime":   trace.ServerTime,
		"responsetime": trace.ResponseTime,
		"totaltime":    trace.TotalTime,
	}
	pt, err := client.NewPoint(t, tags, fields, time.Now()) //并插入对应字段和tag，如果表不存在自动创建
	if err != nil {
		log.Error("Create table fail: ", err)
	}
	bp.AddPoint(pt)
	err = cli.Write(bp)
	if err != nil {
		log.Error("Inster fields fail: ", err)
	} else {
		requestLogger := log.WithFields(log.Fields{
			"module": module,
			"url":    trace.URL,
			"mothod": md,
		})
		requestLogger.Info("insert sucess.")
	}
}

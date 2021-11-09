package module

import (
	"api_metrics/config"
	"strings"
	"time"

	"github.com/go-resty/resty/v2"
	client "github.com/influxdata/influxdb1-client/v2"
	log "github.com/sirupsen/logrus"
	"github.com/thedevsaddam/gojsonq/v2"
)

type Request struct {
	URL            string
	Mothod         string
	Status         int
	Proto          string //协议
	DNSLookup      int64  //DNS 查询时间
	ConnTime       int64  //获取一个连接的耗时
	TCPConnTime    int64  //TCP 连接耗时
	TLSHandshake   int64  //TLS握手耗时
	ServerTime     int64  //服务器处理耗时
	ResponseTime   int64  //响应耗时
	TotalTime      int64  //总耗时
	RequestAttempt int    //请求执行次数
	RemoteAddr     string //远程服务器地址，IP:PORT格式
}

func Post_Trace(url string, body string) Request {
	client := resty.New() // 创建一个restry客户端
	client.SetCloseConnection(true).SetTimeout(time.Second * 5)

	resp, err := client.R().
		EnableTrace().                                 //开启trace
		SetHeader("Content-Type", "application/json"). //默认请求头
		SetBody(body).                                 //匹配body
		Post(url)

	status, _ := gojsonq.New().FromString(resp.String()).FindR("error.code") //获取返回值是否包含error.code，有就认为失败
	if err != nil || status != nil {
		requestLogger := log.WithFields(log.Fields{"err_conn": err, "err_status": status, "url": url, "body": body})
		requestLogger.Error("Client mothod error.")

		status := 0
		req := &Request{
			URL:          url,
			Status:       status,
			Proto:        "",
			DNSLookup:    0,
			ConnTime:     0,
			TCPConnTime:  0,
			TLSHandshake: 0,
			ServerTime:   0,
			ResponseTime: 0,
			TotalTime:    0,
			RemoteAddr:   "",
		}
		return *req
	} else {
		status := 1
		ti := resp.Request.TraceInfo()
		req := &Request{
			URL:            url,
			Status:         status,
			Proto:          resp.Proto(),
			DNSLookup:      ti.DNSLookup.Nanoseconds(),
			ConnTime:       ti.ConnTime.Nanoseconds(),
			TCPConnTime:    ti.TCPConnTime.Nanoseconds(),
			TLSHandshake:   ti.TLSHandshake.Nanoseconds(),
			ServerTime:     ti.ServerTime.Nanoseconds(),
			ResponseTime:   ti.ResponseTime.Nanoseconds(),
			TotalTime:      ti.TotalTime.Nanoseconds(),
			RequestAttempt: ti.RequestAttempt,
			RemoteAddr:     ti.RemoteAddr.String(),
		}
		requestLogger := log.WithFields(log.Fields{
			"url":  url,
			"body": body,
		})
		requestLogger.Info("request post sucess.")
		return *req
	}
}

func Get_Trace(url string) Request {
	client := resty.New() // 创建一个restry客户端
	client.SetCloseConnection(true).SetTimeout(time.Second * 5)
	resp, err := client.R().EnableTrace().Get(url)

	status, _ := gojsonq.New().FromString(resp.String()).FindR("error.code") //获取返回值是否包含error.code，有就认为失败
	if err != nil || status != nil {
		requestLogger := log.WithFields(log.Fields{"err_conn": err, "err_status": status})
		requestLogger.Error("Client mothod error.")

		status := 0
		req := &Request{
			URL:          url,
			Status:       status,
			Proto:        "",
			DNSLookup:    0,
			ConnTime:     0,
			TCPConnTime:  0,
			TLSHandshake: 0,
			ServerTime:   0,
			ResponseTime: 0,
			TotalTime:    0,
			RemoteAddr:   "",
		}
		return *req
	} else {
		status := 1
		ti := resp.Request.TraceInfo()
		req := &Request{
			URL:            url,
			Status:         status,
			Proto:          resp.Proto(),
			DNSLookup:      ti.DNSLookup.Nanoseconds(),
			ConnTime:       ti.ConnTime.Nanoseconds(),
			TCPConnTime:    ti.TCPConnTime.Nanoseconds(),
			TLSHandshake:   ti.TLSHandshake.Nanoseconds(),
			ServerTime:     ti.ServerTime.Nanoseconds(),
			ResponseTime:   ti.ResponseTime.Nanoseconds(),
			TotalTime:      ti.TotalTime.Nanoseconds(),
			RequestAttempt: ti.RequestAttempt,
			RemoteAddr:     ti.RemoteAddr.String(),
		}

		requestLogger := log.WithFields(log.Fields{
			"url": url,
		})
		requestLogger.Info("request get sucess.")
		return *req
	}
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

func Writeinflux(cli client.Client, name string, module string, mothod string, trace Request) {
	bp, err := client.NewBatchPoints(client.BatchPointsConfig{
		Database:  config.ReadConfig("influx.db").(string),        //数据库名称
		Precision: config.ReadConfig("influx.precision").(string), //时间精度（很重要，不然循环写入会覆盖之前的数据，influxdb是以时间戳为单位）
	})

	if err != nil {
		log.Error("Connection influxdb fail :", err)
	}

	md := strings.ToLower(mothod)
	t := strings.ToLower(module)

	tags := map[string]string{
		"table":  t,
		"nodeid": name,
	}
	fields := map[string]interface{}{
		"url":            trace.URL,
		"publicip":       trace.RemoteAddr,
		"mothod":         md,
		"proto":          trace.Proto,
		"status":         trace.Status,
		"dnslookup":      trace.DNSLookup,
		"conntime":       trace.ConnTime,
		"tcpconntime":    trace.TCPConnTime,
		"tlsHandshake":   trace.TLSHandshake,
		"servertime":     trace.ServerTime,
		"responsetime":   trace.ResponseTime,
		"requestAttempt": trace.RequestAttempt,
		"totaltime":      trace.TotalTime,
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

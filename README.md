## Api metrics
### HTTP module mothod {GET , POST}

#### Operating parameters:
--module 对应 config.yaml modules中定义的方法模块（可以定义多个module参数），例如：
--module baidu_get --module baidu_post

#### Yaml Config explain:
timed_task.second 超时秒数，默认30秒
influx            influx连接信息
  -precision      时间精度到毫秒（很重要，不然循环写入会覆盖之前的数据，influxdb是以时间戳为单位）默认ms
  -table          默认以mothod定义表明，比如mothod是get，表名就等于get
modules           自定义方法，配置启动参数指定需要启动那个模块
  -header         默认header 指定"Content-Type", "application/json"
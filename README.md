## Api metrics  
### HTTP module mothod {GET , POST}  

#### Operating parameters:  
>--module 对应 config.yaml modules中定义的方法模块（可以定义多个module参数）--depend 对应 config.yaml depends中定义的方法模块，例如：  
>--module baidu_get --module baidu_post  
>--depends baidu_depend  
 
#### Yaml Config explain:  
> timer:
  -second         定时器执行秒间隔   
>influx            influx连接信息     
  -precision      时间精度到毫秒（很重要，不然循环写入会覆盖之前的数据，influxdb是以时间戳为单位）默认ms    
  -table          默认以mothod定义表明，比如mothod是get，表名就等于get    
>depends           定义依赖接口，依赖接口是根据modules url提供，互相依赖  
  -params         过滤接口返回值  
  -default_value  依赖key请求不到，设置默认key   
  -body           依赖接口body  
>modules           自定义方法，配置启动参数指定需要启动那个模块(对应的表名称)   
  -mothod         mothod方法  
  -body           请求body  
  -url            定义url，url名称+url信息，名称自定义 写入数据库中，后续方便报警区分使用  
  -header         默认header 指定"Content-Type", "application/json"    
  -depend         依赖的接口名，没有就不写  

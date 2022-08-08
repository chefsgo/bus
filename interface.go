package bus

import (
	. "github.com/chefsgo/base"
	"github.com/chefsgo/chef"
	"github.com/chefsgo/util"
)

func (this *Module) Register(name string, value Any, override bool) {
	switch config := value.(type) {
	case Driver:
		this.Driver(name, config, override)
	case Config:
		this.Config(name, config, override)
	case Configs:
		this.Configs(config, override)
	case chef.Service:
		this.Service(name, config, override)
	case Filter:
		this.Filter(name, config, override)
	case Handler:
		this.Handler(name, config, override)
	}
}

func (this *Module) configure(name string, config Map) {
	cfg := Config{
		Driver: chef.DEFAULT, Weight: 1, Codec: chef.JSON,
	}
	//如果已经存在了，用现成的改写
	if vv, ok := this.configs[name]; ok {
		cfg = vv
	}

	if driver, ok := config["driver"].(string); ok {
		cfg.Driver = driver
	}

	//分配权重
	if weight, ok := config["weight"].(int); ok {
		cfg.Weight = weight
	}
	if weight, ok := config["weight"].(int64); ok {
		cfg.Weight = int(weight)
	}
	if weight, ok := config["weight"].(float64); ok {
		cfg.Weight = int(weight)
	}

	if weight, ok := config["weight"].(float64); ok {
		cfg.Weight = int(weight)
	}

	if setting, ok := config["setting"].(Map); ok {
		cfg.Setting = setting
	}

	//保存配置
	this.configs[name] = cfg
}
func (this *Module) Configure(global Map) {
	var config Map
	if vvv, ok := global["bus"].(Map); ok {
		config = vvv
	}
	if config == nil {
		return
	}

	//记录上一层的配置，如果有的话
	rootConfig := Map{}

	for key, val := range config {
		if conf, ok := val.(Map); ok {
			this.configure(key, conf)
		} else {
			rootConfig[key] = val
		}
	}

	if len(rootConfig) > 0 {
		this.configure(chef.DEFAULT, rootConfig)
	}
}
func (this *Module) Initialize() {
	if this.initialized {
		return
	}

	// 如果没有配置任何连接时，默认一个
	if len(this.configs) == 0 {
		this.configs[chef.DEFAULT] = Config{
			Driver: chef.DEFAULT, Weight: 1, Codec: chef.JSON,
		}
	} else {
		// 默认分布， 如果想不参与分布，Weight设置为小于0 即可
		for key, config := range this.configs {
			if config.Weight == 0 {
				config.Weight = 1
			}
			this.configs[key] = config
		}

	}

	//拦截器
	this.requestFilters = make([]ctxFunc, 0)
	this.executeFilters = make([]ctxFunc, 0)
	this.responseFilters = make([]ctxFunc, 0)
	for _, filter := range this.filters {
		if filter.Request != nil {
			this.requestFilters = append(this.requestFilters, filter.Request)
		}
		if filter.Execute != nil {
			this.executeFilters = append(this.executeFilters, filter.Execute)
		}
		if filter.Response != nil {
			this.responseFilters = append(this.responseFilters, filter.Response)
		}
	}

	//处理器
	this.foundHandlers = make([]ctxFunc, 0)
	this.errorHandlers = make([]ctxFunc, 0)
	this.failedHandlers = make([]ctxFunc, 0)
	this.deniedHandlers = make([]ctxFunc, 0)
	for _, filter := range this.handlers {
		if filter.Found != nil {
			this.foundHandlers = append(this.foundHandlers, filter.Found)
		}
		if filter.Error != nil {
			this.errorHandlers = append(this.errorHandlers, filter.Error)
		}
		if filter.Failed != nil {
			this.failedHandlers = append(this.failedHandlers, filter.Failed)
		}
		if filter.Denied != nil {
			this.deniedHandlers = append(this.deniedHandlers, filter.Denied)
		}
	}

	this.initialized = true
}
func (this *Module) Connect() {
	if this.connected {
		return
	}

	//记录要参与分布的连接和权重
	weights := make(map[string]int)

	for name, config := range this.configs {
		driver, ok := this.drivers[config.Driver]
		if ok == false {
			panic("Invalid bus driver: " + config.Driver)
		}

		// 建立连接
		connect, err := driver.Connect(name, config)
		if err != nil {
			panic("Failed to connect to bus: " + err.Error())
		}

		// 打开连接
		err = connect.Open()
		if err != nil {
			panic("Failed to open bus connect: " + err.Error())
		}

		inst := Instance{
			this, name, config, connect,
		}

		// 指定委托
		connect.Accept(&inst)

		//注册
		for msgName, _ := range this.services {
			realName := config.Prefix + msgName
			if err := connect.Register(realName); err != nil {
				panic("Failed to register bus: " + err.Error())
			}
		}

		//保存实例
		this.instances[name] = inst

		//只有设置了权重的才参与分布
		if config.Weight > 0 {
			weights[name] = config.Weight
		}
	}

	//hashring分片
	this.weights = weights
	this.hashring = util.NewHashRing(weights)

	this.connected = true
}
func (this *Module) Launch() {
	if this.launched {
		return
	}

	//全部开始来来来
	for _, inst := range this.instances {
		inst.connect.Start()
	}

	this.launched = true
}
func (this *Module) Terminate() {
	for _, ins := range this.instances {
		ins.connect.Close()
	}

	this.launched = false
	this.connected = false
	this.initialized = false
}

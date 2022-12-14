package bus

import (
	. "github.com/chefsgo/base"
)

// Configure 更新配置
func Configure(cfg Map) {
	module.Configure(cfg)
}

// Register 开放给外
func Register(name string, value Any, overrides ...bool) {
	override := true
	if len(overrides) > 0 {
		override = overrides[0]
	}
	module.Register(name, value, override)
}

// Ready 准备运行
// 此方法只单独使用模块时用
func Ready() {
	module.Initialize()
	module.Connect()
}

// Go 直接开跑
// 此方法只单独使用模块时用
func Go() {
	module.Initialize()
	module.Connect()
	module.Launch()
	module.Terminate()
}

package bus

import (
	"strings"

	"github.com/chefsgo/chef"
)

// 收到消息
// 待处理，要改成异步，做线程池
func (this *Instance) Serve(name string, data []byte, next Callback) {
	if strings.HasPrefix(name, this.config.Prefix) {
		name = strings.TrimPrefix(name, this.config.Prefix)
	}

	ctx := &Context{inst: this}
	ctx.Name = name
	ctx.callback = next

	if cfg, ok := this.module.services[ctx.Name]; ok {
		ctx.Config = &cfg
	}

	// 解析元数据
	metadata := chef.Metadata{}
	err := chef.Unmarshal(this.config.Codec, data, &metadata)
	if err == nil {
		ctx.Metadata(metadata)
	}

	//开始执行
	this.serve(ctx)
	chef.CloseMeta(&ctx.Meta)
}

func (this *Instance) serve(ctx *Context) {
	//清理执行线
	ctx.clear()

	//request拦截器
	ctx.next(this.module.requestFilters...)
	ctx.next(this.request)

	//开始执行
	ctx.Next()
}

// request 请求处理
func (this *Instance) request(ctx *Context) {
	ctx.clear()

	//request拦截器
	ctx.next(this.finding)     //存在否
	ctx.next(this.authorizing) //身份验证
	// ctx.next(this.arguing)     //参数处理，invoke自行处理
	ctx.next(this.execute)

	//开始执行
	ctx.Next()

	//response得在这里
	//这里才能被覆盖在request拦截器中
	this.response(ctx)
}

// execute 执行线
func (this *Instance) execute(ctx *Context) {
	ctx.clear()

	//execute拦截器
	ctx.next(this.module.executeFilters...)
	ctx.next(this.action)

	//开始执行
	ctx.Next()
}

func (this *Instance) action(ctx *Context) {
	ctx.Body = ctx.Invoke(ctx.Name, ctx.Value)
}

// response 响应线
func (this *Instance) response(ctx *Context) {
	ctx.clear()

	//response拦截器
	ctx.next(this.module.responseFilters...)
	ctx.next(this.body)

	//开始执行
	ctx.Next()
}

// finding 判断不
func (this *Instance) finding(ctx *Context) {
	if ctx.Config == nil {
		ctx.Found()
	} else {
		ctx.Next()
	}
}

// authorizing token验证
func (this *Instance) authorizing(ctx *Context) {
	// 待处理
	ctx.Next()
}

// arguing 参数解析
// 这里不做解析，invoke本身会处理
// func (this *Instance) arguing(ctx *Context) {
// 	if ctx.Config.Args != nil {
// 		argsValue := Map{}
// 		res := chef.Mapping(ctx.Config.Args, ctx.Value, argsValue, ctx.Config.Nullable, false, ctx.Timezone())
// 		if res != nil && res.Fail() {
// 			ctx.Failed(res)
// 		}

// 		for k, v := range argsValue {
// 			ctx.Args[k] = v
// 		}
// 	}
// 	ctx.Next()
// }

func (this *Instance) found(ctx *Context) {
	ctx.clear()

	//把处理器加入调用列表
	ctx.next(this.module.foundHandlers...)

	ctx.Next()
}

func (this *Instance) error(ctx *Context) {
	ctx.clear()

	//把处理器加入调用列表
	ctx.next(this.module.errorHandlers...)

	ctx.Next()
}

func (this *Instance) failed(ctx *Context) {
	ctx.clear()

	//把处理器加入调用列表
	ctx.next(this.module.failedHandlers...)

	ctx.Next()
}

func (this *Instance) denied(ctx *Context) {
	ctx.clear()

	//把处理器加入调用列表
	ctx.next(this.module.deniedHandlers...)

	ctx.Next()
}

// 最终的默认body响应
func (this *Instance) body(ctx *Context) {
	// switch ctx.Body.(type) {
	// default:
	// 	this.bodyDefault(ctx)
	// }

	if ctx.callback == nil {
		return
	}

	res := ctx.Result()
	if res == nil {
		res = chef.Fail
	}

	echo := &chef.Echo{}
	echo.Code = res.Code()
	echo.Text = res.Error()
	echo.Data = ctx.Body

	bytes, err := chef.Marshal(this.config.Codec, &echo)
	ctx.callback(bytes, err)
}

// bodyDefault 默认的body处理
// func (this *Instance) bodyDefault(ctx *Context) {

// }

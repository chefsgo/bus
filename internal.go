package bus

import (
	"time"

	"github.com/chefsgo/chef"
)

// 请求
func (this *Module) request(meta chef.Metadata, timeout time.Duration) (*chef.Echo, error) {
	locate := this.hashring.Locate(meta.Name)

	if inst, ok := this.instances[locate]; ok {
		//编码数据
		reqBytes, err := chef.Marshal(inst.config.Codec, &meta)
		if err != nil {
			return nil, err
		}

		realName := inst.config.Prefix + meta.Name
		resBytes, err := inst.connect.Request(realName, reqBytes, timeout)
		if err != nil {
			return nil, err
		}

		echo := &chef.Echo{Code: chef.OK.Code(), Text: chef.OK.Error()}

		//解码返回的数据
		err = chef.Unmarshal(inst.config.Codec, resBytes, echo)
		if err != nil {
			return nil, err
		}

		return echo, nil
	}

	return nil, errInvalidConnection
}

func (this *Module) Request(meta chef.Metadata, timeout time.Duration) (*chef.Echo, error) {
	return this.request(meta, timeout)
}

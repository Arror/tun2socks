package moon

import (
	"encoding/json"

	"github.com/xjasonlyu/tun2socks/v2/engine"

	_ "golang.org/x/mobile/bind"
)

type Instance struct{}

func (i *Instance) Start(data []byte) error {
	config := &engine.Config{}
	if err := json.Unmarshal(data, config); err != nil {
		return err
	}
	engine.Insert(config)
	return engine.Start()
}

func (i *Instance) Stop() error {
	return engine.Stop()
}

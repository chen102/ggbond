package routermanage

import (
	"errors"
	"fmt"

	"github.com/chen102/ggbond/conn/store"
)

var ErrorRouterManager error = errors.New("router manager error")

type RouterManager struct {
	store.ITCPStore // 存储具体数据
}
type RouterHandle func(msgid, connid int32, parameter []byte) error

func NewTCPRouter(store store.ITCPStore) *RouterManager {
	return &RouterManager{
		store,
	}
}

func (r *RouterManager) RegisterRoute(routeid int32, handler RouterHandle) error {
	if exists := r.Exist(routeid); exists {
		return fmt.Errorf("%w: %s ", ErrorRouterManager, "route ID already exists")
	}

	if _, err := r.Set(routeid, handler); err != nil {
		return fmt.Errorf("%w: %s ", ErrorRouterManager, err)
	}
	return nil
}

func (r *RouterManager) HandleMessage(routeid, connid, msgid int32, parameter []byte) error {
	handler, err := r.Get(routeid)
	if err != nil {
		return fmt.Errorf("%w: %s ", ErrorRouterManager, err)
	}
	return handler.(RouterHandle)(msgid, connid, parameter)
}

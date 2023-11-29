package routermanage

import (
	"errors"
	"fmt"

	"github.com/chen102/ggbond/store"
)

var ErrorRouterManager error = errors.New("router manager error")

type RouterManager struct {
	store.IStore // 存储具体数据
}

var _ IRouterManage = (*RouterManager)(nil)

func NewTCPRouterManager(store store.IStore) *RouterManager {
	return &RouterManager{
		store,
	}
}

func (r *RouterManager) RegisterRoute(routeID int32, handler RouterHandle) error {
	if exists := r.Exist(routeID); exists {
		return fmt.Errorf("%w: %s ", ErrorRouterManager, "route ID already exists")
	}

	if _, err := r.Set(routeID, handler); err != nil {
		return fmt.Errorf("%w: %s ", ErrorRouterManager, err)
	}
	return nil
}

func (r *RouterManager) HandleMessage(routeID int32, parameter []byte) error {
	handler, err := r.Get(routeID)
	if err != nil {
		return fmt.Errorf("%w: %s ", ErrorRouterManager, err)
	}
	return handler.(RouterHandle)(parameter)
}

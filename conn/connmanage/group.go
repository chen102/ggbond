package connmanage

import "errors"

type ConnGroup struct {
	groups map[string]GroupHook
	conns  map[int32]map[int32]struct{}
}

type GroupHook interface {
	ID() int32
	Name() string
}

func NewConnGroup() *ConnGroup {
	return &ConnGroup{
		groups: make(map[string]GroupHook),
		conns:  make(map[int32]map[int32]struct{}),
	}
}

func (m *ConnGroup) AddGroup(g GroupHook) error {
	if _, ok := m.groups[g.Name()]; ok {
		return errors.New("group already exists")
	}
	m.groups[g.Name()] = g
	return nil
}
func (m *ConnGroup) RemoveGroup(g GroupHook) error {
	if _, ok := m.groups[g.Name()]; !ok {
		return errors.New("group not exists")
	}
	delete(m.conns, g.ID())
	delete(m.groups, g.Name())
	return nil
}

func (m *ConnGroup) Group(g GroupHook) (map[int32]struct{}, error) {
	if _, ok := m.groups[g.Name()]; !ok {
		return nil, errors.New("group not exists")
	}
	return m.conns[g.ID()], nil
}
func (m *ConnGroup) AddConnToGroup(g GroupHook, conn int32) error {
	if _, ok := m.groups[g.Name()]; !ok {
		return errors.New("group not exists")
	}
	if _, ok := m.conns[g.ID()]; !ok {
		m.conns[g.ID()] = make(map[int32]struct{})
	}
	m.conns[g.ID()][conn] = struct{}{}
	return nil
}
func (m *ConnGroup) RemoveConnFromGroup(g GroupHook, conn int32) error {
	if _, ok := m.groups[g.Name()]; !ok {
		return errors.New("group not exists")
	}
	delete(m.conns[g.ID()], conn)
	return nil
}
func (m *ConnGroup) ClearGroup(g GroupHook) error {
	if _, ok := m.groups[g.Name()]; !ok {
		return errors.New("group not exists")
	}
	delete(m.conns, g.ID())
	return nil
}

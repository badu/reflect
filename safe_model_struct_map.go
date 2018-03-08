package reflector

import (
	. "reflect"
)

func (s *safeModelsMap) set(key Type, value *Model) {
	s.l.Lock()
	s.m[key] = value
	// we don't use defer here, because it's not needed
	s.l.Unlock()
}

func (s *safeModelsMap) get(key Type) *Model {
	s.l.RLock()
	defer s.l.RUnlock()
	return s.m[key]
}

//for listing in debug mode
func (s *safeModelsMap) getMap() map[Type]*Model {
	s.l.RLock()
	defer s.l.RUnlock()
	return s.m
}

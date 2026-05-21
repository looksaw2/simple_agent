package state

import (
	"github.com/looksaw/simple_agent_with_golang/internal/core/types"
)

// 多个节点同时写入数据
type Reducer func(current any, update any) any

// 定义具体的Key
type Key[T any] string

// 定义具体的存储
type Store struct {
	values   types.Map
	reducers map[string]Reducer
}

// 定义初始化Store的方式
func NewStore(reducers map[string]Reducer) *Store {
	return &Store{
		values:   make(types.Map),
		reducers: reducers,
	}
}

// 应用对应的Map资源
func (s *Store) Patch(updater types.Map) {
	for k, v := range updater {
		if reducer, exists := s.reducers[k]; exists {
			s.values[k] = reducer(s.values[k], v)
		} else {
			s.values[k] = v
		}
	}
}

// 得到对应的应用资源
func (s *Store) Values() types.Map {
	return s.values.ShallowCopy()
}

func GetSafe[T any](m types.Map, k Key[T]) T {
	val, exists := m[string(k)]
	if !exists {
		var zero T
		return zero
	}
	tVal, ok := val.(T)
	if !ok {
		var zeroT T
		return zeroT
	}
	return tVal
}

package types

type NodeID string

const (
	Start NodeID = "start"
	End   NodeID = "end"
)

// 一个Map的图的引擎
type Map map[string]any

// ShallowCopy
func (m Map) ShallowCopy() Map {
	nm := make(Map, len(m))
	for k, v := range m {
		nm[k] = v
	}
	return nm
}

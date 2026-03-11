package response

import "skeyevss/core/tps"

type ListResp[T any] struct {
	List  T     `json:"list,omitempty,optional"`
	Count int64 `json:"count,omitempty,optional"`
}

func NewListResp[T any]() *ListResp[T] {
	return new(ListResp[T])
}

func (r *ListResp[T]) Empty() *ListResp[T] {
	return &ListResp[T]{
		List:  *new(T),
		Count: 0,
	}
}

type OptionsResp struct {
	List  []*tps.OptionItem `json:"list"`
	Count int64             `json:"count"`
}

type ListWithMapResp[T any, K comparable] struct {
	List  T                      `json:"list,omitempty,optional"`
	Count int64                  `json:"count,omitempty,optional"`
	Maps  map[K]interface{}      `json:"maps,omitempty,optional"`
	Ext   map[string]interface{} `json:"ext,omitempty,optional"`
}

type ListWithExtResp[T, T1 any, T2 comparable] struct {
	List   T                  `json:"list,omitempty,optional"`
	Count  int64              `json:"count,omitempty,optional"`
	Slices T1                 `json:"slices,omitempty,optional"`
	Ext    map[T2]interface{} `json:"ext,omitempty,optional"`
}

package permissions

import (
	"context"
	"errors"
	"log"

	"skeyevss/core/pkg/functions"
)

var maps = make(map[IdType]*Item)

type Verify struct {
	ctx context.Context
}

func New(ctx context.Context) *Verify {
	return &Verify{ctx: ctx}
}

func (v *Verify) ConvStrSlice(ids []string) []IdType {
	var records []IdType
	for _, item := range ids {
		records = append(records, IdType(item))
	}

	return records
}

func (v *Verify) Authentication(super uint, uniqueId IdType, permissionUniqueIds []string) error {
	item, ok := maps[uniqueId]
	if !ok {
		return errors.New("权限未配置")
	}

	if super > 0 {
		return nil
	}

	if item.Universal {
		return nil
	}

	if !functions.Contains(uniqueId, v.ConvStrSlice(permissionUniqueIds)) {
		return errors.New("无权限[001]")
	}

	if item.Super {
		return errors.New("无权限[002]")
	}

	return nil
}

func initVerifyData(data []*Item, level int) {
	for _, item := range data {
		item.Level = level
		if _, ok := maps[item.UniqueId]; ok {
			log.Fatalf("权限id重复定义, id: %s, name: %s", item.UniqueId, item.Name)
		}

		maps[item.UniqueId] = item
		if len(item.Children) > 0 {
			initVerifyData(item.Children, item.Level+1)
		}
	}
}

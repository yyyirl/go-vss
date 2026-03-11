// @Title        common
// @Description  main
// @Create       yiyiyi 2025/12/1 13:22

package logs

import (
	"path"
	"path/filepath"

	"skeyevss/core/pkg/functions"
)

type item struct {
	Id    uint   `json:"id"`
	Path  string `json:"path"`
	Ext   string `json:"ext"`
	Name  string `json:"name"`
	IsDir bool   `json:"isDir"`
	Level int    `json:"level"`
}

func logFilesTidy(data []*functions.FileTreeNode, parent string) []*item {
	var (
		id      uint = 0
		records []*item
		call    func(data []*functions.FileTreeNode, parent string)
	)
	call = func(data []*functions.FileTreeNode, parent string) {
		for _, v := range data {
			id += 1
			var (
				_parent = path.Join(parent, v.Name)
				ext     = filepath.Ext(v.Name)
			)
			if ext == ".log" {
				records = append(records, &item{
					Id:    id,
					Path:  path.Join(parent, v.Name),
					Name:  v.Name,
					Ext:   ext,
					IsDir: v.IsDir,
					Level: v.Level,
				})
			}
			if len(v.Children) > 0 {
				call(v.Children, _parent)
			}
		}
	}

	call(data, parent)
	if len(records) <= 0 {
		return []*item{}
	}
	return records
}

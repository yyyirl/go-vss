// @Title        save
// @Description  main
// @Create       yiyiyi 2025/8/18 11:55

package svideo

import (
	"fmt"
	"os"

	"skeyevss/core/pkg/functions"
)

func (s *SVideo) Delete(filenames []string) {
	for _, item := range filenames {
		_ = os.Remove(item)
	}
}

func (s *SVideo) SetUrgent(state bool, filename string) error {
	var data = s.Parse(filename, 0)
	if data == nil {
		return nil
	}

	if state == data.Urgent {
		return nil
	}

	data.Urgent = state
	return functions.Mv(
		fmt.Sprintf("%s/%s", s.SaveDir, filename),
		fmt.Sprintf("%s/%s", s.SaveDir, s.MakeFileName(data)),
	)
}

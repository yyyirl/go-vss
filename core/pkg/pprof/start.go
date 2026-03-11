// @Title        start
// @Description  main
// @Create       yiyiyi 2025/9/9 15:50

package pprof

import (
	"fmt"
	"log"
	"net/http"
	_ "net/http/pprof"

	"skeyevss/core/pkg/functions"
)

func Start(port uint, dir string) {
	if err := functions.MakeDir(dir); err != nil {
		panic(err)
	}

	go func() {
		log.Println(http.ListenAndServe(fmt.Sprintf("localhost:%d", port), nil))
	}()
}

/**
 * @Author:         yi
 * @Description:    main
 * @Version:        1.0.0
 * @Date:           2022/11/6 10:42
 */
package elasticsearch

import (
	"errors"
	"fmt"
	"log"
	"os"
	"strings"

	elastic "github.com/olivere/elastic/v7"

	"skeyevss/core/constants"
	"skeyevss/core/pkg/functions"
	"skeyevss/core/tps"
)

var ErrNotFound = errors.New("没有更多数据")

type Elasticsearch struct {
	*elastic.Client
	mode string
}

func New(conf *tps.YamlElasticsearch, mode string) *Elasticsearch {
	c, err := elastic.NewClient(
		// elastic 服务地址
		elastic.SetURL("http://"+conf.Host+":"+conf.Port),
		// 设置走外网 默认内网
		elastic.SetSniff(false),
		// 权限
		elastic.SetBasicAuth(
			conf.Username,
			conf.Password,
		),
		// 设置错误日志输出
		elastic.SetErrorLog(log.New(os.Stderr, "ELASTIC ", log.LstdFlags)),
		// 设置info日志输出
		elastic.SetInfoLog(log.New(os.Stdout, "ELASTIC INFO ", log.LstdFlags)),
		// 日志输出
		elastic.SetTraceLog(&elasticsearchLogger{mode}),
	)

	if err != nil {
		panic(err)
	}

	return &Elasticsearch{c, mode}
}

type elasticsearchLogger struct {
	mode string
}

func (this elasticsearchLogger) Printf(_ string, v ...interface{}) {
	if this.mode == constants.ENV_DEVELOPMENT {
		if len(v) <= 0 {
			return
		}

		var str = fmt.Sprintf("%v", v[0])
		if strings.Index(str, "HTTP/1.1 200 OK") == 0 {
			return
		}

		functions.PrintStyle("red", "------------------------- an elasticsearch request start ... \n")
		println(strings.TrimSpace(str))
		functions.PrintStyle("red", "------------------------- an elasticsearch request end ... \n")
	}
}

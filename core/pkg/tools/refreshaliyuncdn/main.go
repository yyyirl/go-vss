/**
 * @Author:         yi
 * @Description:    刷新阿里云cdn
 * @Version:        1.0.0
 * @Date:           2023/8/11 16:45
 */
package main

import (
	"flag"
	"os"

	"github.com/aliyun/alibaba-cloud-sdk-go/services/cdn"

	"skeyevss/core/pkg/functions"
)

var (
	accessId  = flag.String("id", "", "阿里云access id")
	accessKey = flag.String("key", "", "阿里云access key")
	directory = flag.String("directory", "", "刷新目录 directory")
)

func main() {
	flag.Parse()

	if accessId == nil || *accessId == "" {
		functions.PrintStyle("red", "id不能为空")
		println(params())
		os.Exit(1)
	}

	if accessKey == nil || *accessKey == "" {
		functions.PrintStyle("red", "key不能为空")
		println(params())
		os.Exit(1)
	}

	if directory == nil || *directory == "" {
		functions.PrintStyle("red", "directory不能为空")
		println(params())
		os.Exit(1)
	}

	functions.PrintStyle("yellow", "参数: id: ", *accessId)
	functions.PrintStyle("yellow", "参数: key: ", *accessKey)
	functions.PrintStyle("yellow", "参数: directory: ", *directory)

	// 更新cdn缓存
	client, err := cdn.NewClientWithAccessKey(
		"",
		*accessId,
		*accessKey,
	)

	if err != nil {
		functions.PrintStyle("red", "阿里云客户端创建失败, err: ", err.Error())
		os.Exit(1)
	}

	var request = cdn.CreateRefreshObjectCachesRequest()
	request.Scheme = "https"
	request.ObjectType = "Directory"
	request.ObjectPath = *directory

	if _, err = client.RefreshObjectCaches(request); err != nil {
		functions.PrintStyle("yellow", "刷新失败, err: ", err.Error())
		os.Exit(1)
	}
}

func params() string {
	return `请求参数:
    -id: 阿里云access id
    -key: 阿里云access key
    -directory: 刷新地址 例如:https://www.guga.co/
	
	refresh_alicdn -id ${SN_ALIYUN_ACCESS_ID} -key ${SN_ALIYUN_ACCESS_KEY} -directory '${SN_ALIYUN_FRONTEND_CDN_REFRESH_PATH}'
`
}

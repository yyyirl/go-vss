// This file is auto-generated, don't edit it. Thanks.
package main

import (
	"bytes"
	"encoding/json"
	"flag"
	"os"

	openapi "github.com/alibabacloud-go/darabonba-openapi/v2/client"
	openapiutil "github.com/alibabacloud-go/openapi-util/service"
	util "github.com/alibabacloud-go/tea-utils/v2/service"
	"github.com/alibabacloud-go/tea/tea"

	"skeyevss/core/pkg/functions"
)

const (
	add    = "AddBackendServers"
	remove = "RemoveBackendServers"
)

var (
	accessId       = flag.String("id", "", "阿里云access id")
	accessKey      = flag.String("key", "", "阿里云access key")
	regionId       = flag.String("region-id", "", "负载均衡实例所属地域的ID [cn-beijing]")
	loadBalancerId = flag.String("load-balancer-id", "", "负载均衡实例ID")
	backendServer  = flag.String("backend-servers", "", "要添加的后端服务器列表")
	port           = flag.String("port", "", "slb监听的端口")
	Type           = flag.String("type", "", "添加或删除[add remove]")
)

func main() {
	flag.Parse()

	if accessId == nil || *accessId == "" {
		functions.PrintStyle("red", "id 不能为空")
		println(params())
		os.Exit(1)
	}

	if accessKey == nil || *accessKey == "" {
		functions.PrintStyle("red", "key 不能为空")
		println(params())
		os.Exit(1)
	}

	if regionId == nil || *regionId == "" {
		functions.PrintStyle("red", "region-id 不能为空")
		println(params())
		os.Exit(1)
	}

	if loadBalancerId == nil || *loadBalancerId == "" {
		functions.PrintStyle("red", "load-balancer-id 不能为空")
		println(params())
		os.Exit(1)
	}

	if backendServer == nil || *backendServer == "" {
		functions.PrintStyle("red", "backend-servers 不能为空")
		println(params())
		os.Exit(1)
	}

	if Type == nil || *Type == "" {
		functions.PrintStyle("red", "type 不能为空")
		println(params())
		os.Exit(1)
	}

	if port == nil || *port == "" {
		functions.PrintStyle("red", "port 不能为空")
		println(params())
		os.Exit(1)
	}

	res, err := backendServersAction(*accessId, *accessKey, *regionId, *loadBalancerId, *backendServer, *port, *Type)
	if err != nil {
		println("error: slb后台服务器操作失败", err.Error())
		os.Exit(1)
	}

	println("success: slb后台服务器操作成功")
	jj, _ := json.Marshal(res)
	var str bytes.Buffer
	_ = json.Indent(&str, jj, "", "    ")
	println(str.String())
}

func params() string {
	return `请求参数:
	-id: 阿里云access id
	-key: 阿里云access key
	-region-id: 负载均衡实例所属地域的ID [cn-beijing] 
	-load-balancer-id: 负载均衡实例ID
	-backend-servers: 要添加的后端服务器列表
	-port: slb监听的端口
	-type: 添加或删除[add remove]

	aliyunslbbackendserver -id 阿里云accessid -key 阿里云accesskey -region-id cn-beijing -load-balancer-id slbid -backend-servers ecsid -port 443 -type add
	aliyunslbbackendserver -id 阿里云accessid -key 阿里云accesskey -region-id cn-beijing -load-balancer-id slbid -backend-servers ecsid -port 443 -type remove
`
}

// API 相关
func CreateApiInfo(Type string) (result *openapi.Params) {
	var actionType = add
	if Type == "remove" {
		actionType = remove
	}

	result = &openapi.Params{
		// 接口名称
		Action: tea.String(actionType),
		// 接口版本
		Version: tea.String("2014-05-15"),
		// 接口协议
		Protocol: tea.String("HTTPS"),
		// 接口 HTTP 方法
		Method:   tea.String("POST"),
		AuthType: tea.String("AK"),
		Style:    tea.String("RPC"),
		// 接口 PATH
		Pathname: tea.String("/"),
		// 接口请求体内容格式
		ReqBodyType: tea.String("json"),
		// 接口响应体内容格式
		BodyType: tea.String("json"),
	}

	return result
}

// 使用AK&SK初始化账号Client
func createClient(accessKeyId *string, accessKeySecret *string) (result *openapi.Client, _err error) {
	var config = &openapi.Config{
		AccessKeyId:     accessKeyId,
		AccessKeySecret: accessKeySecret,
	}

	// Endpoint 请参考 https://api.aliyun.com/product/Slb
	config.Endpoint = tea.String("slb.aliyuncs.com")
	result = &openapi.Client{}
	result, _err = openapi.NewClient(config)

	return result, _err
}

func backendServersAction(accessId, accessKey, regionId, loadBalancerId, backendServer, port, Type string) (map[string]interface{}, error) {
	client, err := createClient(
		tea.String(accessId),
		tea.String(accessKey),
	)
	if err != nil {
		return nil, err
	}

	// query params
	var queries = map[string]interface{}{}
	queries["RegionId"] = tea.String(regionId)
	queries["LoadBalancerId"] = tea.String(loadBalancerId)
	queries["BackendServers"] = itemJsonString(backendServer, port)

	// 复制代码运行请自行打印 API 的返回值
	// 返回值为 Map 类型，可从 Map 中获得三类数据：响应体 body、响应头 headers、HTTP 返回的状态码 statusCode。
	return client.CallApi(
		CreateApiInfo(Type),
		&openapi.OpenApiRequest{
			Query: openapiutil.Query(queries),
		},
		&util.RuntimeOptions{},
	)
}

// [{"ServerId": "i-2zeeqch74saglprlfrer", "Weight": "100", "Type": "ecs", "Port": "443", "Description": "172.31.89.126"}]
func itemJsonString(serverId, port string) string {
	return `[{"ServerId": "` + serverId + `", "Weight": "100", "Type": "ecs", "Port": "` + port + `", "Description": "` + serverId + `"}]`
}

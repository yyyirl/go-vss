/**
 * @Author:         yi
 * @Description:    request
 * @Version:        1.0.0
 * @Date:           2022/11/17 11:38
 */
package functions

import (
	"bytes"
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"strings"
	"time"
)

type req struct {
	mode string
}

func NewRequest(mode string) *req {
	return &req{
		mode: mode,
	}
}

func (this *req) GetWithHeader(url string, header map[string]string) ([]byte, error) {
	client := &http.Client{}
	req, err := http.NewRequest("GET", url, nil)

	for key, item := range header {
		req.Header.Add(key, item)
	}

	if err != nil {
		return nil, err
	}
	// 处理返回结果
	resp, _ := client.Do(req)
	defer func(Body io.ReadCloser) {
		_ = Body.Close()
	}(resp.Body)

	defer func(body io.ReadCloser) {
		_ = body.Close()
	}(resp.Body)

	if resp.Status != "200 OK" {
		return nil, errors.New(resp.Status)
	}

	body, _ := ioutil.ReadAll(resp.Body)
	return body, nil
}

// PostJson post提交json
func (this *req) PostJson(u string, params interface{}, dump bool) (body []byte, err error) {
	return this.ReqJson("POST", u, params, nil, dump)
}

func (this *req) PostFormWithHeader(u string, params interface{}, headers map[string]string, dump bool) (body []byte, err error) {
	paramsJson, _ := JSONMarshal(params)
	req, err := http.NewRequest("POST", u, bytes.NewBuffer(paramsJson))
	if err != nil {
		return
	}

	if this.mode == "dev" && dump {
		LogInfo("请求参数: \n", params)
	}

	for key, value := range headers {
		req.Header.Add(key, value)
	}

	var client = &http.Client{Timeout: 20 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return
	}

	defer func(Body io.ReadCloser) {
		_ = Body.Close()
	}(resp.Body)

	body, _ = ioutil.ReadAll(resp.Body)
	if resp.Status != "200 OK" {
		return body, errors.New(resp.Status)
	}

	if this.mode == "dev" && dump {
		LogInfo("响应: \n", string(body))
	}

	return body, nil
}

// PostJson post提交json
func (this *req) ReqJson(method, u string, params interface{}, reqFn func(req *http.Request), dump bool) (body []byte, err error) {
	paramsJson, _ := JSONMarshal(params)
	req, err := http.NewRequest(method, u, bytes.NewBuffer(paramsJson))
	if err != nil {
		return
	}

	if this.mode == "dev" && dump {
		// logger.Log("info", "\n 请求参数: %+v \n", params)
	}

	req.Header.Set("Content-Type", "application/json")
	if reqFn != nil {
		reqFn(req)
	}

	timeout := 20 * time.Second
	client := &http.Client{
		Timeout: timeout,
		// Transport: &http.Transport{
		//	Dial: func(netW, addr string) (net.Conn, error) {
		//		deadline := time.Now().Add(25 * time.Second)
		//		c, err := net.DialTimeout(netW, addr, time.Second*25)
		//		if err != nil {
		//			return nil, err
		//		}
		//		_ = c.SetDeadline(deadline)
		//		return c, nil
		//	},
		// },
	}
	resp, err := client.Do(req)
	if err != nil {
		return
	}

	defer func(Body io.ReadCloser) {
		_ = Body.Close()
	}(resp.Body)

	body, _ = ioutil.ReadAll(resp.Body)
	if resp.Status != "200 OK" {
		return body, errors.New(resp.Status)
	}

	if this.mode == "dev" && dump {
		// logger.Log("info", "\n 响应: %+v \n", string(body))
	}

	return body, nil
}

// Get http get
func Get(url string) ([]byte, error) {
	res, err := http.Get(url)
	if err != nil {
		return nil, err
	}

	defer func() {
		_ = res.Body.Close()
	}()

	if res.StatusCode != 200 {
		return nil, fmt.Errorf("http get error-> status = %d", res.StatusCode)
	}

	robots, err := ioutil.ReadAll(res.Body)
	if err != nil {
		return nil, err
	}

	return robots, nil
}

func PostXml(url, data string, call func(*http.Response, error)) {
	resp, err := http.Post(
		url,
		"text/xml",
		strings.NewReader(data),
	)

	if call != nil {
		call(resp, err)
	}

	defer func() {
		_ = resp.Body.Close()
	}()
}

// GetRedirect 获取301地址
func GetRedirect(u string) string {
	client := &http.Client{
		CheckRedirect: func(req *http.Request, via []*http.Request) error {
			return http.ErrUseLastResponse
		},
	}

	res, err := client.Get(u)
	if err != nil {
		return u
	}

	if res.StatusCode != 302 {
		return u
	}

	return res.Header.Get("Location")
}

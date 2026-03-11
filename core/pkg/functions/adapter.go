/**
 * @Author:         yi
 * @Description:    Adapter
 * @Version:        1.0.0
 * @Date:           2022/12/29 15:56
 */
package functions

import (
	"bytes"
	"fmt"
	"io"
	"net/http"
	"sync"
)

type Adapter struct {
	pool sync.Pool
}

// https://cloud.tencent.com/developer/article/1532122
func NewAdapter() *Adapter {
	return &Adapter{
		pool: sync.Pool{
			New: func() interface{} {
				return bytes.NewBuffer(make([]byte, 4096))
			},
		},
	}
}

func (this *Adapter) GetReqBody(r *http.Request) ([]byte, error) {
	var buffer = this.pool.Get().(*bytes.Buffer)
	buffer.Reset()
	defer func() {
		if buffer != nil {
			this.pool.Put(buffer)
			buffer = nil
		}
	}()
	_, err := io.Copy(buffer, r.Body)
	if err != nil {
		return nil, fmt.Errorf("adapter io.copy failure error:%v", err)
	}

	return buffer.Bytes(), nil
}

// func ParseReqJsonToTRFParams(r *http.Request) {
//	b, err := functions.NewAdapter().GetReqBody(r)
//	if err != nil {
//		return
//	}
//
//	// b, err := ioutil.ReadAll(r.Body)
//	// if err != nil {
//	// 	return nil, err
//	// }
//
//	var data types.TRFParams
//	if err := functions.JSONUnmarshal(b, &data); err != nil {
//		return
//	}
//
//	jj, _ := json.Marshal(data)
//	var str bytes.Buffer
//	_ = json.Indent(&str, jj, "", "    ")
//	fmt.Printf("\n format: %+v \n", str.String())
//
//	return
// }

/**
 * @Author:         yi
 * @Description:    json
 * @Version:        1.0.0
 * @Date:           2022/10/10 12:07
 */
package functions

import jsonIter "github.com/json-iterator/go"

// json
var j = jsonIter.ConfigCompatibleWithStandardLibrary

func JSONMarshal(data interface{}) ([]byte, error) {
	return j.Marshal(&data)
}

func JSONUnmarshal(input []byte, v interface{}) error {
	return j.Unmarshal(input, &v)
}

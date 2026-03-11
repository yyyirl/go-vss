/**
 * @Author:         yi
 * @Description:    string
 * @Version:        1.0.0
 * @Date:           2024/12/23 17:13
 */
package functions

import (
	"encoding/json"
	"fmt"
	"math/rand"
	"net/url"
	"reflect"
	"strconv"
	"strings"
	"time"
	"unicode"
	"unicode/utf8"

	"github.com/rs/xid"

	"skeyevss/core/tps"
)

// origin 源字符串
// index 切割位置索引
// index
// SplitWithIndex 将字符串从指定位置切割 返回数组
func SplitWithIndex(origin string, args ...int) [2]string {
	rawStrSlice := []byte(origin)
	var length = len(args)
	if length <= 0 {
		return [2]string{"", ""}
	} else if length == 1 {
		return [2]string{
			string(rawStrSlice[:args[0]]),
			string(rawStrSlice[args[0]:]),
		}
	} else {
		return [2]string{
			string(rawStrSlice[:args[0]]),
			string(rawStrSlice[args[1]:]),
		}
	}
}

func RandString(length int) string {
	r := rand.New(rand.NewSource(time.Now().UnixNano()))
	bytes := make([]byte, 0, length)
	for i := 0; i < length; i++ {
		b := r.Intn(26) + 65
		bytes = append(bytes, byte(b))
	}
	return string(bytes)
}

func RandWithString(source string, length int) string {
	rand.Seed(time.Now().UnixNano())

	var (
		sourceRunes = []rune(source)
		n           = len(sourceRunes)
	)
	if n == 0 || length <= 0 {
		return ""
	}

	result := make([]rune, length)
	for i := range result {
		result[i] = sourceRunes[rand.Intn(n)]
	}

	return string(result)
}

func TrimBlankChar(str string) string {
	return strings.Replace(
		strings.Replace(
			str,
			"	",
			"",
			-1,
		),
		"\n",
		"",
		-1,
	)
}

// Substr 字符串截取
func Substr(source string, start int, end int) string {
	if source == "" {
		return ""
	}

	var (
		r      = []rune(source)
		length = len(r)
	)

	if start < 0 || end > length || start > end || (start == 0 && end == length) {
		return source
	}

	if end >= length {
		end = length
	}

	return string(r[start:end])
}

func StrToMap(input string) (map[string]interface{}, error) {
	if input == "" {
		return nil, nil
	}

	var data = make(map[string]interface{})
	if err := json.Unmarshal([]byte(input), &data); err != nil {
		return nil, err
	}

	return data, nil
}

func MapToStr(input map[string]interface{}) (string, error) {
	if input == nil {
		return "", nil
	}

	b, err := JSONMarshal(input)
	if err != nil {
		return "", err
	}

	return string(b), nil
}

// 首字母大写
func Capitalize(s string) string {
	if len(s) == 0 {
		return s
	}

	firstRune, size := utf8.DecodeRuneInString(s)
	var capitalizedFirstRune rune

	// 检查并转换首字符
	if unicode.IsLetter(firstRune) { // 如果是字母
		capitalizedFirstRune = unicode.ToUpper(firstRune)
	} else {
		capitalizedFirstRune = firstRune
	}

	return string(capitalizedFirstRune) + s[size:]
}

func GetCCode(length int) string {
	var (
		randomStrSlice = strings.Split(Md5String(xid.New().String()), "")

		code = []string{firstChar()}
		i    = 0
		_len = len(randomStrSlice)
		n    = _len - 1
	)

	for {
		if i >= length-1 {
			break
		}

		if i == 3 || (i+1)%6 == 0 && i != 0 {
			i++
			code = append(code, "-")
			continue
		}
		i++

		if n < 0 {
			n = _len - 1
		}

		code = append(code, randomStrSlice[n])
		n--
	}

	return strings.ToUpper(strings.Join(code, ""))
}

func firstChar() string {
	var (
		str   = strings.Builder{}
		chars = []string{"A", "B", "C", "D", "E", "F", "G", "H", "I", "J", "K", "L", "M", "N", "O", "P", "Q", "R", "S", "T", "U", "V", "W", "X", "Y", "Z"}
	)

	rand.Seed(time.Now().UnixNano())
	str.WriteString(chars[rand.Intn(len(chars))])

	return str.String()
}

func GetCode() string {
	return fmt.Sprintf(
		"%06v",
		rand.New(rand.NewSource(time.Now().UnixNano())).Int31n(1000000),
	)
}

func GetFirstNChars(s string, n int) string {
	runes := []rune(s)
	if len(runes) <= n {
		return s
	}
	return string(runes[:n])
}

func GetLastNChars(s string, n int) string {
	runes := []rune(s)
	if len(runes) <= n {
		return s
	}
	return string(runes[len(runes)-n:])
}

func GetMiddleNChars(s string, n int) string {
	runes := []rune(s)
	length := len(runes)

	if length < n {
		return s
	}

	start := (length - n) / 2
	return string(runes[start : start+n])
}

func GetSubstringFromLastChar(s string, char string) string {
	var lastIndex = strings.LastIndex(s, char)
	if lastIndex == -1 {
		return ""
	}

	return s[:lastIndex]
}

// func ReverseStr(s string) string {
// 	var (
// 		str    = []rune(s)
// 		length = len(str)
// 	)
// 	for i := 0; i < length/2; i++ {
// 		str[i], str[length-1-i] = str[length-1-i], str[i]
// 	}
// 	return string(str)
// }

func ReverseStr(str string) string {
	if str == "" {
		return ""
	}

	var s = []rune(str)
	for i, j := 0, len(s)-1; i < j; i, j = i+1, j-1 {
		s[i], s[j] = s[j], s[i]
	}
	return string(s)
}

func GenerateRandomString(length int) (string, error) {
	var byteArray = make([]byte, length)
	_, err := rand.Read(byteArray)
	if err != nil {
		return "", err
	}

	for i := range byteArray {
		byteArray[i] = charset[int(byteArray[i])%len(charset)]
	}

	return string(byteArray), nil
}

// func InterfaceToInt64(v interface{}) (int64, error) {
// 	if tmp, ok := v.(float64); ok {
// 		return int64(tmp), nil
// 	}
//
// 	if tmp, ok := v.(int64); ok {
// 		return tmp, nil
// 	}
//
// 	val, ok := v.(json.Number)
// 	if ok {
// 		f, err := val.Int64()
// 		if err != nil {
// 			return 0, nil
// 		}
//
// 		return f, nil
// 	}
//
// 	return 0, errors.New(fmt.Sprintf("类型错误, 输入类型: %T, 值: %v", v, v))
// }

// func InterfaceToNumber[T tps.Number](v interface{}) (T, error) {
// 	val, ok := v.(json.Number)
// 	if ok {
// 		f, err := val.Int64()
// 		if err != nil {
// 			return 0, nil
// 		}
//
// 		return T(f), nil
// 	}
//
// 	switch v.(type) {
// 	case int:
// 		return T(int(v)), nil
// 	case int8:
// 		return T(int8(v)), nil
// 	case int32:
// 		return T(int32(v)), nil
// 	case int64:
// 		return T(int64(v)), nil
// 	case uint:
// 		return T(uint(v)), nil
// 	case uint8:
// 		return T(uint8(v)), nil
// 	case uint32:
// 		return T(uint32(v)), nil
// 	case uint64:
// 		return T(uint64(v)), nil
// 	case float64:
// 		return T(float6(v)), nil
// 	case float32:
// 		return T(float3(v)), nil
// 	}
//
// 	return 0, errors.New(fmt.Sprintf("类型错误, 输入类型: %T, 值: %v", v, v))
// }

// func InterfaceToNumber[uint](v interface{}) (uint, error) {
// 	if tmp, ok := v.(float64); ok {
// 		return uint(tmp), nil
// 	}
//
// 	if tmp, ok := v.(uint); ok {
// 		return tmp, nil
// 	}
//
// 	val, ok := v.(json.Number)
// 	if ok {
// 		f, err := val.Int64()
// 		if err != nil {
// 			return 0, nil
// 		}
//
// 		return uint(f), nil
// 	}
//
// 	return 0, errors.New(fmt.Sprintf("类型错误, 输入类型: %T, 值: %v", v, v))
// }

// func InterfaceToNumber[int](v interface{}) (int, error) {
// 	if tmp, ok := v.(float64); ok {
// 		return int(tmp), nil
// 	}
//
// 	if tmp, ok := v.(int); ok {
// 		return tmp, nil
// 	}
//
// 	val, ok := v.(json.Number)
// 	if ok {
// 		f, err := val.Int64()
// 		if err != nil {
// 			return 0, nil
// 		}
//
// 		return int(f), nil
// 	}
//
// 	return 0, errors.New(fmt.Sprintf("类型错误, 输入类型: %T, 值: %v", v, v))
// }

func SliceInterfaceToNumber[T tps.Number](ipt interface{}) ([]T, error) {
	if ipt == nil {
		return nil, nil
	}

	var data []T
	if err := ConvInterface(ipt, &data); err != nil {
		return nil, err
	}

	return data, nil
}

func InterfaceToNumber[T tps.Number](value interface{}) (T, error) {
	if v, ok := value.(json.Number); ok {
		f, err := v.Int64()
		if err != nil {
			return 0, nil
		}

		return T(f), nil
	}

	var zero T
	if value == nil {
		return zero, fmt.Errorf("cannot convert nil to number")
	}

	if v, ok := value.(T); ok {
		return v, nil
	}

	val := reflect.ValueOf(value)
	switch val.Kind() {
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		return T(val.Int()), nil
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
		return T(val.Uint()), nil
	case reflect.Float32, reflect.Float64:
		return T(val.Float()), nil
	case reflect.String:
		str := val.String()
		if str == "" {
			return zero, fmt.Errorf("empty string cannot be converted to number")
		}
		return parseStringToNumber[T](str)
	default:
		return zero, fmt.Errorf("unsupported type: %T", value)
	}
}

func parseStringToNumber[T tps.Number](str string) (T, error) {
	var zero T
	var v any
	var err error

	switch any(zero).(type) {
	case int:
		v, err = strconv.Atoi(str)
	case int8:
		var i int64
		i, err = strconv.ParseInt(str, 10, 8)
		v = int8(i)
	case int16:
		var i int64
		i, err = strconv.ParseInt(str, 10, 16)
		v = int16(i)
	case int32:
		var i int64
		i, err = strconv.ParseInt(str, 10, 32)
		v = int32(i)
	case int64:
		v, err = strconv.ParseInt(str, 10, 64)
	case uint:
		var u uint64
		u, err = strconv.ParseUint(str, 10, 64)
		v = uint(u)
	case uint8:
		var u uint64
		u, err = strconv.ParseUint(str, 10, 8)
		v = uint8(u)
	case uint16:
		var u uint64
		u, err = strconv.ParseUint(str, 10, 16)
		v = uint16(u)
	case uint32:
		var u uint64
		u, err = strconv.ParseUint(str, 10, 32)
		v = uint32(u)
	case uint64:
		v, err = strconv.ParseUint(str, 10, 64)
	case float32:
		var f float64
		f, err = strconv.ParseFloat(str, 32)
		v = float32(f)
	case float64:
		v, err = strconv.ParseFloat(str, 64)
	default:
		return zero, fmt.Errorf("unsupported number type: %T", zero)
	}

	if err != nil {
		return zero, fmt.Errorf("failed to parse string '%s' to %T: %v", str, zero, err)
	}
	return v.(T), nil
}

type ExtractBaseURLType struct {
	Username,
	Password,
	Scheme,
	IP,
	Url,
	Host string
	Port uint
}

func ExtractBaseURL(rawURL string) (*ExtractBaseURLType, error) {
	if strings.Index(rawURL, "://") < 0 {
		rawURL = "http://" + rawURL
	}

	parsedURL, err := url.Parse(rawURL)
	if err != nil {
		return nil, err
	}

	var (
		password,
		username string
	)
	if parsedURL.User != nil {
		username = parsedURL.User.Username()
		password, _ = parsedURL.User.Password()
	}

	var (
		portStr        = parsedURL.Port()
		port    uint64 = 80
	)
	if portStr != "" {
		port, err = strconv.ParseUint(portStr, 10, 16)
		if err != nil {
			return nil, err
		}
	}

	return &ExtractBaseURLType{
		Username: username,
		Password: password,
		Scheme:   parsedURL.Scheme,
		IP:       strings.Split(parsedURL.Host, ":")[0],
		Port:     uint(port),
		Host:     parsedURL.Host,
		Url:      fmt.Sprintf("%s://%s", parsedURL.Scheme, parsedURL.Host),
	}, nil
}

func TruncateString(s string, n int) string {
	if len(s) <= n {
		return s
	}
	return s[:n]
}

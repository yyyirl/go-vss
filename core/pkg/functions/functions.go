package functions

import (
	"bytes"
	"crypto/md5"
	"database/sql"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"math"
	"math/rand"
	"net"
	"net/http"
	"os"
	"reflect"
	"regexp"
	"runtime"
	"strconv"
	"strings"
	"sync/atomic"
	"time"

	"github.com/rs/xid"
)

func GetEnvDefault(key, defVal string) string {
	val, ex := os.LookupEnv(key)
	if !ex {
		return defVal
	}
	return val
}

func ToUniqueId(data interface{}) (string, error) {
	b, err := JSONMarshal(data)
	if err != nil {
		return "", err
	}

	return Md5String(string(b)), nil
}

func Caller(skip int) string {
	pc, file, line, status := runtime.Caller(skip)
	caller := runtime.FuncForPC(pc).Name()

	return fmt.Sprintf(", file: %s:%d:%t\nfuncName:%s", file, line, status, caller)
}

func CallerFile(skip int) string {
	_, file, line, _ := runtime.Caller(skip)
	currentDir, _ := os.Getwd()

	return strings.ReplaceAll(file, currentDir, "") + ":" + strconv.Itoa(line)
}

func CallerFileFull(skip int) string {
	_, file, line, _ := runtime.Caller(skip)
	return file + ":" + strconv.Itoa(line)
}

func IsPtrStruct(in interface{}) bool {
	var v = reflect.ValueOf(in)

	if v.Kind() != reflect.Ptr {
		return false
	}

	if v.Elem().Kind() != reflect.Struct {
		return false
	}

	return true
}

func IsStruct(in interface{}) bool {
	return reflect.ValueOf(in).Kind() == reflect.Struct
}

func IsPtr(in interface{}) bool {
	return reflect.ValueOf(in).Kind() == reflect.Ptr
}

func IsSlice(in interface{}) bool {
	return reflect.ValueOf(in).Kind() == reflect.Slice
}

func IsMap(in interface{}) bool {
	return reflect.ValueOf(in).Kind() == reflect.Map
}

// 获取本机网卡IP
func GetLocalIP() (string, error) {
	var (
		addresses []net.Addr
		addr      net.Addr
		ipNet     *net.IPNet // IP地址
		isIpNet   bool
		err       error
	)
	// 获取所有网卡
	if addresses, err = net.InterfaceAddrs(); err != nil {
		return "", err
	}
	// 取第一个非lo的网卡IP
	for _, addr = range addresses {
		// 这个网络地址是IP地址: ipv4, ipv6
		if ipNet, isIpNet = addr.(*net.IPNet); isIpNet && !ipNet.IP.IsLoopback() {
			// 跳过IPV6
			if ipNet.IP.To4() != nil {
				return ipNet.IP.String(), nil
			}
		}
	}

	return "", errors.New("未获取到本机ip")
}

func GetIP(r *http.Request) (string, error) {
	ip := r.Header.Get("X-Forwarded-For")
	for _, i := range strings.Split(ip, ",") {
		if net.ParseIP(i) != nil {
			return i, nil
		}
	}

	ip = r.Header.Get("X-Real-IP")
	if net.ParseIP(ip) != nil {
		return ip, nil
	}

	ip, _, err := net.SplitHostPort(r.RemoteAddr)
	if err != nil {
		return "", err
	}

	if net.ParseIP(ip) != nil {
		return ip, nil
	}

	return "", errors.New("no valid ip found")
}

func GetMacAddr() ([]string, error) {
	netInterfaces, err := net.Interfaces()
	if err != nil {
		return nil, err
	}

	var macAddrs []string
	for _, netInterface := range netInterfaces {
		macAddr := netInterface.HardwareAddr.String()
		if len(macAddr) == 0 {
			continue
		}

		macAddrs = append(macAddrs, macAddr)
	}

	return macAddrs, nil
}

// 生成唯一id
func UniqueId() string {
	id := xid.New()
	return id.String()
}

const charset = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"

var GenerateUniqueIDCount int64

func GenerateUniqueID(length uint) string {
	atomic.AddInt64(&GenerateUniqueIDCount, 1)
	var (
		now    = time.Now().UnixNano() + GenerateUniqueIDCount
		result = make([]byte, length)
		chars  = fmt.Sprintf("%d", now) + charset
	)
	rand.Seed(now)
	for i := range result {
		result[i] = chars[rand.Intn(len(chars))]
	}

	return string(result)
}

// md5 字符串加密
func Md5String(str string) string {
	h := md5.New()
	h.Write([]byte(str))
	cipherStr := h.Sum(nil)

	return hex.EncodeToString(cipherStr)
}

func MapKeys[K comparable, V any](m map[K]V) []K {
	var keys []K
	for k := range m {
		keys = append(keys, k)
	}
	return keys
}

func MapValues[K comparable, V any](m map[K]V) []V {
	var values []V
	for _, v := range m {
		values = append(values, v)
	}
	return values
}

func SliceToMap[T any, K comparable](records []T, call func(item T) K) map[K]T {
	var maps = make(map[K]T)
	for _, item := range records {
		maps[call(item)] = item
	}

	return maps
}

func StructToMap(in interface{}, tagName string, call func(key string, val interface{}) interface{}) map[string]interface{} {
	var (
		data = make(map[string]interface{})
		v    = reflect.ValueOf(in)
	)

	if v.Kind() == reflect.Ptr {
		v = v.Elem()
	}

	if v.Kind() != reflect.Struct {
		return nil
	}

	var t = v.Type()
	for i := 0; i < t.NumField(); i++ {
		field := t.Field(i)
		var val = field.Tag.Get(tagName)
		if val != "" && val != "-" {
			val = strings.Split(val, ",")[0]
			data[val] = v.Field(i).Interface()

			if call != nil {
				data[val] = call(val, data[val])
			} else {
				if v, ok := data[val].(sql.NullString); ok {
					data[val] = v.String
					continue
				}
			}
		}
	}

	return data
}

func IsArray(arr interface{}) bool {
	var v = reflect.ValueOf(arr)
	return v.Kind() == reflect.Array || (v.Kind() == reflect.Slice && v.Len() > 0)
}

func ConvStringToType(input string, data interface{}) error {
	return JSONUnmarshal([]byte(input), data)
}

func ToString(input interface{}) (string, error) {
	b, err := JSONMarshal(input)
	if err != nil {
		return "", nil
	}

	return string(b), nil
}

func ToInterface(data interface{}) interface{} {
	return data
}

func ModifyField(obj interface{}, fieldName string, newValue interface{}) error {
	var v = reflect.ValueOf(obj)
	if v.Kind() != reflect.Ptr || v.IsNil() {
		return fmt.Errorf("obj must be a non-nil pointer")
	}

	v = v.Elem()
	if v.Kind() != reflect.Struct {
		return fmt.Errorf("obj must point to a struct")
	}

	var field = v.FieldByName(fieldName)
	if !field.IsValid() {
		return fmt.Errorf("no such field: %s in obj", fieldName)
	}

	if !field.CanSet() {
		return fmt.Errorf("cannot set field %s", fieldName)
	}

	var newVal = reflect.ValueOf(newValue)
	if field.Type() != newVal.Type() {
		return fmt.Errorf("provided value type doesn't match obj field type")
	}

	field.Set(newVal)
	return nil
}

func Trim(str string) string {
	if str == "" {
		return ""
	}

	return regexp.MustCompile("^\\s*|\\s*$").ReplaceAllString(str, "")
}

func HttpGet(url string) ([]byte, error) {
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

func ConvInterface(input, val interface{}) error {
	b, err := JSONMarshal(input)
	if err != nil {
		return err
	}

	return JSONUnmarshal(b, val)
}

func PageOffset(page, limit int) int {
	if page <= 0 {
		return 0
	}

	page = page - 1

	if limit == 0 {
		limit = 20
	}

	return page * limit
}

func TickerWithDuration(duration int64, call func()) *time.Ticker {
	var (
		now    = NewTimer().Now()
		ticker = time.NewTicker(time.Second * 1)
	)
	go func() {
		for t := range ticker.C {
			if t.Unix()-now >= duration {
				ticker.Stop()
				call()
				break
			}
		}
	}()

	return ticker
}

func GetFieldValue(data interface{}, name string) (interface{}, error) {
	var v = reflect.ValueOf(data)
	if v.Kind() == reflect.Ptr {
		v = v.Elem()
	}

	var field = v.FieldByName(name)
	if !field.IsValid() {
		return nil, fmt.Errorf("no such field: %s", name)
	}

	return field.Interface(), nil
}

func ServerNode() string {
	var host = GetEnvDefault("SKEYEVSS_INTERNAL_IP", "")
	if host == "" {
		return ""
	}

	var (
		arr  = strings.Split(host, ".")
		node = ""
	)
	if len(arr) == 4 {
		node = strings.Join([]string{arr[2], arr[3]}, "-")
	}

	if host == "localhost" {
		return "localhost"
	}

	return node
}

func IsSimpleType(in interface{}) (bool, reflect.Type) {
	var t = reflect.TypeOf(in)

	if t.Kind() == reflect.Ptr {
		t = t.Elem()
	}

	switch t.Kind() {
	case reflect.String,
		reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64,
		reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64,
		reflect.Float32, reflect.Float64,
		reflect.Bool:
		return true, t

	default:
		return false, t
	}
}

// func IsSimpleType(in interface{}) bool {
// 	var t = reflect.ValueOf(in)
//
// 	if t.Kind() == reflect.Ptr {
// 		in = t.Interface()
// 	}
//
// 	switch in.(type) {
// 	case string,
// 		int, int8, int16, int32, int64,
// 		uint, uint8, uint16, uint32, uint64,
// 		float32, float64,
// 		bool:
// 		return true
//
// 	default:
// 		return false
// 	}
//
// 	// if _, ok := in.(bool); ok {
// 	// 	return true
// 	// }
// 	// if _, ok := in.(string); ok {
// 	// 	return true
// 	// }
// 	// if _, ok := in.(float32); ok {
// 	// 	return true
// 	// }
// 	// if _, ok := in.(float64); ok {
// 	// 	return true
// 	// }
// 	// if _, ok := in.(int); ok {
// 	// 	return true
// 	// }
// 	// if _, ok := in.(int8); ok {
// 	// 	return true
// 	// }
// 	// if _, ok := in.(int16); ok {
// 	// 	return true
// 	// }
// 	// if _, ok := in.(int32); ok {
// 	// 	return true
// 	// }
// 	// if _, ok := in.(int64); ok {
// 	// 	return true
// 	// }
// 	// if _, ok := in.(uint); ok {
// 	// 	return true
// 	// }
// 	// if _, ok := in.(uint8); ok {
// 	// 	return true
// 	// }
// 	// if _, ok := in.(uint16); ok {
// 	// 	return true
// 	// }
// 	// if _, ok := in.(uint32); ok {
// 	// 	return true
// 	// }
// 	// if _, ok := in.(uint64); ok {
// 	// 	return true
// 	// }
// 	//
// 	// return false
// }

func OffsetCall(limit, total int, call func(start, end int)) {
	if limit >= total {
		call(0, total)
		return
	}

	var counter = int(math.Ceil(float64(total) / float64(limit)))
	for i := 0; i < counter; i++ {
		var (
			start = i * limit
			end   = start + limit
		)
		if start >= total-1 {
			start = total - 1
		}

		if end >= total {
			end = total
		}

		call(start, end)
	}
}

// ---------------------------------------- functions

func DataToString(input interface{}, def string) string {
	b, err := JSONMarshal(input)
	if err != nil {
		return def
	}

	return string(b)
}

func MapStructureHook(from reflect.Type, to reflect.Type, data interface{}) (interface{}, error) {
	if from.Kind() == reflect.String && to.Kind() == reflect.Uint {
		if val, ok := data.(json.Number); ok {
			v, err := val.Int64()
			if err != nil {
				return nil, err
			}

			return uint(v), nil
		}
	}

	if from.Kind() == reflect.String && to.Kind() == reflect.Uint64 {
		if val, ok := data.(json.Number); ok {
			v, err := val.Int64()
			if err != nil {
				return nil, err
			}

			return uint64(v), nil
		}
	}

	return data, nil
}

// ---------------------------------- conv

func ConvBytes[T any](data []byte) (T, error) {
	var zero T
	if len(data) == 0 {
		return zero, errors.New("empty input data")
	}

	// 获取类型信息
	var (
		target     T
		targetType = reflect.TypeOf(target)
	)

	switch targetType.Kind() {
	case reflect.String:
		return any(string(data)).(T), nil

	case reflect.Bool:
		b, err := strconv.ParseBool(string(data))
		if err != nil {
			return zero, fmt.Errorf("bool conversion failed: %v", err)
		}
		return any(b).(T), nil

	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		// 特殊处理 time.Time 类型
		if _, ok := any(target).(time.Time); ok {
			t, err := parseTime(data)
			if err != nil {
				return zero, err
			}
			return any(t).(T), nil
		}

		bits := targetType.Bits()
		intVal, err := strconv.ParseInt(string(data), 10, bits)
		if err != nil {
			return zero, fmt.Errorf("integer conversion failed: %v", err)
		}

		switch bits {
		case 8:
			return any(int8(intVal)).(T), nil
		case 16:
			return any(int16(intVal)).(T), nil
		case 32:
			return any(int32(intVal)).(T), nil
		default:
			return any(intVal).(T), nil
		}

	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
		var bits = targetType.Bits()
		uintVal, err := strconv.ParseUint(string(data), 10, bits)
		if err != nil {
			return zero, fmt.Errorf("unsigned integer conversion failed: %v", err)
		}

		switch bits {
		case 8:
			return any(uint8(uintVal)).(T), nil
		case 16:
			return any(uint16(uintVal)).(T), nil
		case 32:
			return any(uint32(uintVal)).(T), nil
		default:
			return any(uintVal).(T), nil
		}

	case reflect.Float32, reflect.Float64:
		var bits = targetType.Bits()
		floatVal, err := strconv.ParseFloat(string(data), bits)
		if err != nil {
			return zero, fmt.Errorf("float conversion failed: %v", err)
		}

		if bits == 32 {
			return any(float32(floatVal)).(T), nil
		}
		return any(floatVal).(T), nil

	default:
		if err := JSONUnmarshal(data, &target); err != nil {
			return zero, fmt.Errorf("JSON unmarshal failed: %v", err)
		}

		return target, nil
	}
}

// parseTime 尝试解析常见时间格式
func parseTime(data []byte) (time.Time, error) {
	formats := []string{
		time.RFC3339,
		"2006-01-02 15:04:05",
		"2006-01-02",
		time.RFC1123,
	}

	strData := string(data)
	for _, format := range formats {
		t, err := time.Parse(format, strData)
		if err == nil {
			return t, nil
		}
	}

	return time.Time{}, errors.New("time conversion failed - unrecognized format")
}

func PrintJson(data interface{}) {
	jj, _ := json.Marshal(data)
	var str bytes.Buffer
	_ = json.Indent(&str, jj, "", "    ")
	fmt.Printf("\n format: %+v \n", str.String())
}

func BoolToByte(val bool) []byte {
	if val {
		return []byte{1}
	}

	return []byte{0}
}

func ByteToBool(val []byte) bool {
	if len(val) == 1 {
		return val[0] == 1
	}

	return false
}

func MergeMaps[K comparable, V any](maps ...map[K]V) map[K]V {
	var records = make(map[K]V)
	for _, m := range maps {
		for k, v := range m {
			records[k] = v
		}
	}
	return records
}

func RoundFloat(f float64, precision int) float64 {
	var shift = math.Pow10(precision)
	return math.Round(f*shift) / shift
}

func FormatDurationPrecise(d time.Duration, precision int) string {
	switch {
	case d < time.Microsecond:
		return fmt.Sprintf("%.*fns", precision, float64(d.Nanoseconds()))
	case d < time.Millisecond:
		return fmt.Sprintf("%.*fµs", precision, float64(d.Microseconds()))
	case d < time.Second:
		return fmt.Sprintf("%.*fms", precision, float64(d.Milliseconds()))
	case d < time.Minute:
		return fmt.Sprintf("%.*fs", precision, d.Seconds())
	case d < time.Hour:
		return fmt.Sprintf("%.*fm", precision, d.Minutes())
	default:
		return fmt.Sprintf("%.*fh", precision, d.Hours())
	}
}

const (
	B  = 1
	KB = 1024 * B
	MB = 1024 * KB
	GB = 1024 * MB
	TB = 1024 * GB
	PB = 1024 * TB
)

func ByteSize(bytes uint64) string {
	switch {
	case bytes >= PB:
		return fmt.Sprintf("%.2fPB", float64(bytes)/PB)
	case bytes >= TB:
		return fmt.Sprintf("%.2fTB", float64(bytes)/TB)
	case bytes >= GB:
		return fmt.Sprintf("%.2fGB", float64(bytes)/GB)
	case bytes >= MB:
		return fmt.Sprintf("%.2fMB", float64(bytes)/MB)
	case bytes >= KB:
		return fmt.Sprintf("%.2fKB", float64(bytes)/KB)
	default:
		return fmt.Sprintf("%dB", bytes)
	}
}

func CheckURLExists(url string) (bool, error) {
	var client = &http.Client{Timeout: 10 * time.Second}
	resp, err := client.Head(url)
	if err != nil {
		return false, err
	}
	defer resp.Body.Close()

	if resp.StatusCode >= 200 && resp.StatusCode < 400 {
		return true, nil
	}

	return false, nil
}

// @Title        结构体转message
// @Description  main
// @Create       yirl 2025/3/24 14:26

package main

import (
	"reflect"
	"strconv"

	"skeyevss/core/repositories/models/roles"
)

func main() {
	var (
		p = new(roles.Roles)
		t = reflect.TypeOf(p).Elem()
	)

	var message = `message RoleRow{`
	for i := 0; i < t.NumField(); i++ {
		var (
			field     = t.Field(i)
			fieldType = field.Type
			name      = field.Tag.Get("json")
			Type      string
		)
		if name == "" {
			continue
		}

		switch fieldType.Kind() {
		case reflect.Uint, reflect.Uint32:
			Type = "uint32"

		case reflect.Uint64:
			Type = "uint64"

		case reflect.Int32:
			Type = "int32"

		case reflect.Int, reflect.Int64:
			Type = "int64"

		case reflect.Float64, reflect.Float32:
			Type = "float"

		case reflect.Bool:
			Type = "bool"

		default:
			Type = "string"
		}

		message += "\n	" + Type + " " + name + " = " + strconv.Itoa(i+1) + ";"
	}

	message += "\n}"

	println(message)
}

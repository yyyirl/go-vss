package tps

import (
	"fmt"
	"runtime"
)

type Number interface {
	int | int8 | int32 | int64 | uint | uint8 | uint32 | uint64 | float64 | float32
}

type Simple interface {
	int | int8 | int32 | int64 | uint | uint8 | uint32 | uint64 | float64 | float32 | string
}

type NumberDerived interface {
	~int | ~int8 | ~int32 | ~int64 | ~uint | ~uint8 | ~uint32 | ~uint64
}

type MapValue interface {
	int | int8 | int32 | int64 | uint | uint8 | uint32 | uint64 | float64 | float32 | map[string]interface{} | map[string]int | map[string]int64 | map[string]map[string]interface{} | map[string]map[int]interface{} | interface{}
}

type SliceCompareValue interface {
	int | int8 | int32 | int64 | uint | uint8 | uint32 | uint64 | float64 | float32 | string
}

type MI = map[string]interface{}

type FormFile struct {
	B        []byte
	Ext      string
	FileName string
}

type EmailDecrypt struct {
	Expire int64  `json:"expire"`
	Type   string `json:"type"`
	Email  string `json:"email"`
	Code   string `json:"code"`
}

// ------------------------------------------ language

type CharLang struct {
	ZHSimplified  string `json:"zhSimplified" db:"zhSimplified"`
	ZHTraditional string `json:"zhTraditional" db:"zhTraditional"`
	EN            string `json:"en" db:"en"`
}

type Lang struct {
	ID      uint      `json:"id"`
	Content *CharLang `json:"content"`
}

// ------------------------------------------ user

type (
	TokenItem struct {
		Userinfo MI    `json:"userinfo"`
		Expire   int64 `json:"expire"`
	}
)

// ------------------------------------------ options

type OptionItem struct {
	Title    string        `json:"title"`
	Value    any           `json:"value"`
	Disabled bool          `json:"disabled,omitempty,optional"`
	Raw      interface{}   `json:"raw,omitempty,optional"`
	Children []*OptionItem `json:"children,omitempty,optional"`
}

// ------------------------------------------ error

type XError struct {
	Message string
}

func NewErr(message string) *XError {
	pc, file, line, status := runtime.Caller(1)
	return &XError{Message: fmt.Sprintf("%s, file: %s:%d:%t\nfuncName:%s", message, file, line, status, runtime.FuncForPC(pc).Name())}
}

func NewErrWithSkip(skip int, message string) *XError {
	pc, file, line, status := runtime.Caller(skip)
	return &XError{Message: fmt.Sprintf("%s, file: %s:%d:%t\nfuncName:%s", message, file, line, status, runtime.FuncForPC(pc).Name())}
}

func (e *XError) Error() string {
	return e.Message
}

type DownloadType struct {
	Success   bool   `json:"success"`
	Progress  string `json:"progress"`
	Url       string `json:"url"`
	LocalFile string `json:"focalFile"`
	Error     string `json:"error"`
	Cancel    bool   `json:"cancel"`
}

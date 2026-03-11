/**
 * @Author:         yi
 * @Description:    version
 * @Version:        1.0.0
 * @Date:           2022/12/19 12:01
 */
package functions

import (
	"strings"
)

const (
	// 不一致
	VUEQ int64 = iota - 1 // -1
	// 等于
	VEQ // 0
	// 大于
	VGT // 1
	// 小于
	VLT // 2
)

type Version struct{}

func NewVersion() *Version {
	return &Version{}
}

func (this *Version) Compare(verA, verB string) int64 {
	var (
		verStrArrA = this.splitStrByNet(verA)
		verStrArrB = this.splitStrByNet(verB)
		lenStrA    = len(verStrArrA)
		lenStrB    = len(verStrArrB)
	)

	if lenStrA != lenStrB {
		return VUEQ
	}

	return this.compareArrStr(verStrArrA, verStrArrB)
}

// 比较版本号字符串数组
func (this *Version) compareArrStr(verA, verB []string) int64 {
	for index := range verA {
		var littleResult = this.compareLittleVer(verA[index], verB[index])
		if littleResult != VEQ {
			return littleResult
		}
	}

	return VEQ
}

// 比较小版本号字符串
func (this *Version) compareLittleVer(verA, verB string) int64 {
	var (
		lengthA = len(verA)
		lengthB = len(verB)
	)
	if lengthB > lengthA {
		verA = verA + strings.Repeat("0", lengthB-lengthA)
	}

	if lengthB < lengthA {
		if lengthB-lengthA <= 0 {
			verB = verB + strings.Repeat("0", 0)
		} else {
			verB = verB + strings.Repeat("0", lengthB-lengthA)
		}
	}

	var (
		bytesA = []byte(verA)
		bytesB = []byte(verB)
		lenA   = len(bytesA)
		lenB   = len(bytesB)
	)

	if lenA > lenB {
		return VGT
	}

	if lenA < lenB {
		return VLT
	}

	// 如果长度相等则按byte位进行比较
	return this.compareByBytes(bytesA, bytesB)
}

// 按byte位进行比较小版本号
func (this *Version) compareByBytes(verA, verB []byte) int64 {
	for index := range verA {
		if verA[index] > verB[index] {
			return VGT
		}

		if verA[index] < verB[index] {
			return VLT
		}
	}

	return VEQ
}

// 按“.”分割版本号为小版本号的字符串数组
func (this *Version) splitStrByNet(strV string) []string {
	return strings.Split(strV, ".")
}

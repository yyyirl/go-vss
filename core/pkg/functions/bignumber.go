// @Title        bigint
// @Description  main
// @Create       yiyiyi 2025/9/29 11:03

package functions

import (
	"fmt"
	"math/big"
)

type BigNumber struct {
	value *big.Int
}

func NewBigNumber(numStr string) (*BigNumber, error) {
	num := new(big.Int)
	_, ok := num.SetString(numStr, 10)
	if !ok {
		return nil, fmt.Errorf("invalid number string: %s", numStr)
	}
	return &BigNumber{value: num}, nil
}

// 比较方法
func (bn *BigNumber) Compare(other *BigNumber) int {
	return bn.value.Cmp(other.value)
}

func (bn *BigNumber) GreaterThan(other *BigNumber) bool {
	return bn.value.Cmp(other.value) > 0
}

func (bn *BigNumber) LessThan(other *BigNumber) bool {
	return bn.value.Cmp(other.value) < 0
}

func (bn *BigNumber) EqualTo(other *BigNumber) bool {
	return bn.value.Cmp(other.value) == 0
}

// 数学运算
func (bn *BigNumber) AddOne() *BigNumber {
	one := big.NewInt(1)
	newValue := new(big.Int).Add(bn.value, one)
	return &BigNumber{value: newValue}
}

func (bn *BigNumber) Add(n int64) *BigNumber {
	newValue := new(big.Int).Add(bn.value, big.NewInt(n))
	return &BigNumber{value: newValue}
}

func (bn *BigNumber) String() string {
	return bn.value.String()
}

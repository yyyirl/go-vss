/**
 * @Author:         yi
 * @Description:    encryption
 * @Version:        1.0.0
 * @Date:           2022/10/11 11:25
 */
package functions

import (
	"bytes"
	"crypto/aes"
	"crypto/cipher"
	"encoding/base64"
	"errors"
	"regexp"
	"time"

	"golang.org/x/crypto/bcrypt"

	"skeyevss/core/tps"
)

type Crypto struct {
	Key []byte
}

// NewCrypto 初始化
func NewCrypto(key []byte) *Crypto {
	return &Crypto{Key: key}
}

// padding 填充明文
func (t *Crypto) padding(src []byte, blockSize int) []byte {
	padNum := blockSize - len(src)%blockSize
	return append(src, bytes.Repeat([]byte{byte(padNum)}, padNum)...)
}

// unPadding 去除填充数据
func (t *Crypto) unPadding(src []byte) []byte {
	if len(src) <= 0 {
		return src
	}

	var (
		n  = len(src)
		n1 = n - int(src[n-1])
	)
	if n1 > len(src) || n1 <= 0 {
		return src
	}

	return src[:n1]
}

// encryptAES AES加密
func (t *Crypto) encryptAES(src, key []byte) ([]byte, error) {
	block, _ := aes.NewCipher(key)
	src = t.padding(src, block.BlockSize())

	if len(src)%block.BlockSize() != 0 {
		return nil, errors.New("illegality plain text")
	}

	cipher.NewCBCEncrypter(block, key).CryptBlocks(src, src)
	return src, nil
}

// decryptAES AES解密
func (t *Crypto) decryptAES(src, key []byte) ([]byte, error) {
	block, err := aes.NewCipher(key)
	if err != nil {
		return nil, err
	}

	if len(src)%block.BlockSize() != 0 {
		return nil, errors.New("illegality plain text")
	}

	cipher.NewCBCDecrypter(block, key).CryptBlocks(src, src)
	return t.unPadding(src), nil
}

// Encrypt 加密
func (t *Crypto) Encrypt(pass []byte) (string, error) {
	p, err := t.encryptAES(pass, t.Key)
	if err != nil {
		return "", err
	}

	return base64.StdEncoding.EncodeToString(p), nil
}

// 解密 Decrypt
func (t *Crypto) Decrypt(cipherText string) (string, error) {
	bytesPass, err := base64.StdEncoding.DecodeString(cipherText)
	if err != nil {
		return "", err
	}

	res, err := t.decryptAES(bytesPass, t.Key)
	if err != nil {
		return "", err
	}

	return string(res), nil
}

/*
replace pkg/encryption => ../../pkg/encryption
require pkg/encryption v0.0.0

s, err := encryption.NewCrypto([]byte("lbV2lcq1mLoXL3GV")).Encrypt([]byte("我来测试"))
if err != nil {
	panic(err)
}

aaaaa, err := encryption.NewCrypto([]byte("lbV2lcq1mLoXL3GV")).Decrypt(s)
if err != nil {
	panic(err)
}
println(aaaaa)
*/

// 验证密码
func ValidatePwd(hash string, password string) bool {
	if err := bcrypt.CompareHashAndPassword([]byte(hash), []byte(password)); err != nil {
		return false
	}

	return true
}

// 加密密码
func GeneratePwd(password string) (string, error) {
	hash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.MinCost)
	if err != nil {
		return "", err
	}
	return string(hash), nil
}

func IsBcryptHash(str string) bool {
	return len(str) == 60 && regexp.MustCompile(`^\$2[aby]\$.{56}`).MatchString(str)
}

func MakeTokenVASE(key string, expire time.Duration, data tps.TokenItem) (string, error) {
	data.Expire = int64(expire)
	b, err := JSONMarshal(data)
	if err != nil {
		return "", err
	}

	encrypt, err := NewCrypto([]byte(key)).Encrypt(b)
	if err != nil {
		return "", err
	}

	return encrypt, nil
}

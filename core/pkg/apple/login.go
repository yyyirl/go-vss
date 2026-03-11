/**
 * @Author:         yi
 * @Description:    login
 * @Version:        1.0.0
 * @Date:           2022/10/20 22:13
 */
package apple

import (
	"context"
	"crypto/rsa"
	"encoding/base64"
	"errors"
	"fmt"
	"math/big"
	"runtime/debug"
	"strings"

	"github.com/Timothylock/go-signin-with-apple/apple"
	"github.com/golang-jwt/jwt"

	"skeyevss/core/pkg/functions"
)

// https://www.cnblogs.com/goloving/p/14334798.html
const (
	PublicKeyReqUrl = "https://appleid.apple.com/auth/keys"
	// Url             = "https://appleid.apple.com"
	// applicationClientId = "com.***.***" com.fio.LongCam
)

type JwtClaims struct {
	jwt.StandardClaims
}

type JwtHeader struct {
	Kid string `json:"kid"`
	Alg string `json:"alg"`
}

type JwtKeys struct {
	Kty string `json:"kty"`
	Kid string `json:"kid"`
	Use string `json:"use"`
	Alg string `json:"alg"`
	N   string `json:"n"`
	E   string `json:"e"`
}

func AuthToken(applePrivateKey []byte, teamId, clientId, keyId, code string) (string, error) {
	secret, err := apple.GenerateClientSecret(
		string(applePrivateKey),
		teamId,
		clientId,
		keyId,
	)
	client := apple.New()
	vReq := apple.AppValidationTokenRequest{
		ClientID:     clientId,
		ClientSecret: secret,
		Code:         code,
	}

	var resp apple.ValidationResponse
	err = client.VerifyAppToken(context.Background(), vReq, &resp)
	if err != nil {
		return "", err
	}

	if resp.Error != "" {
		return "", err
	}

	// Get the unique user ID
	/*unique, err := apple.GetUniqueID(resp.IDToken)
	if err != nil {
		return nil, err
	}

	// Get the email
	claim, err := apple.GetClaims(resp.IDToken)
	if err != nil {
		return nil, err
	}
	//email := (*claim)["email"]
	//emailVerified := (*claim)["email_verified"]
	//isPrivateEmail := (*claim)["is_private_email"]
	// Voila!


	logger.Log("info", "resp.IDToken:%+v", resp.IDToken)
	logger.Log("info", "claim:%+v", claim)
	logger.Log("info", "unique:%+v", unique)*/

	return resp.IDToken, nil
}

// 认证客户端传递过来的token是否有效
func VerifyIdentityToken(cliToken string) (*JwtClaims, error) {
	defer func() {
		if err := recover(); err != nil {
			functions.LogError(fmt.Sprintf("apple token panic!!!\n %s\ntoken: %s\n \nStack: %s", functions.Caller(4), cliToken, string(debug.Stack())))
		}
	}()

	if cliToken == "" {
		return nil, errors.New("token 不能为空")
	}

	// 数据由 头部、载荷、签名 三部分组成
	cliTokenArr := strings.Split(cliToken, ".")
	if len(cliTokenArr) < 3 {
		return nil, errors.New("cliToken Split err")
	}

	// 解析cliToken的header获取kid
	cliHeader, err := jwt.DecodeSegment(cliTokenArr[0])
	if err != nil {
		return nil, err
	}

	var jHeader JwtHeader
	err = functions.JSONUnmarshal(cliHeader, &jHeader)
	if err != nil {
		return nil, err
	}

	// 效验pubKey 及 token
	token, err := jwt.ParseWithClaims(cliToken, &JwtClaims{}, func(token *jwt.Token) (interface{}, error) {
		return GetRSAPublicKey(jHeader.Kid), nil
	})

	if err != nil {
		return nil, err
	}

	// 信息验证
	claims, ok := token.Claims.(*JwtClaims)
	if ok && token.Valid {
		/*if claims.Issuer != Url || claims.Audience != applicationClientId || claims.Subject != clientId {
			return nil, errors.New("verify token info fail, info is not match")
		}*/
		return claims, nil
	}

	return nil, errors.New("token claims parse fail")
}

// 向苹果服务器获取解密signature所需要用的publicKey
func GetRSAPublicKey(kid string) *rsa.PublicKey {
	response, err := functions.HttpGet(PublicKeyReqUrl)
	if err != nil {
		return nil
	}

	var jKeys map[string][]JwtKeys
	err = functions.JSONUnmarshal(response, &jKeys)
	if err != nil {
		return nil
	}

	// 获取验证所需的公钥
	var pubKey rsa.PublicKey
	// 通过cliHeader的kid比对获取n和e值 构造公钥
	for _, data := range jKeys {
		for _, val := range data {
			if val.Kid == kid {
				nBin, _ := base64.RawURLEncoding.DecodeString(val.N)
				nData := new(big.Int).SetBytes(nBin)

				eBin, _ := base64.RawURLEncoding.DecodeString(val.E)
				eData := new(big.Int).SetBytes(eBin)

				pubKey.N = nData
				pubKey.E = int(eData.Uint64())
				break
			}
		}
	}

	if pubKey.E <= 0 {
		return nil
	}

	return &pubKey
}

// @Title        main
// @Description  cert
// @Create       yirl 2025/3/18 17:41

package cert

import (
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"crypto/x509/pkix"
	"encoding/pem"
	"fmt"
	"math/big"
	"os"
	"time"
)

type Certificate struct{}

func NewCertificate() *Certificate {
	return &Certificate{}
}

func (c *Certificate) Make(certFile, keyFile string) error {
	// 检查证书文件是否存在
	if _, err := os.Stat(certFile); os.IsNotExist(err) {
		return c.GenerateSelfSignedCert(certFile, keyFile)
	}

	// 读取现有证书
	certPEM, err := os.ReadFile(certFile)
	if err != nil {
		return err
	}

	// 解码证书
	block, _ := pem.Decode(certPEM)
	if block == nil || block.Type != "CERTIFICATE" {
		return fmt.Errorf("failed to decode certificate")
	}

	// 解析证书
	cert, err := x509.ParseCertificate(block.Bytes)
	if err != nil {
		return err
	}

	// 检查证书是否过期
	if time.Now().After(cert.NotAfter) {
		return c.GenerateSelfSignedCert(certFile, keyFile)
	}

	return nil
}

// GenerateSelfSignedCert 生成自签名证书
func (c *Certificate) GenerateSelfSignedCert(certFile, keyFile string) error {
	privateKey, err := c.GeneratePrivateKey()
	if err != nil {
		return err
	}

	var (
		notBefore = time.Now()
		notAfter  = notBefore.Add(365 * 24 * time.Hour) // 有效期一年
	)
	serialNumber, err := rand.Int(rand.Reader, big.NewInt(1<<62))
	if err != nil {
		return err
	}

	// 创建自签名证书模板
	var (
		template = x509.Certificate{
			SerialNumber: serialNumber,
			Subject: pkix.Name{
				Organization: []string{"skeyevss"},
			},
			NotBefore:             notBefore,
			NotAfter:              notAfter,
			KeyUsage:              x509.KeyUsageKeyEncipherment | x509.KeyUsageDigitalSignature,
			ExtKeyUsage:           []x509.ExtKeyUsage{x509.ExtKeyUsageServerAuth},
			BasicConstraintsValid: true,
		}
	)
	derBytes, err := x509.CreateCertificate(rand.Reader, &template, &template, &privateKey.PublicKey, privateKey)
	if err != nil {
		return err
	}

	certOut, err := os.Create(certFile)
	if err != nil {
		return err
	}

	if err := pem.Encode(certOut, &pem.Block{Type: "CERTIFICATE", Bytes: derBytes}); err != nil {
		return err
	}
	if err := certOut.Close(); err != nil {
		return err
	}

	keyOut, err := os.Create(keyFile)
	if err != nil {
		return err
	}

	if err := pem.Encode(keyOut, &pem.Block{Type: "PRIVATE KEY", Bytes: x509.MarshalPKCS1PrivateKey(privateKey)}); err != nil {
		return err
	}
	return keyOut.Close()
}

// GeneratePrivateKey 生成私钥
func (c *Certificate) GeneratePrivateKey() (*rsa.PrivateKey, error) {
	return rsa.GenerateKey(rand.Reader, 2048)
}

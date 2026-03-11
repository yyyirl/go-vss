package pkg

import (
	"fmt"

	"skeyevss/core/pkg/functions"
)

type (
	UtilsConfig struct {
		Expire uint64
	}

	Utils struct {
		config *UtilsConfig
	}

	ConnAuthorization struct {
		ConnType string `json:"connType"`
		Key      string `json:"key"`
		ID       uint64 `json:"id"`
		Expire   uint64 `json:"expire"`
	}
)

func NewUtils(config *UtilsConfig) *Utils {
	return &Utils{config: config}
}

func (m *Utils) MakeClientId(connType string, id uint64) string {
	if id <= 0 {
		return fmt.Sprintf("%s:%s", connType, functions.UniqueId())
	}

	return fmt.Sprintf("%s:%d", connType, id)
}

func (m *Utils) MakeConnAuthorization(key, connType string, id uint64) (string, error) {
	b, err := functions.JSONMarshal(&ConnAuthorization{
		ConnType: connType,
		Key:      key,
		ID:       id,
		Expire:   m.config.Expire,
	})
	if err != nil {
		return "", err
	}

	encrypt, err := functions.NewCrypto([]byte(key)).Encrypt(b)
	if err != nil {
		return "", err
	}

	return encrypt, nil
}

func (m *Utils) VerifyConnAuthorization(key, connType, cipherText string) error {
	content, err := functions.NewCrypto([]byte(key)).Decrypt(cipherText)
	if err != nil {
		return err
	}

	var data ConnAuthorization
	if err := functions.JSONUnmarshal([]byte(content), &data); err != nil {
		return err
	}

	if data.ConnType != connType {
		return fmt.Errorf("connection type not match")
	}

	return nil
}

/**
 * @Author:         yi
 * @Description:    cloudkit
 * @Version:        1.0.0
 * @Date:           2022/7/18 10:30
 */
package functions

import (
	"bytes"
	"context"
	"crypto/ecdsa"
	"crypto/x509"
	"errors"
	"fmt"
	"os/exec"
	"reflect"
	"strings"

	"skeyevss/core/pkg/functions/sc"
	"skeyevss/core/pkg/icloud"
)

// https://github.com/lukasmalkmus/icloud-go
// openssl ecparam -name prime256v1 -genkey -noout -out eckey.pem
// openssl ec -in eckey.pem -pubout

type Cloudkit struct {
	keyPath,
	container,
	keyID string

	env icloud.Environment // icloud.Development
}

func NewCloudKit(keyPath, container, keyID string, env icloud.Environment) *Cloudkit {
	return &Cloudkit{
		keyPath,
		container,
		keyID,
		env,
	}
}

var cloudKitPriKey *ecdsa.PrivateKey

func (this *Cloudkit) getCloudKitPriKey() (*ecdsa.PrivateKey, error) {
	if cloudKitPriKey != nil {
		return cloudKitPriKey, nil
	}

	if exists, err := PathExists(this.keyPath); err != nil {
		return nil, err
	} else {
		if !exists {
			return nil, errors.New("文件不存在")
		}
	}

	b, err := exec.Command("openssl", "ec", "-outform", "der", "-in", this.keyPath).Output()
	if err != nil {
		return nil, err
	}

	cloudKitPriKey, err := x509.ParseECPrivateKey(b)
	if err != nil {
		return nil, err
	}

	return cloudKitPriKey, nil
}

func (this *Cloudkit) modifyClient() (*icloud.Client, error) {
	privateKey, err := this.getCloudKitPriKey()
	if err != nil {
		return nil, err
	}

	return icloud.NewClient(this.container, this.keyID, privateKey, this.env, "/records/modify")
}

func (this *Cloudkit) assetUploadClient() (*icloud.Client, error) {
	privateKey, err := this.getCloudKitPriKey()
	if err != nil {
		return nil, err
	}

	return icloud.NewClient(this.container, this.keyID, privateKey, this.env, "/assets/upload")
}

func (this *Cloudkit) recordsDeleteClient() (*icloud.Client, error) {
	privateKey, err := this.getCloudKitPriKey()
	if err != nil {
		return nil, err
	}

	return icloud.NewClient(this.container, this.keyID, privateKey, this.env, "/records/delete")
}

func (this *Cloudkit) queryClient() (*icloud.Client, error) {
	privateKey, err := this.getCloudKitPriKey()
	if err != nil {
		return nil, err
	}

	return icloud.NewClient(this.container, this.keyID, privateKey, this.env, "/records/query")
}

func (this *Cloudkit) Send(data []icloud.RecordOperation) (*icloud.RecordsResponse, error) {
	client, err := this.modifyClient()
	if err != nil {
		return nil, err
	}

	return client.Records.Modify(context.Background(), icloud.Public, icloud.RecordsRequest{
		Operations: data,
	})
}

func (this *Cloudkit) ToCloudkitField(_data interface{}) icloud.Fields {
	var data []icloud.Field

	if v, ok := _data.(map[string]interface{}); ok {
		for key, value := range v {
			data = append(data, icloud.Field{
				Name:  key,
				Value: value,
			})
		}

		return data
	}

	var v = reflect.ValueOf(_data)
	if v.Kind() == reflect.Ptr {
		v = v.Elem()
	}

	var t = v.Type()
	for i := 0; i < t.NumField(); i++ {
		field := t.Field(i)
		var val = field.Tag.Get("db")
		if val != "" {
			data = append(data)
			val = strings.Split(val, ",")[0]
			var _v = v.Field(i).Interface()
			if _v == nil {
				continue
			}

			data = append(data, icloud.Field{
				Name:  strings.Split(val, ",")[0],
				Value: _v,
			})
		}
	}

	return data
}

func (this *Cloudkit) FindUniqueId(data []icloud.Field) string {
	for _, item := range data {
		if item.Name == "uniqueId" {
			return item.Value.(string)
		}
	}

	return UniqueId()
}

func (this *Cloudkit) ToCloudkitRecords(Type string, _data []interface{}, optType icloud.OperationType) []icloud.RecordOperation {
	var data []icloud.RecordOperation
	for _, item := range _data {
		var value = this.ToCloudkitField(item)
		data = append(data, icloud.RecordOperation{
			Type: optType,
			Record: icloud.Record{
				Type:   Type,
				Name:   this.FindUniqueId(value),
				Fields: value,
			},
		})
	}

	return data
}

func (this *Cloudkit) FileUploadCert(data []*icloud.AssetUploadTokenOperation) (*icloud.AssetUploadTokensResponse, error) {
	client, err := this.assetUploadClient()
	if err != nil {
		return nil, err
	}

	return client.Records.AssetUpload(context.Background(), icloud.Public, data)
}

func (this *Cloudkit) FileUpload(records []*icloud.AssetUploadTokenOperation) ([]*icloud.UploadRep, error) {
	res, err := this.FileUploadCert(records)
	if err != nil {
		return nil, err
	}

	var final []*icloud.UploadRep
	for _, item := range res.Tokens {
		var (
			cmdStr = fmt.Sprintf(
				`curl -X POST -H "Content-Type: multipart/form-data" -F "file=@%s" "%s"`,
				item.FieldName,
				item.Url,
			)
			cmd = sc.ExecCommand(cmdStr)
			stdout,
			stderr bytes.Buffer
		)

		cmd.Stdout = &stdout
		cmd.Stderr = &stderr
		if err = cmd.Run(); err != nil {
			return nil, err
		}

		var rep icloud.UploadRep
		if err = JSONUnmarshal(stdout.Bytes(), &rep); err != nil {
			return nil, err
		}

		rep.UniqueId = item.RecordName
		final = append(final, &rep)
	}

	return final, nil
}

func (this *Cloudkit) Query(req *icloud.QueryFiltersReq, value interface{}) error {
	client, err := this.queryClient()
	if err != nil {
		return err
	}

	return client.Records.Query(context.Background(), icloud.Public, req, value)
}

type CloudkitDeleteReq struct {
	Table      string `json:"table"`
	RecordName string `json:"recordName"`
}

func (this *Cloudkit) Delete(records []*CloudkitDeleteReq) error {
	if len(records) <= 0 {
		return errors.New("records不能为空")
	}

	var data []icloud.RecordOperation
	for _, item := range records {
		data = append(data, icloud.RecordOperation{
			Type: icloud.ForceDelete,
			Record: icloud.Record{
				Type: item.Table,
				Name: item.RecordName,
			},
		})
	}

	_, err := this.Send(data)
	return err
}

package sip

import (
	"bytes"
	"encoding/xml"
	"io"
	"unicode/utf8"

	"github.com/ghettovoice/gosip/sip"
	"golang.org/x/text/encoding/simplifiedchinese"
	"golang.org/x/text/transform"

	"skeyevss/core/app/sev/vss/internal/types"
)

type Parser[T types.IMessageReceive] struct{}

func NewParser[T types.IMessageReceive]() *Parser[T] {
	return &Parser[T]{}
}

func (p Parser[T]) ToData(req sip.Request) (*T, error) {
	var data T
	if err := p.XMLDecode([]byte(req.Body()), &data); err != nil {
		return nil, err
	}

	return &data, nil
}

func (p Parser[T]) XMLDecode(data []byte, v interface{}) error {
	if err := p.xmlDecode(data, v); err == nil {
		return nil
	}

	body, err := p.GbkToUtf8(data)
	if err != nil {
		return err
	}
	return p.xmlDecode(body, v)
}

func (p Parser[T]) xmlDecode(data []byte, v interface{}) error {
	decoder := xml.NewDecoder(bytes.NewReader(data))
	decoder.CharsetReader = func(charset string, input io.Reader) (io.Reader, error) {
		if utf8.Valid(data) {
			return input, nil
		}
		return simplifiedchinese.GB18030.NewDecoder().Reader(input), nil
	}
	return decoder.Decode(v)
}

func (p Parser[T]) GbkToUtf8(s []byte) ([]byte, error) {
	reader := transform.NewReader(bytes.NewReader(s), simplifiedchinese.GBK.NewDecoder())
	d, e := io.ReadAll(reader)
	if e != nil {
		return nil, e
	}
	return d, nil
}

package hlsdl

import (
	"crypto/aes"
	"crypto/cipher"
	"encoding/hex"
	"errors"
	"io/ioutil"
	"os"
	"strings"
)

func (hlsDl *HlsDl) Decrypt(segment *Segment) ([]byte, error) {

	file, err := os.Open(segment.Path)
	if err != nil {
		return nil, err
	}

	data, err := ioutil.ReadAll(file)
	if err != nil {
		return nil, err
	}

	if segment.Key != nil {
		key, iv, err := hlsDl.GetKey(segment)
		if err != nil {
			return nil, err
		}
		return decrypt(data, key, iv)
	}

	return data, nil
}

func decrypt(data, key, iv []byte) ([]byte, error) {
	block, err := aes.NewCipher(key)
	if err != nil {
		return nil, err
	}

	cbc := cipher.NewCBCDecrypter(block, iv)
	cbc.CryptBlocks(data, data)

	return zeroUnPadding(data), nil
}

func zeroUnPadding(origData []byte) []byte {
	length := len(origData)
	unpadding := int(origData[length-1])
	return origData[:(length - unpadding)]
}

func (hlsDl *HlsDl) GetKey(segment *Segment) (key []byte, iv []byte, err error) {
	res, err := hlsDl.client.Get(segment.Key.URI)
	if err != nil {
		return nil, nil, err
	}

	if res.StatusCode != 200 {
		return nil, nil, errors.New("Failed to get descryption key")
	}

	key, err = ioutil.ReadAll(res.Body)
	if err != nil {
		return nil, nil, err
	}

	if segment.Key.IV != "" {
		iv, err = hex.DecodeString(strings.TrimPrefix(segment.Key.IV, "0x"))
		if err != nil {
			return nil, nil, err
		}
	} else {
		iv = []byte{0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, byte(segment.SeqId)}
	}

	return
}

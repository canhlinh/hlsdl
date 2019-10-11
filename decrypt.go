package hlsdl

import (
	"bytes"
	"crypto/aes"
	"crypto/cipher"
	"errors"
	"io/ioutil"
	"log"
	"os"
)

func (hlsDl *HlsDl) Decrypt(segment *Segment) ([]byte, error) {

	file, err := os.Open(segment.Path)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	data, err := ioutil.ReadAll(file)
	if err != nil {
		return nil, err
	}

	if segment.Key != nil {
		key, iv, err := hlsDl.GetKey(segment)
		if err != nil {
			return nil, err
		}

		log.Println("Descrypting", len(key), len(iv))
		data, err = AES128Decrypt(data, key, iv)
	}

	syncByte := uint8(71) //0x47
	bLen := len(data)
	for j := 0; j < bLen; j++ {
		if data[j] == syncByte {
			data = data[j:]
			break
		}
	}

	return data, nil
}

func AES128Decrypt(crypted, key, iv []byte) ([]byte, error) {
	block, err := aes.NewCipher(key)
	if err != nil {
		return nil, err
	}
	blockSize := block.BlockSize()
	if len(iv) == 0 {
		iv = key
	}
	blockMode := cipher.NewCBCDecrypter(block, iv[:blockSize])
	origData := make([]byte, len(crypted))
	blockMode.CryptBlocks(origData, crypted)
	origData = pkcs5UnPadding(origData)
	return origData, nil
}

func pkcs5Padding(cipherText []byte, blockSize int) []byte {
	padding := blockSize - len(cipherText)%blockSize
	padText := bytes.Repeat([]byte{byte(padding)}, padding)
	return append(cipherText, padText...)
}

func pkcs5UnPadding(origData []byte) []byte {
	length := len(origData)
	unPadding := int(origData[length-1])
	return origData[:(length - unPadding)]
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

	iv = []byte(segment.Key.IV)
	return
}

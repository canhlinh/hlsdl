package hlsdl

import (
	"bytes"
	"crypto/aes"
	"crypto/cipher"
	"encoding/binary"
)

const (
	syncByte = uint8(71) //0x47
)

func decryptAES128(data, key, iv []byte) error {
	block, err := aes.NewCipher(key)
	if err != nil {
		return err
	}
	blockSize := block.BlockSize()
	blockMode := cipher.NewCBCDecrypter(block, iv[:blockSize])
	blockMode.CryptBlocks(data, data)
	c := pkcs5UnPadding(data)
	copy(data, c)
	return nil
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
func defaultIV(seqID uint64) []byte {
	buf := make([]byte, 16)
	binary.BigEndian.PutUint64(buf[8:], seqID)
	return buf
}

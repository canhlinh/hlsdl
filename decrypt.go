package hlsdl

import (
	"bytes"
	"crypto/aes"
	"crypto/cipher"
	"encoding/binary"
	"fmt"
	"io"
	"os"
)

const (
	syncByte = uint8(71) //0x47
)

func decryptAES128(crypted, key, iv []byte) ([]byte, error) {
	block, err := aes.NewCipher(key)
	if err != nil {
		return nil, err
	}
	blockSize := block.BlockSize()
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
func defaultIV(seqID uint64) []byte {
	buf := make([]byte, 16)
	binary.BigEndian.PutUint64(buf[8:], seqID)
	return buf
}

func decryptAES128Stream(filePath string, w io.Writer, key, iv []byte) error {
	file, err := os.Open(filePath)
	if err != nil {
		return err
	}
	defer file.Close()

	block, err := aes.NewCipher(key)
	if err != nil {
		return fmt.Errorf("failed to create AES cipher: %w", err)
	}

	blockSize := block.BlockSize()
	if len(iv) < blockSize {
		return fmt.Errorf("IV length must be at least %d bytes", blockSize)
	}

	// Create a CBC decrypter
	blockMode := cipher.NewCBCDecrypter(block, iv[:blockSize])

	// Create buffers for processing blocks
	buffer := make([]byte, blockSize)
	decrypted := make([]byte, blockSize)

	// Read and decrypt each block
	for {
		n, err := file.Read(buffer)
		if err != nil && err != io.EOF {
			return fmt.Errorf("failed to read encrypted data: %w", err)
		}
		if n == 0 {
			break
		}

		if n < blockSize {
			return fmt.Errorf("encrypted data size is not a multiple of block size")
		}

		// Decrypt the block
		blockMode.CryptBlocks(decrypted, buffer)

		// Write the decrypted data
		if _, err := w.Write(decrypted); err != nil {
			return fmt.Errorf("failed to write decrypted data: %w", err)
		}
	}

	return nil
}

func decodeStream(filePath string, w io.Writer) error {
	file, err := os.Open(filePath)
	if err != nil {
		return err
	}
	defer file.Close()

	if _, err := io.Copy(w, file); err != nil {
		return fmt.Errorf("failed to copy data: %w", err)
	}
	return nil
}

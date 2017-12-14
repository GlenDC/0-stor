package processing

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"errors"
	"io"
)

type EncryptionType uint8

const (
	EncryptionTypeAES EncryptionType = iota

	MaxStandardEncryptionType = EncryptionTypeAES
)

func NewAESEncrypterDecrypter(privateKey []byte) (*AESEncrypterDecrypter, error) {
	block, err := aes.NewCipher(privateKey)
	if err != nil {
		return nil, err
	}

	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return nil, err
	}

	nonceSize := gcm.NonceSize()
	return &AESEncrypterDecrypter{
		gcm:       gcm,
		nonce:     make([]byte, nonceSize),
		nonceSize: nonceSize,
	}, nil
}

type AESEncrypterDecrypter struct {
	gcm       cipher.AEAD
	nonce     []byte
	nonceSize int
}

func (ed *AESEncrypterDecrypter) WriteProcess(cipher []byte) (plain []byte, err error) {
	if len(cipher) <= ed.nonceSize {
		return nil, errors.New("malformed ciphertext")
	}
	return ed.gcm.Open(nil, cipher[:ed.nonceSize], cipher[ed.nonceSize:], nil)
}

func (ed *AESEncrypterDecrypter) ReadProcess(plain []byte) (cipher []byte, err error) {
	_, err = io.ReadFull(rand.Reader, ed.nonce)
	if err != nil {
		return nil, err
	}
	return ed.gcm.Seal(ed.nonce, ed.nonce, plain, nil), nil
}

var (
	_ Processor = (*AESEncrypterDecrypter)(nil)
)

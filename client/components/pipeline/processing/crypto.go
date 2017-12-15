package processing

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"errors"
	"fmt"
	"io"
)

// NewEncrypterDecrypter returns a new instance for the given encryption type.
// If the encryption type is not registered as a valid encryption type, an error is returned.
//
// The given privateKey is not optional and has to be defined.
func NewEncrypterDecrypter(et EncryptionType, privateKey []byte) (Processor, error) {
	edc, ok := _EncryptionTypeValueToConstructorMapping[et]
	if !ok {
		return nil, fmt.Errorf("'%s' is not a valid/registered EncryptionType value", et)
	}
	return edc(privateKey)
}

// NewAESEncrypterDecrypter creates a new encrypter-decrypter processor,
// using the Golang std AES implementation as its internal algorithm.
//
// See AESEncrypterDecrypter for more information.
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

// AESEncrypterDecrypter defines a processor, which encrypts and decrypts,
// using the Golang std AES implementation as its internal algorithm.
//
// It will encrypt plain text to cipher text while writing,
// and it will decrypt cipher text to plain text while reading.
//
// What specific AES algorithm is used, depends on the given key size:
//
//    * 16 bytes: AES_128
//    * 24 bytes: AES_192
//    * 32 bytes: AES_256
//
// When giving a key of a size other than these 3,
// NewAESEncrypterDecrypter will return an error.
type AESEncrypterDecrypter struct {
	gcm       cipher.AEAD
	nonce     []byte
	nonceSize int
}

// WriteProcess implements Processor.WriteProcess
func (ed *AESEncrypterDecrypter) WriteProcess(plain []byte) (cipher []byte, err error) {
	_, err = io.ReadFull(rand.Reader, ed.nonce)
	if err != nil {
		return nil, err
	}
	return ed.gcm.Seal(ed.nonce, ed.nonce, plain, nil), nil
}

// ReadProcess implements Processor.ReadProcess
func (ed *AESEncrypterDecrypter) ReadProcess(cipher []byte) (plain []byte, err error) {
	if len(cipher) <= ed.nonceSize {
		return nil, errors.New("malformed ciphertext")
	}
	return ed.gcm.Open(nil, cipher[:ed.nonceSize], cipher[ed.nonceSize:], nil)
}

// SharedWriteBuffer implements Processor.SharedWriteBuffer
func (ed *AESEncrypterDecrypter) SharedWriteBuffer() bool { return false }

// SharedReadBuffer implements Processor.SharedReadBuffer
func (ed *AESEncrypterDecrypter) SharedReadBuffer() bool { return false }

var (
	_ Processor = (*AESEncrypterDecrypter)(nil)
)

func init() {
	// register our standard encryption types, which is just one for now
	RegisterEncrypterDecrypter(EncryptionTypeAES, "aes",
		func(privateKey []byte) (Processor, error) {
			return NewAESEncrypterDecrypter(privateKey)
		})
}

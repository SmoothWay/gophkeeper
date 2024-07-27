package encrypt

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/sha256"
	"encoding/hex"
)

func EncodeMsg(msgBytes []byte, secret string) string {
	key := sha256.Sum256([]byte(secret))

	aesblock, err := aes.NewCipher(key[:])
	if err != nil {
		panic(err)
	}
	aesgcm, err := cipher.NewGCM(aesblock)
	if err != nil {
		panic(err)
	}

	nonce := key[len(key)-aesgcm.NonceSize():]

	dst := aesgcm.Seal(nil, nonce, msgBytes, nil) // зашифровываем
	return hex.EncodeToString(dst)
}

func DecodeMsg(msg string, secret string) string {
	key := sha256.Sum256([]byte(secret))

	aesblock, err := aes.NewCipher(key[:])
	if err != nil {
		panic(err)
	}
	aesgcm, err := cipher.NewGCM(aesblock)
	if err != nil {
		panic(err)
	}

	decodedMsg, err := hex.DecodeString(msg)
	if err != nil {
		panic(err)
	}

	nonce := key[len(key)-aesgcm.NonceSize():]
	decrypted, err := aesgcm.Open(nil, nonce, decodedMsg, nil)
	if err != nil {
		panic(err)
	}

	return string(decrypted)
}

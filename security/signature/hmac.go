package signature

import (
	"crypto/hmac"
	"encoding/hex"
	"encoding/json"
	"hash"
)

// CreateHMAC generates an HMAC hash of a message string using the given secret key and hash function.
// It returns the hex-encoded HMAC string.
func CreateHMAC(message, secret string, hashFunc func() hash.Hash) (string, error) {
	key := []byte(secret)
	h := hmac.New(hashFunc, key)
	_, err := h.Write([]byte(message))
	if err != nil {
		return "", err
	}

	return hex.EncodeToString(h.Sum(nil)), nil
}

// VerifyHMAC checks if the received HMAC is valid by re-generating the expected HMAC
// using the same message, secret key, and hash function.
// Returns true if the HMACs match, otherwise false.
func VerifyHMAC(message, receivedHMAC, secret string, hashFunc func() hash.Hash) (bool, error) {
	expectedHMAC, err := CreateHMAC(message, secret, hashFunc)
	if err != nil {
		return false, err
	}

	return hmac.Equal([]byte(expectedHMAC), []byte(receivedHMAC)), nil
}

// CreateGenericHMAC generates an HMAC hash for any type of message using the given secret key and hash function.
// The message is marshaled into bytes if it is not already a string or byte slice.
// Returns the hex-encoded HMAC string.
func CreateGenericHMAC(message any, secret string, hashFunc func() hash.Hash) (string, error) {
	var msgBytes []byte
	switch v := any(message).(type) {
	case []byte:
		msgBytes = v
	case string:
		msgBytes = []byte(v)
	default:
		var err error
		msgBytes, err = json.Marshal(v)
		if err != nil {
			return "", err
		}
	}

	key := []byte(secret)
	h := hmac.New(hashFunc, key)
	_, err := h.Write(msgBytes)
	if err != nil {
		return "", err
	}

	return hex.EncodeToString(h.Sum(nil)), nil
}

// VerifyGenericHMAC checks if the received HMAC is valid for any generic message type.
// It regenerates the expected HMAC using the same secret and hash function and compares it to the received one.
// Returns true if the HMACs match, otherwise false.
func VerifyGenericHMAC(message any, receivedHMAC, secret string, hashFunc func() hash.Hash) (bool, error) {
	expectedHMAC, err := CreateGenericHMAC(message, secret, hashFunc)
	if err != nil {
		return false, err
	}

	return hmac.Equal([]byte(expectedHMAC), []byte(receivedHMAC)), nil
}

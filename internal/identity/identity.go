package identity

import (
	"crypto/ed25519"
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"encoding/hex"
)

func GenerateKeypair() (pubKeyB64 string, privKeyB64 string, err error) {
	pub, priv, err := ed25519.GenerateKey(rand.Reader)
	if err != nil {
		return "", "", err
	}
	return base64.StdEncoding.EncodeToString(pub), base64.StdEncoding.EncodeToString(priv), nil
}

func DeviceIDFromPublicKey(pubKeyB64 string) (string, error) {
	pub, err := base64.StdEncoding.DecodeString(pubKeyB64)
	if err != nil {
		return "", err
	}
	h := sha256.Sum256(pub)
	return hex.EncodeToString(h[:16]), nil
}

func Sign(privKeyB64 string, message []byte) (string, error) {
	priv, err := base64.StdEncoding.DecodeString(privKeyB64)
	if err != nil {
		return "", err
	}
	sig := ed25519.Sign(ed25519.PrivateKey(priv), message)
	return base64.StdEncoding.EncodeToString(sig), nil
}

func Verify(pubKeyB64 string, message []byte, sigB64 string) (bool, error) {
	pub, err := base64.StdEncoding.DecodeString(pubKeyB64)
	if err != nil {
		return false, err
	}
	sig, err := base64.StdEncoding.DecodeString(sigB64)
	if err != nil {
		return false, err
	}
	return ed25519.Verify(ed25519.PublicKey(pub), message, sig), nil
}

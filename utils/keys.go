package utils

import (
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"encoding/pem"
)

const BIT_SIZE = 4096

func GenerateRSAKeyPairString() (string, string, error) {
	privKey, pubKey, err := generateRSAKeyPair()
	if err != nil {
		return "", "", err
	}
	pubKeyStr, err := publicKeyToString(pubKey)
	if err != nil {
		return "", "", err
	}
	privKeyStr, err := privateKeyToString(privKey)
	if err != nil {
		return "", "", err
	}

	return pubKeyStr, privKeyStr, nil

}

func generateRSAKeyPair() (*rsa.PrivateKey, *rsa.PublicKey, error) {
	privateKey, err := rsa.GenerateKey(rand.Reader, BIT_SIZE)
	if err != nil {
		return nil, nil, err
	}
	publicKey := &privateKey.PublicKey
	return privateKey, publicKey, nil
}

func privateKeyToString(privateKey *rsa.PrivateKey) (string, error) {
	privateKeyBytes := x509.MarshalPKCS1PrivateKey(privateKey)
	privateKeyPEM := &pem.Block{
		Type:  "RSA PRIVATE KEY",
		Bytes: privateKeyBytes,
	}
	privateKeyString := string(pem.EncodeToMemory(privateKeyPEM))
	return privateKeyString, nil
}

func publicKeyToString(publicKey *rsa.PublicKey) (string, error) {
	publicKeyBytes, err := x509.MarshalPKIXPublicKey(publicKey)
	if err != nil {
		return "", err
	}
	publicKeyPEM := &pem.Block{
		Type:  "PUBLIC KEY",
		Bytes: publicKeyBytes,
	}
	publicKeyString := string(pem.EncodeToMemory(publicKeyPEM))
	return publicKeyString, nil
}

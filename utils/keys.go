package utils

import (
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"encoding/pem"
	"golang.org/x/crypto/ssh"
	"strings"
)

const BIT_SIZE = 2048

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

func MakeSSHKeyPair() (string, string, error) {
    privateKey, err := rsa.GenerateKey(rand.Reader, 4096)
    if err != nil {
        return "", "", err
    }

    // generate and write private key as PEM
    var privKeyBuf strings.Builder

    privateKeyPEM := &pem.Block{Type: "RSA PRIVATE KEY", Bytes: x509.MarshalPKCS1PrivateKey(privateKey)}

    if err := pem.Encode(&privKeyBuf, privateKeyPEM); err != nil {
        return "", "", err
    }

    // generate and write public key
    pub, err := ssh.NewPublicKey(&privateKey.PublicKey)
    if err != nil {
        return "", "", err
    }

    var pubKeyBuf strings.Builder
    pubKeyBuf.Write(ssh.MarshalAuthorizedKey(pub))

    return pubKeyBuf.String(), privKeyBuf.String(), nil
}

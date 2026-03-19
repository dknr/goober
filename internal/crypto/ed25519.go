package crypto

import (
	"crypto/ed25519"
	"crypto/rand"
	"errors"
	"fmt"
	"os"
)

type KeyPair struct {
	PublicKey  *PublicKey
	PrivateKey *PrivateKey
}

type PublicKey struct {
	key []byte
}

type PrivateKey struct {
	key []byte
}

func (k *PublicKey) String() string {
	if k == nil || k.key == nil {
		return ""
	}
	return "ssh-ed25519 " + string(k.key)
}

func (k *PrivateKey) String() string {
	if k == nil || k.key == nil {
		return ""
	}
	return string(k.key)
}

func (k *PublicKey) WriteFile(path string) error {
	if k == nil || k.key == nil {
		return errors.New("cannot write nil public key")
	}
	return os.WriteFile(path, k.key, 0644)
}

func (k *PrivateKey) WriteFile(path string) error {
	if k == nil || k.key == nil {
		return errors.New("cannot write nil private key")
	}
	return os.WriteFile(path, k.key, 0600)
}

// GenerateKeyPair creates a new ed25519 key pair
func GenerateKeyPair() (*KeyPair, error) {
	publicKey, privateKey, err := ed25519.GenerateKey(rand.Reader)
	if err != nil {
		return nil, fmt.Errorf("failed to generate key pair: %w", err)
	}

	return &KeyPair{
		PublicKey:  &PublicKey{key: publicKey[:]},
		PrivateKey: &PrivateKey{key: privateKey[:]},
	}, nil
}

// ReadPrivateKey reads an ed25519 private key from a file
func ReadPrivateKey(path string) (*PrivateKey, error) {
	keyData, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("failed to read private key: %w", err)
	}

	return &PrivateKey{key: keyData}, nil
}

// ReadPublicKey reads an ed25519 public key from a file
func ReadPublicKey(path string) (*PublicKey, error) {
	keyData, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("failed to read public key: %w", err)
	}

	return &PublicKey{key: keyData}, nil
}

// Sign signs data with the private key
func (k *PrivateKey) Sign(data []byte) ([]byte, error) {
	if k == nil {
		return nil, errors.New("private key is nil")
	}
	return ed25519.Sign(k.key, data), nil
}

// Verify verifies a signature with the public key
func (k *PublicKey) Verify(data, signature []byte) bool {
	if k == nil {
		return false
	}
	return ed25519.Verify(k.key, data, signature)
}

// ToSSHFormat converts a public key to SSH format
func (k *PublicKey) ToSSHFormat() (string, error) {
	if k == nil || k.key == nil {
		return "", errors.New("public key is nil")
	}
	return "ssh-ed25519 " + string(k.key), nil
}
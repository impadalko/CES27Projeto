package sign

import (
    "crypto"

    // Cryptographically secure random number generator
    "crypto/rand"

    // Public/private key cryptography implementation
    "crypto/rsa"

    // Cryptographically secure hash implementation
    "crypto/sha256"
)

func Hash(data []byte) []byte {
    hash := sha256.Sum256(data)
    return hash[:]
}

func Sign(privKey *PrivateKey, hash []byte) ([]byte, error) {
    // func rsa.SignPKCS1v15(rand io.Reader, priv *rsa.PrivateKey, hash crypto.Hash, hashed []byte) ([]byte, error)
    return rsa.SignPKCS1v15(rand.Reader, (*rsa.PrivateKey)(privKey), crypto.SHA256, hash)
}

func Verify(pubKey *PublicKey, hash []byte, signature []byte) error {
    // func rsa.VerifyPKCS1v15(pub *rsa.PublicKey, hash crypto.Hash, hashed []byte, sig []byte) error
    return rsa.VerifyPKCS1v15((*rsa.PublicKey)(pubKey), crypto.SHA256, hash, signature)
}
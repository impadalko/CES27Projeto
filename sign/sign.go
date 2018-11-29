package sign

// Handles hashing, signing of hash with private key and verification of signature
// using public key

import (
    "crypto"

    // Cryptographically secure random number generator
    "crypto/rand"

    // Public/private key cryptography implementation
    "crypto/rsa"

    // Cryptographically secure hash implementation
    "crypto/sha256"
)

// Generates a 256 bit (32 bytes) checksum from @data
func Hash(data []byte) []byte {
    hash := sha256.Sum256(data)
    return hash[:]
}

// @hash is hashed using the crypto.SHA256 hash function. Returns the signed hash
// of message. Provides authenticity but not confidentiality.
func Sign(privKey *rsa.PrivateKey, hash []byte) ([]byte, error) {
    // func rsa.SignPKCS1v15(rand io.Reader, priv *rsa.PrivateKey, hash crypto.Hash, hashed []byte) ([]byte, error)
    return rsa.SignPKCS1v15(rand.Reader, privKey, crypto.SHA256, hash)
}

// Verifies if @signature is a valid signature of @hash using the correspondent
// private key of @pubKey
func Verify(pubKey *rsa.PublicKey, hash []byte, signature []byte) error {
    // func rsa.VerifyPKCS1v15(pub *rsa.PublicKey, hash crypto.Hash, hashed []byte, sig []byte) error
    return rsa.VerifyPKCS1v15(pubKey, crypto.SHA256, hash, signature)
}

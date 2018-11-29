package sign

// Functions for public and private RSA key generation, besides writing and reading
// of such keys from files in a encoded way

import (
    "os"
    "io/ioutil"
    "errors"

    // Cryptographically secure random number generator
    "crypto/rand"
    
    // Public/private key cryptography implementation
    "crypto/rsa"

    // Serialize public and private keys to text format
    "crypto/x509"

    // PEM = Privacy Enhanced Mail
    "encoding/pem"
)

const KeySizeBits = 2048

/*
type rsa.PrivateKey struct {
    rsa.PublicKey            // public part.
    D         *big.Int   // private exponent
    Primes    []*big.Int // prime factors of N, has >= 2 elements.
}

type rsa.PublicKey struct {
    N *big.Int // modulus
    E int      // public exponent
}

type pem.Block struct {
    Type    string            // The type, taken from the preamble (i.e. "RSA PRIVATE KEY").
    Headers map[string]string // Optional headers.
    Bytes   []byte            // The decoded bytes of the contents. Typically a DER encoded ASN.1 structure.
}
*/

type PrivateKey rsa.PrivateKey
type PublicKey  rsa.PublicKey

// Generates RSA private key with size @KeySizeBits
func GenerateKey() (*PrivateKey, error) {
    // func rsa.GenerateKey(random io.Reader, bits int) (*rsa.PrivateKey, error)
    rsaPrivKey, err := rsa.GenerateKey(rand.Reader, KeySizeBits)
    return (*PrivateKey)(rsaPrivKey), err
}

// Retrieves RSA public key from given @privKey
func (privKey *PrivateKey) GetPublicKey() *PublicKey {
    return (*PublicKey)(&privKey.PublicKey)
}

// Writes encoded @privKey into a file with name @filename
func (privKey *PrivateKey) WriteToPemFile(filename string) error {
    PrivKeyFile, err := os.Create(filename)
    if err != nil {
        return err
    }

    // func x509.MarshalPKCS1PrivateKey(key *rsa.PrivateKey) []byte
    bytes := x509.MarshalPKCS1PrivateKey((*rsa.PrivateKey)(privKey))

    // func pem.Encode(out io.Writer, b *pem.Block) error
    err = pem.Encode(PrivKeyFile,
        &pem.Block{
            Type: "RSA PRIVATE KEY",
            Bytes: bytes,
        },
    )
    if err != nil {
        return err
    }

    err = PrivKeyFile.Close()
    if err != nil {
        return err
    }

    return nil
}

// Writes encoded @pubKey into a file with name @filename
func (pubKey *PublicKey) WriteToPemFile(filename string) error {
    PubKeyFile, err := os.Create(filename)
    if err != nil {
        return err
    }

    // func x509.MarshalPKCS1PublicKey(key *rsa.PublicKey) []byte
    bytes := x509.MarshalPKCS1PublicKey((*rsa.PublicKey)(pubKey))

    // func pem.Encode(out io.Writer, b *pem.Block) error
    err = pem.Encode(PubKeyFile,
        &pem.Block{
            Type: "RSA PUBLIC KEY",
            Bytes: bytes,
        },
    )
    if err != nil {
        return err
    }

    err = PubKeyFile.Close()
    if err != nil {
        return err
    }

    return nil
}

// Retrieves encoded @privKey from file with name @filename
func PrivateKeyFromPemFile(filename string) (*PrivateKey, error) {
    bytes, err := ioutil.ReadFile(filename)
    if err != nil {
        return nil, err
    }

    // func pem.Decode(data []byte) (p *pem.Block, rest []byte)
    pem, rest := pem.Decode(bytes)
    if pem == nil || pem.Type != "RSA PRIVATE KEY" || len(rest) != 0 {
        return nil, errors.New("Error decoding PEM key")
    }

    // func x509.ParsePKCS1PrivateKey(der []byte) (*rsa.PrivateKey, error)
    rsaPrivKey, err := x509.ParsePKCS1PrivateKey(pem.Bytes)
    if err != nil {
        return nil, err
    }

    return (*PrivateKey)(rsaPrivKey), nil
}

// Retrieves encoded @pubKey from file with name @filename
func PublicKeyFromPemFile(filename string) (*PublicKey, error) {
    bytes, err := ioutil.ReadFile(filename)
    if err != nil {
        return nil, err
    }

    // func pem.Decode(data []byte) (p *pem.Block, rest []byte)
    pem, rest := pem.Decode(bytes)
    if pem == nil || pem.Type != "RSA PUBLIC KEY" || len(rest) != 0 {
        return nil, errors.New("Error decoding PEM key")
    }

    // func x509.ParsePKCS1PublicKey(der []byte) (*rsa.PublicKey, error)
    rsaPubKey, err := x509.ParsePKCS1PublicKey(pem.Bytes)
    if err != nil {
        return nil, err
    }

    return (*PublicKey)(rsaPubKey), nil
}

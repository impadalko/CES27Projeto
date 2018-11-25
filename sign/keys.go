package sign

import (
    "os"
    "io/ioutil"

    // Cryptographically secure random number generator
    "crypto/rand"
    
    // Public/private key cryptography implementation
    "crypto/rsa"

    // Serialize public and private keys to text format
    "crypto/x509"
    "encoding/pem"
)

const KeySizeBits = 2048

type PrivateKey rsa.PrivateKey
type PublicKey  rsa.PublicKey

func GenerateKey() (*PrivateKey, error) {
    rsaPrivKey, err := rsa.GenerateKey(rand.Reader, KeySizeBits)
    privKey := (*PrivateKey)(rsaPrivKey) // solve type-checker complaint
    return privKey, err
}

func (privKey *PrivateKey) GetPublicKey() *PublicKey {
    return (*PublicKey)(&privKey.PublicKey)
}

func (privKey *PrivateKey) WriteToPemFile(filename string) error {
    PrivKeyFile, err := os.Create(filename)
    if err != nil {
        return err
    }
    rsaPrivKey := (*rsa.PrivateKey)(privKey) // solve type-checker complaint
    err = pem.Encode(PrivKeyFile,
        &pem.Block{
            Type: "RSA PRIVATE KEY",
            Bytes: x509.MarshalPKCS1PrivateKey(rsaPrivKey),
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

func (pubKey *PublicKey) WriteToPemFile(filename string) error {
    PubKeyFile, err := os.Create(filename)
    if err != nil {
        return err
    }
    rsaPubKey := (*rsa.PublicKey)(pubKey) // solve type-checker complaint
    err = pem.Encode(PubKeyFile,
        &pem.Block{
            Type: "RSA PUBLIC KEY",
            Bytes: x509.MarshalPKCS1PublicKey(rsaPubKey),
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

type DecodeKeyError struct {
}

func (err *DecodeKeyError) Error() string {
    return "Error decoding PEM key"
}

func PrivateKeyFromPemFile(filename string) (*PrivateKey, error) {
    bytes, err := ioutil.ReadFile(filename)
    if err != nil {
        return nil, err
    }

    pem, remaining := pem.Decode(bytes)
    if pem == nil || pem.Type != "RSA PRIVATE KEY" || len(remaining) > 0 {
        return nil, &DecodeKeyError{}
    }

    rsaPrivKey, err := x509.ParsePKCS1PrivateKey(pem.Bytes)
    if err != nil {
        return nil, err
    }

    return (*PrivateKey)(rsaPrivKey), nil
}

func PublicKeyFromPemFile(filename string) (*PublicKey, error) {
    bytes, err := ioutil.ReadFile(filename)
    if err != nil {
        return nil, err
    }

    pem, remaining := pem.Decode(bytes)
    if pem == nil || pem.Type != "RSA PUBLIC KEY" || len(remaining) > 0 {
        return nil, &DecodeKeyError{}
    }

    rsaPubKey, err := x509.ParsePKCS1PublicKey(pem.Bytes)
    if err != nil {
        return nil, err
    }

    return (*PublicKey)(rsaPubKey), nil
}

type TestError struct {
    msg string
}

func (err *TestError) Error() string {
    return err.msg
}

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
*/

func TestWriteKeysToPemFile() error {
	privKey, err := GenerateKey()
	if err != nil {
		return err
	}
	pubKey := privKey.GetPublicKey()
	
	err = privKey.WriteToPemFile("priv_key.pem")
	if err != nil {
		return err
	}

	err = pubKey.WriteToPemFile("pub_key.pem")
	if err != nil {
		return err
    }

    readPrivKey, err := PrivateKeyFromPemFile("priv_key.pem")
    if err != nil {
		return err
    }

    if readPrivKey.D.Cmp(privKey.D) != 0 ||
        readPrivKey.PublicKey.N.Cmp(privKey.PublicKey.N) != 0 ||
        readPrivKey.PublicKey.E != privKey.PublicKey.E {
        return &TestError{"Error reading private key from file"}
    }

    readPubKey, err := PublicKeyFromPemFile("pub_key.pem")
    if err != nil {
		return err
    }

    if readPubKey.N.Cmp(pubKey.N) != 0 || readPubKey.E != pubKey.E {
        return &TestError{"Error reading public key from file"}
    }

    return nil
}
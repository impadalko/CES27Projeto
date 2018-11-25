package sign

import (
    "fmt"
    "os"

    // Cryptographically secure random number generator
    "crypto/rand"
    
    // Public/private key cryptography implementation
    "crypto/rsa"

    // Serialize public and private keys to text format
    "crypto/x509"
    "encoding/pem"
)

func checkError(err error) {
    if err != nil {
        fmt.Println("An error occurred! Please try again!")
        os.Exit(1)
    }
}

/*
type PrivateKey struct {
    PublicKey            // public part.
    D         *big.Int   // private exponent
    Primes    []*big.Int // prime factors of N, has >= 2 elements.
}

type PublicKey struct {
    N *big.Int // modulus
    E int      // public exponent
}
*/

func genTest() {
    // Generate public and private key pair
    // func GenerateKey(random io.Reader, bits int) (*PrivateKey, error)
    PrivKey, err := rsa.GenerateKey(rand.Reader, 1024)
    checkError(err)
    PubKey := PrivKey.PublicKey
    fmt.Println("Public and private key pair generated succesfully\n")

    // Write private key to text file in pem format
    PrivKeyFile, err := os.Create("rsa_priv.pem")
    checkError(err)
    err = pem.Encode(PrivKeyFile,
        &pem.Block{
            Type: "RSA PRIVATE KEY",
            Bytes: x509.MarshalPKCS1PrivateKey(PrivKey),
        },
    )
    err = PrivKeyFile.Close()
    checkError(err)
    fmt.Println("Private key written to file succesfully\n")

    // Write public key to text file in pem format
    PubKeyFile, err := os.Create("rsa_pub.pem")
    checkError(err)
    err = pem.Encode(PubKeyFile,
        &pem.Block{
            Type: "RSA PUBLIC KEY",
            Bytes: x509.MarshalPKCS1PublicKey(&PubKey),
        },
    )
    err = PubKeyFile.Close()
    checkError(err)
    fmt.Println("Public key written to file succesfully\n")
}

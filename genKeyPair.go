package main

import (
    "fmt"
    "crypto/rand"
    "crypto/rsa"
    "crypto/x509"
    "encoding/pem"
    "os"
)

func checkError(err error) {
    if err != nil {
        fmt.Println("An error occurred! Please try again!")
        os.Exit(1)
    }
}

func main() {
    // Generate Key-Pair
    PrivKey, err := rsa.GenerateKey(rand.Reader, 1024)
    checkError(err)
    PubKey := PrivKey.PublicKey

    // Encode Private Key to PEM format
    PrivKeyPem := pem.EncodeToMemory(
        &pem.Block{
            Type: "RSA PRIVATE KEY",
            Bytes: x509.MarshalPKCS1PrivateKey(PrivKey),
        },
    )

    // Encode Public Key to PEM format
    PubKeyPem := pem.EncodeToMemory(
        &pem.Block{
            Type: "RSA PUBLIC KEY",
            Bytes: x509.MarshalPKCS1PublicKey(&PubKey),
        },
    )

    // Write Private Key to file
    PrivKeyFile, err := os.Create("rsa.pem")
    checkError(err)
    fmt.Fprintf(PrivKeyFile, string(PrivKeyPem[:]))
    PrivKeyFile.Close()

    // Write Public Key to file
    PubKeyFile, err := os.Create("rsa_pub.pem")
    checkError(err)
    fmt.Fprintf(PubKeyFile, string(PubKeyPem[:]))
    PubKeyFile.Close()
}

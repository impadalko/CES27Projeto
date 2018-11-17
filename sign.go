package main

import (
    "fmt"
    "io/ioutil"
    "os"

    "crypto"

    // Cryptographically secure random number generator
    "crypto/rand"

    // Public/private key cryptography implementation
    "crypto/rsa"

    // Cryptographically secure hash implementation
    "crypto/sha256"
    
    // Serialize structs to text format
    "encoding/json"

    // Serialize public and private keys to text format
    "crypto/x509"
    "encoding/pem"
)

// TODO: Complete this struct
type Transaction struct {
    Data string
}

func checkError(err error) {
    if err != nil {
        fmt.Println("An error occurred! Please try again!")
        os.Exit(1)
    }
}

/*
type Block struct {
    Type    string            // The type, taken from the preamble (i.e. "RSA PRIVATE KEY").
    Headers map[string]string // Optional headers.
    Bytes   []byte            // The decoded bytes of the contents. Typically a DER encoded ASN.1 structure.
}
*/

func main(){
    // Read private key file and parse it
    // TODO: Pass private key file as argument
    privKeyBytes, err := ioutil.ReadFile("rsa_priv.pem")
    checkError(err)
    privKeyPem, _ := pem.Decode(privKeyBytes)
    if privKeyPem == nil || privKeyPem.Type != "RSA PRIVATE KEY" {
        fmt.Println("Unable to find the private key");
        os.Exit(1);
    }
    privKey, err := x509.ParsePKCS1PrivateKey(privKeyPem.Bytes)
    checkError(err)

    // Sign the struct
    // TODO: Think on how get the transaction (query the miner?)
    payload, err := json.Marshal(Transaction{ Data: "test" })
    checkError(err)
    hashed := sha256.Sum256(payload)

    // func SignPKCS1v15(rand io.Reader, priv *PrivateKey, hash crypto.Hash, hashed []byte) ([]byte, error)
    signature, err := rsa.SignPKCS1v15(rand.Reader, privKey, crypto.SHA256, hashed[:])
    checkError(err)

    fmt.Println("Data signed sucessfully\n")
    fmt.Printf("Signature: %x\n\n", signature)

    pubKeyData, err := ioutil.ReadFile("rsa_pub.pem")
    checkError(err)
    
    pubKeyPem, _ := pem.Decode(pubKeyData)
    if pubKeyPem == nil {
        fmt.Println("Unable to find PEM formatted block");
        os.Exit(1);
    }

    pubKey, err := x509.ParsePKCS1PublicKey(pubKeyPem.Bytes)
    checkError(err)

    // func VerifyPKCS1v15(pub *PublicKey, hash crypto.Hash, hashed []byte, sig []byte) error
    err = rsa.VerifyPKCS1v15(pubKey, crypto.SHA256, hashed[:], signature)
    if err == nil {
        fmt.Println("Signature verified successfully")
    } else {
        fmt.Println("Failed to verify signature")
    }

    // TODO: append signature to the struct (different struct)
    // TODO: send signed struct to miner
}

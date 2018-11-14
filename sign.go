package main

import(
    "crypto"
    "crypto/rand"
    "crypto/rsa"
    "crypto/sha256"
    "crypto/x509"
    "encoding/json"
    "encoding/pem"
    "fmt"
    "io/ioutil"
    "os"
)

// TODO: Complete this struct
type tx struct {
    Data string
}

func checkError(err error) {
    if err != nil {
        fmt.Println("An error occurred! Please try again!")
        os.Exit(1)
    }
}

func main(){
    // Read private key file and parse it
    // TODO: Pass private key file as argument
    data, err := ioutil.ReadFile("rsa.pem")
    checkError(err)
    privPem, _ := pem.Decode(data)
    privKey, err := x509.ParsePKCS1PrivateKey(privPem.Bytes)
    checkError(err)

    // Sign the struct
    // TODO: Think on how get the transaction (query the miner?)
    payload, err := json.Marshal(tx{Data: "test"})
    checkError(err)
    hashed := sha256.Sum256(payload)
    signature, err := rsa.SignPKCS1v15(rand.Reader, privKey, crypto.SHA256, hashed[:])
    checkError(err)
    fmt.Printf("%x\n", signature)

    // TODO: append signature to the struct (different struct)
    // TODO: send signed struct to miner
}

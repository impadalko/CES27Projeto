package main

import(
    "crypto/x509"
    "encoding/pem"
    "fmt"
    "io/ioutil"
    "os"
)

func checkError(err error) {
    if err != nil {
        fmt.Println("An error occurred! Please try again!")
        os.Exit(1)
    }
}

func main(){
    data, err := ioutil.ReadFile("rsa.pem")
    checkError(err)
    privPem, _ := pem.Decode(data)
    privKey, err := x509.ParsePKCS1PrivateKey(privPem.Bytes)
    checkError(err)

    fmt.Println(privKey)
}

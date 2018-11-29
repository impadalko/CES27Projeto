package main

import (
	"fmt"
	"os"
	"bufio"
	"strings"
	"strconv"
	"encoding/hex"

	"github.com/impadalko/CES27Projeto/blockchain"
	"github.com/impadalko/CES27Projeto/network"
	"github.com/impadalko/CES27Projeto/sign"
	"github.com/impadalko/CES27Projeto/util"
)

func main() {
	node := NewNode(util.RandomString(8))
	err := node.Listen()
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
	node.PrintInfo()
	
	if len(os.Args) == 2 {
		// connect to another peer and join its network
		peerAddr := os.Args[1]
		conn, err := node.JoinNetwork(peerAddr)
		if err != nil {
			fmt.Println(err)
			os.Exit(1)
		}
		// request copy of the blockchain of the peer
		fmt.Fprintf(conn, "REQUEST-BLOCKCHAIN\n")
		go node.StartHandleConnection(conn)
	} else {
		// start own blockchain and network
		node.BlockChain = blockchain.New(util.Now(), []byte{})
		node.PrintBlocks()
	}

	go node.Start()

	reader := bufio.NewReader(os.Stdin)
	for {
		text, err := reader.ReadString('\n')
		if err != nil {
			break
		}
		fmt.Println()
		text = strings.TrimSpace(text)
		split := strings.Split(text, " ")

		if len(split) == 0 {
			fmt.Println("Invalid command")
			fmt.Println()
			continue
		}

		command := split[0]

		if command == "info" {

			node.PrintInfo()

		} else if command == "peers" {

			node.PrintPeers()

		} else if command == "conns" {

			node.PrintConns()

		} else if command == "blocks" {

			node.PrintBlocks()

		} else if len(split) >= 2 && command == "add" {

			message := strings.Join(split[1:], " ")
			node.AddBlockFromData(util.Now(), []byte(message))
			node.PrintBlocks()

		} else if len(split) == 2 && command == "cast" {

			blockIndex, err := strconv.Atoi(split[1])
			if err == nil {
				block := node.BlockChain.GetBlock(blockIndex)
				message := fmt.Sprintf("BLOCK-ADD %s\n", block.String())
				node.Broadcast(message)
			} else {
				fmt.Println("Invalid command")
				fmt.Println()
			}

		} else if len(split) == 1 && command == "verify" {

			if node.VerifyConsistency() {
				fmt.Println("The blockchain is consistent")
				fmt.Println()
			} else {
				fmt.Println("The blockchain is NOT consistent")
				fmt.Println()
			}

		} else if len(split) == 2 && command == "genkey" {

			privKey, err := sign.GenerateKey()
			if err != nil {
				fmt.Println(err)
			} else {
				pubKey := &privKey.PublicKey
				keyName := split[1]
				privateFilename := fmt.Sprintf("%s_priv.pem", keyName)
				publicFilename := fmt.Sprintf("%s_pub.pem", keyName)
				sign.WritePrivateKeyToPemFile(privKey, privateFilename)
				sign.WritePublicKeyToPemFile(pubKey, publicFilename)
				fmt.Printf("Generated private key %s saved to %s\n", keyName, privateFilename)
				fmt.Printf("Generated public key %s saved to %s\n", keyName, publicFilename)
				fmt.Println()
			}
		
		} else if len(split) == 2 && command == "privkey" {

			keyName := split[1]
			privateFilename := fmt.Sprintf("%s_priv.pem", keyName)
			privKey, err := sign.PrivateKeyFromPemFile(privateFilename)
			if err != nil {
				fmt.Println(err)
			} else {
				node.UsePrivateKey(privKey)
				fmt.Println("Using private key:", keyName)
				fmt.Println()
			}

		} else if len(split) == 2 && command == "pubkey" {

			keyName := split[1]
			publicFilename := fmt.Sprintf("%s_pub.pem", keyName)
			pubKey, err := sign.PublicKeyFromPemFile(publicFilename)
			if err != nil {
				fmt.Println(err)
			} else {
				node.UsePublicKey(pubKey)
				fmt.Println("Using public key:", keyName)
				fmt.Println()
			}

		} else if len(split) == 2 && command == "sign" {

			hash, err := hex.DecodeString(split[1])
			if err != nil {
				fmt.Println("Invalid hash")
				fmt.Println()
			} else {
				if node.PrivateKey == nil {
					fmt.Println("Please use a private key with privkey command")
					fmt.Println()
				} else {
					signature, err := sign.Sign(node.PrivateKey, hash)
					if err != nil {
						fmt.Println(err)
						fmt.Println()
					} else {
						fmt.Println("Signature added to the blockchain")
						fmt.Println()
						node.AddBlockFromData(util.Now(), signature)
					}
				}
			}

		} else {
			fmt.Println("Invalid command")
			fmt.Println()
		}
	}
}

func Tests() {
	var err error

	err = blockchain.TestBlockToStringAndFromString()
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}

	err = sign.TestWriteAndReadPemFile()
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}

	err = sign.TestSignAndVerify()
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
	
	err = network.TestNodeJoinNetwork()
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}

	fmt.Println("ALL TESTS PASSED")
}
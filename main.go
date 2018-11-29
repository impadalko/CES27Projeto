package main

import (
	"bufio"
	"encoding/hex"
	"errors"
	"fmt"
	"os"
	"strconv"
	"strings"
	"io/ioutil"

	"github.com/impadalko/CES27Projeto/blockchain"
	"github.com/impadalko/CES27Projeto/network"
	"github.com/impadalko/CES27Projeto/sign"
	"github.com/impadalko/CES27Projeto/util"
)

func main() {
	//Tests()

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

		err = HandleCommand(command, split)
		if err != nil {
			fmt.Println(err)
			fmt.Println()
		}
	}
}

func HandleCommand(command string, split []string) error {
	
	if command == "info" {
		// Display node id and node address

		node.PrintInfo()

	} else if command == "peers" {
		// Display the list of peers of the node

		node.PrintPeers()

	} else if command == "conns" {
		// Display the list of connections of the node

		node.PrintConns()

	} else if command == "blocks" {
		// Display the list of blocks of the blockchain

		node.PrintBlocks()

	} else if len(split) >= 2 && command == "add" {
		// Add a new block to the blockchain with a supplied hex string

		data, err := hex.DecodeString(split[1])
		if err != nil {
			return err
		}
		node.AddBlockFromData(util.Now(), data)
		node.PrintBlocks()

	} else if len(split) == 2 && command == "cast" {
		// Broadcast a block to all the peers

		blockIndex, err := strconv.ParseInt(split[1], 10, 64)
		if err != nil {
			return err
		}
		block, err := node.BlockChain.GetBlock(blockIndex)
		if err != nil {
			return err
		}
		message := fmt.Sprintf("BLOCK-ADD %s\n", block.String())
		node.Broadcast(message)

	} else if len(split) == 2 && command == "genkey" {
		// generate a private/public key pair

		privKey, err := sign.GenerateKey()
		if err != nil {
			return err
		}
		pubKey := &privKey.PublicKey
		keyName := split[1]
		
		privateFilename := fmt.Sprintf("%s_priv.pem", keyName)
		err = sign.WritePrivateKeyToPemFile(privKey, privateFilename)
		if err != nil {
			return err
		}
		fmt.Printf("Generated private key %s written to %s\n", keyName, privateFilename)

		publicFilename := fmt.Sprintf("%s_pub.pem", keyName)
		err = sign.WritePublicKeyToPemFile(pubKey, publicFilename)
		if err != nil {
			return err
		}
		fmt.Printf("Generated public key %s written to %s\n", keyName, publicFilename)
		fmt.Println()
	
	} else if len(split) == 2 && command == "privkey" {
		// use a supplied private key for signing and verification

		keyName := split[1]
		privateFilename := fmt.Sprintf("%s_priv.pem", keyName)
		privKey, err := sign.PrivateKeyFromPemFile(privateFilename)
		if err != nil {
			return err
		}
		node.UsePrivateKey(keyName, privKey)
		fmt.Println("Using private key:", keyName)
		fmt.Println()

	} else if len(split) == 2 && command == "pubkey" {
		// use a supplied public key for signing and verification

		keyName := split[1]
		publicFilename := fmt.Sprintf("%s_pub.pem", keyName)
		pubKey, err := sign.PublicKeyFromPemFile(publicFilename)
		if err != nil {
			return err
		}
		node.UsePublicKey(keyName, pubKey)
		fmt.Println("Using public key:", keyName)
		fmt.Println()

	} else if len(split) == 2 && command == "sign" {
		// sign a hash using the current private key and add the signature to the blockchain

		if node.PrivateKey == nil {
			return errors.New("Please use a private key with privkey command")
		}
		hash, err := hex.DecodeString(split[1])
		if err != nil {
			return err
		}
		signature, err := sign.Sign(node.PrivateKey, hash)
		if err != nil {
			return err
		}
		blockIndex, err := node.AddBlockFromData(util.Now(), signature)
		if err != nil {
			return err
		}
		fmt.Printf("The document with hash %s was signed with key %s and added to the blockchain in block %d\n\n", 
			util.HexString(split[1]), node.KeyName, blockIndex)

	} else if len(split) == 3 && command == "verify" {
		// verify a signature present in the blockchain using a supplied hash
		// and the current public key
		
		if node.PublicKey == nil {
			return errors.New("Please use a public key with pubkey command")
		}
		blockIndex, err := strconv.ParseInt(split[1], 10, 64)
		if err != nil {
			return err
		}
		hash, err := hex.DecodeString(split[2])
		if err != nil {
			return err
		}
		block, err := node.GetBlock(blockIndex)
		if err != nil {
			return err
		}
		signature := block.Data
		err = sign.Verify(node.PublicKey, hash, signature)
		if err != nil {
			fmt.Println("The signature is INVALID")
			fmt.Println()
		} else {
			fmt.Println("The signature is VALID")
			fmt.Printf("The document with hash %s was signed by %s in the block %d\n\n",
				util.HexString(split[2]), node.KeyName, blockIndex)
		}

	} else if len(split) == 2 && command == "hash" {
		// return the SHA256 hash of a file, given its filename
		
		filename := split[1]
		data, err := ioutil.ReadFile(filename)
		if err != nil {
			return err
		}
		hash := sign.Hash(data)
		fmt.Println("The SHA256 hash of the file given is:")
		fmt.Println(hex.EncodeToString(hash))
		fmt.Println()

	} else {

		return errors.New("Invalid Command")

	}
	return nil
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
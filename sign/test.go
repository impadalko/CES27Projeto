package sign

type TestError struct {
    msg string
}

func (err *TestError) Error() string {
    return err.msg
}

func TestWriteAndReadPemFile() error {
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

func TestSignAndVerify() error {
	privKey, err := GenerateKey()
	if err != nil {
		return err
	}
	pubKey := privKey.GetPublicKey()

	data := []byte("Hello World!")
	hash := Hash(data)

	signature, err := Sign(privKey, hash)
	if err != nil {
		return err
	}

	err = Verify(pubKey, hash, signature)
	if err != nil {
		return err
	}

	return nil
}
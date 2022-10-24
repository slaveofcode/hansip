package age_encryption

import (
	"io"
	"log"
	"os"
	"path/filepath"
	"strings"

	"filippo.io/age"
)

func EncryptFile(filePath, locationPath string, publicKeys []string) (string, error) {
	f, err := os.Open(filePath)
	if err != nil {
		return "", err
	}

	fName := filepath.Base(filePath) + ".age"
	destEncFile := filepath.Join(locationPath, fName)
	fe, err := os.Create(destEncFile)
	if err != nil {
		return "", err
	}

	var recipients []age.Recipient
	for _, publicKey := range publicKeys {
		rec, err := age.ParseX25519Recipient(publicKey)
		if err != nil {
			return "", err
		}
		recipients = append(recipients, rec)
	}

	wc, err := age.Encrypt(fe, recipients...)
	if err != nil {
		return "", err
	}

	_, err = io.Copy(wc, f)
	if err != nil {
		return "", err
	}

	if err := wc.Close(); err != nil {
		log.Println("Unable Close Writer File")
		return "", err
	}

	if err := fe.Close(); err != nil {
		log.Println("Unable Close Encrypted File")
		return "", err
	}

	if err := f.Close(); err != nil {
		log.Println("Unable Close Source File")
		return "", err
	}

	return destEncFile, nil
}

func DecryptFile(filePath, locationPath, secretKey string) (string, error) {
	f, err := os.Open(filePath)
	if err != nil {
		return "", err
	}

	defer f.Close()

	fName := filepath.Base(filePath)
	// realFName := fName[:len(fName)-len(filepath.Ext(fName))]
	realFName := strings.Trim(fName, ".age")

	destDecFile := filepath.Join(locationPath, realFName)
	fOut, err := os.Create(destDecFile)
	if err != nil {
		return "", err
	}

	defer fOut.Close()

	// _, sec := parseIdentity()
	identity, err := age.ParseX25519Identity(secretKey)
	// identities, err := age.ParseIdentities(strings.NewReader(sec))
	if err != nil {
		return "", err
	}

	r, err := age.Decrypt(f, identity)
	if err != nil {
		return "", err
	}

	if _, err = io.Copy(fOut, r); err != nil {
		return "", err
	}

	return destDecFile, nil
}

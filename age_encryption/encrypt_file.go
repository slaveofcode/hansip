package age_encryption

import (
	"io"
	"log"
	"os"
	"path/filepath"
	"strings"

	"filippo.io/age"
)

func parseIdentity() (pub string, sec string) {
	pub = "age1fcnvt02ahjprhxye8yqhdcq8g3lfyj7a3mgqvwv29468asnya3cqm4d3zv"
	sec = "AGE-SECRET-KEY-1VW8ZNL7SEDWTU4X8HNMQKZE5EEM2EN7KZHMQ73NTZ8HYCCSJKJEQZG9WLP"
	return pub, sec
	// keyFile, err := os.Open("./keys/key-pass.age")
	// if err != nil {
	// 	log.Fatalf("Failed to open private keys file: %v", err)
	// }

	// pass := "2d1ty2vigen" // TODO: Test only
	// identities := []age.Identity{
	// 	&LazyScryptIdentity{
	// 		Passphrase: func() (string, error) {
	// 			return pass, nil
	// 		},
	// 	},
	// }

	// rr := bufio.NewReader(keyFile)
	// var in io.Reader
	// if start, _ := rr.Peek(len(armor.Header)); string(start) == armor.Header {
	// 	in = armor.NewReader(rr)
	// } else {
	// 	in = rr
	// }

	// reader, err := age.Decrypt(in, identities...)

	// if err != nil {
	// 	errorf("%v", err)
	// }

	// buf := &bytes.Buffer{}

	// if _, err := io.Copy(buf, reader); err != nil {
	// 	errorf("%v", err)
	// }
}

func errorf(format string, v ...interface{}) {
	log.Printf("age: error: "+format, v...)
}

func EncryptFile(filePath, locationPath string) (string, error) {
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

	pub, _ := parseIdentity()
	rec, err := age.ParseX25519Recipient(pub)
	if err != nil {
		return "", err
	}

	wc, err := age.Encrypt(fe, rec)
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

func DecryptFile(filePath, locationPath string) (string, error) {
	f, err := os.Open(filePath)
	if err != nil {
		return "", err
	}

	defer f.Close()

	fName := filepath.Base(filePath)
	// realFName := fName[:len(fName)-len(filepath.Ext(fName))]
	realFName := strings.Trim(fName, ".age") + ".md"

	destDecFile := filepath.Join(locationPath, realFName)
	fOut, err := os.Create(destDecFile)
	if err != nil {
		return "", err
	}

	defer fOut.Close()

	_, sec := parseIdentity()
	identity, err := age.ParseX25519Identity(sec)
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

package main

import (
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"crypto/x509/pkix"
	"encoding/pem"
	"flag"
	"fmt"
	"math/big"
	"net"
	"os"
	"time"

	"github.com/blackchip-org/chatty/internal/passwd"
	"github.com/blackchip-org/chatty/irc"
	"github.com/boltdb/bolt"
)

var (
	dataFile string
)

func init() {
	flag.StringVar(&dataFile, "data", "chatty.data", "create this file as the data store")
}

func main() {
	err := run()
	if err != nil {
		fmt.Printf("error: %v\n", err)
		os.Exit(1)
	}
}

func run() error {
	boltOpts := bolt.Options{Timeout: 5 * time.Second}
	db, err := bolt.Open(dataFile, 0600, &boltOpts)
	if err != nil {
		return fmt.Errorf("unable to open database %v: %v", dataFile, err)
	}
	defer db.Close()

	err = db.Update(func(tx *bolt.Tx) error {
		config := tx.Bucket([]byte("config"))
		if config != nil {
			return fmt.Errorf("file already exists: %v", dataFile)
		}
		config, err := tx.CreateBucket([]byte("config"))
		if err != nil {
			return err
		}

		if err := selfSign(config); err != nil {
			return err
		}
		if err := serverPass(config); err != nil {
			return err
		}
		return nil
	})
	return err
}

// https://golang.org/src/crypto/tls/generate_cert.go
func selfSign(config *bolt.Bucket) error {
	const rsaBits = 2048
	priv, err := rsa.GenerateKey(rand.Reader, rsaBits)

	serialNumberLimit := new(big.Int).Lsh(big.NewInt(1), 128)
	serialNumber, err := rand.Int(rand.Reader, serialNumberLimit)
	notBefore := time.Now()
	notAfter := time.Now().Add(100 * 365 * 24 * time.Hour)

	template := x509.Certificate{
		SerialNumber: serialNumber,
		Subject: pkix.Name{
			Organization: []string{"chatty"},
		},
		NotBefore: notBefore,
		NotAfter:  notAfter,

		KeyUsage:              x509.KeyUsageKeyEncipherment | x509.KeyUsageDigitalSignature,
		ExtKeyUsage:           []x509.ExtKeyUsage{x509.ExtKeyUsageServerAuth},
		BasicConstraintsValid: true,

		IPAddresses: []net.IP{net.ParseIP("127.0.0.1")},
	}

	derBytes, err := x509.CreateCertificate(rand.Reader, &template, &template, &priv.PublicKey, priv)
	if err != nil {
		return fmt.Errorf("failed to create certificate: %s", err)
	}
	cert := pem.EncodeToMemory(&pem.Block{
		Type:  "CERTIFICATE",
		Bytes: derBytes,
	})
	key := pem.EncodeToMemory(&pem.Block{
		Type:  "RSA PRIVATE KEY",
		Bytes: x509.MarshalPKCS1PrivateKey(priv),
	})

	config.Put(irc.ConfigCert, cert)
	config.Put(irc.ConfigKey, key)
	return nil
}

func serverPass(config *bolt.Bucket) error {
	plaintext, err := passwd.Random()
	if err != nil {
		return err
	}
	salt, err := passwd.Salt()
	if err != nil {
		return err
	}
	pass := passwd.Encode([]byte(plaintext), salt)
	config.Put(irc.ConfigPass, pass)
	config.Put(irc.ConfigSalt, salt)
	fmt.Printf("connection password is: %v\n", plaintext)
	return nil
}

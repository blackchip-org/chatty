package main

import (
	"flag"
	"fmt"
	"os"
	"time"

	"github.com/blackchip-org/chatty/internal/security"
	"github.com/blackchip-org/chatty/irc"
	"github.com/boltdb/bolt"
)

var (
	dataFile   string
	noPassword bool
)

func init() {
	flag.StringVar(&dataFile, "data", "chatty.data", "create this file as the data store")
	flag.BoolVar(&noPassword, "no-password", false, "do not set a connection password")
}

func main() {
	flag.Parse()
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
		config := tx.Bucket(irc.BucketConfig)
		if config != nil {
			return fmt.Errorf("file already exists: %v", dataFile)
		}

		for _, bucket := range irc.Buckets {
			tx.CreateBucketIfNotExists(bucket)
		}
		config = tx.Bucket(irc.BucketConfig)
		if err := selfSign(config); err != nil {
			return err
		}
		if !noPassword {
			if err := serverPass(config); err != nil {
				return err
			}
		}

		opers := tx.Bucket(irc.BucketOpers)
		if err := operPass(opers); err != nil {
			return err
		}
		return nil
	})
	return err
}

// https://golang.org/src/crypto/tls/generate_cert.go
func selfSign(config *bolt.Bucket) error {
	cert, key, err := security.SelfSignCert()
	if err != nil {
		return err
	}
	config.Put(irc.ConfigCert, cert)
	config.Put(irc.ConfigKey, key)
	return nil
}

func serverPass(config *bolt.Bucket) error {
	plaintext, err := security.RandomPassword()
	if err != nil {
		return err
	}
	salt, err := security.Salt()
	if err != nil {
		return err
	}
	pass := security.EncodePassword([]byte(plaintext), salt)
	config.Put(irc.ConfigPass, pass)
	config.Put(irc.ConfigSalt, salt)
	fmt.Printf("connection password is: %v\n", plaintext)
	return nil
}

func operPass(opers *bolt.Bucket) error {
	plaintext, err := security.RandomPassword()
	if err != nil {
		return err
	}
	salt, err := security.Salt()
	if err != nil {
		return err
	}
	pass := security.EncodePassword([]byte(plaintext), salt)
	oper, err := opers.CreateBucket(irc.DefaultOper)
	if err != nil {
		return err
	}
	oper.Put(irc.OperPass, pass)
	oper.Put(irc.OperSalt, salt)
	fmt.Printf("\nserver operator:\n\t/OPER irc %v\n", plaintext)
	return nil
}

// main for generating aes keys and encrypting passwords
// TODO review this process
package main

import (
	"encoding/base64"
	"encoding/hex"
	"flag"
	"fmt"
	"log"
	"os"
	"wrs/tk/packages/cryptography/aes"
	"wrs/tk/packages/cryptography/cli"

	"github.com/pkg/errors"
)

var ArgFormat string

var ArgEncrypt bool
var ArgDecrypt bool

var ArgKey string
var ArgGenerate bool

var ArgOutput string
var ArgInput string

func init() {
	log.SetFlags(log.LstdFlags | log.Lshortfile)

	flag.StringVar(&ArgFormat, "f", "", "Format")

	flag.BoolVar(&ArgEncrypt, "e", false, "Encrypt Message")
	flag.BoolVar(&ArgDecrypt, "d", false, "Decrypt Message")

	flag.StringVar(&ArgKey, "k", "", "Path to AES key")
	flag.BoolVar(&ArgGenerate, "g", false, "Generate AES key")

	flag.StringVar(&ArgOutput, "o", "", "Output file")
	flag.StringVar(&ArgInput, "i", "", "Input file")
}

func main() {
	flag.Parse()

	if ArgKey == "" {
		log.Fatal("Need to provide a file path to an existing or to be created key with the -k option")
	}

	if ArgGenerate {
		if num := cli.WriteAESKey(ArgKey, 64); num <= 0 {
			log.Fatal(fmt.Sprintf("Error generating AES key: %d bytes written\n", num))
		}
	}
	if ArgEncrypt && ArgDecrypt {
		fmt.Printf("Cannot simultaneously encrypt and decrypt a message")
		return
	}

	key, err := aes.KeyFromFile(ArgKey, 64)
	if err != nil {
		log.Fatalf("%+v", err)
	}

	if ArgEncrypt {
		encrypted, err := aes.Encrypt(key, []byte(flag.Arg(0)))
		if err != nil {
			log.Fatalf("%+v", err)
		}

		if ArgOutput == "" {
			switch ArgFormat {
			case "hex":
				fmt.Printf(encrypted.String())
			default:
				fmt.Printf("%s -> %x:%x + %x\n", flag.Arg(0), encrypted.Payload(), encrypted.Tag(), encrypted.Nonce())
			}
		} else {
			f, err := os.OpenFile(ArgOutput, os.O_CREATE|os.O_WRONLY, 0644)
			if err != nil {
				panic(err.Error())
			}
			defer f.Close()

			b := encrypted.Bytes()
			n, err := f.Write(b)
			if n != len(b) {
				log.Fatal(fmt.Sprintf("encrypted bytes is size %d, but %d bytes were written\n", len(b), n))
			}
			if err != nil {
				log.Fatalf("%+v", err)
			}
		}
	} else if ArgDecrypt {
		if ArgInput == "" {
			cipherText, err := hex.DecodeString(flag.Arg(0))
			if err != nil {
				log.Fatalf("%+v", errors.Wrap(err, "error decoding hex"))
			}

			plaintext, err := aes.Decrypt(key, aes.AESDataFromBytes(cipherText), 0)
			if ArgOutput == "" {
				fmt.Printf("%s -> %s\n", flag.Arg(0), plaintext)
			} else {
				f, err := os.OpenFile(ArgOutput, os.O_CREATE|os.O_WRONLY, 0644)
				if err != nil {
					panic(err.Error())
				}

				n, err := f.Write(plaintext)
				if n != len(plaintext) {
					log.Fatal(fmt.Sprintf("plaintext is size %d, but %d bytes were written\n", len(plaintext), n))
				}
				if err != nil {
					panic(err.Error())
				}
			}
		} else {
			rawbytes, err := os.ReadFile(ArgInput)
			if err != nil {
				log.Fatalf("%+v", errors.Wrap(err, "error reading input file"))
			}

			plaintext, err := aes.Decrypt(key, aes.AESDataFromBytes(rawbytes), 0)
			if err != nil {
				panic(err.Error())
			}

			if ArgOutput == "" {
				fmt.Printf("%x -> %s\n", rawbytes, plaintext)
			} else {
				f, err := os.OpenFile(ArgOutput, os.O_CREATE|os.O_WRONLY, 0644)
				if err != nil {
					panic(err.Error())
				}

				n, err := f.Write(plaintext)
				if n != len(plaintext) {
					log.Fatal(fmt.Sprintf("plaintext is size %d, but %d bytes were written\n", len(plaintext), n))
				}
				if err != nil {
					panic(err.Error())
				}
			}
		}
	}
}

// $in is a byte array that may or may not be encoded in hex or base 64
// returns $in parsed if neccesary
func parseBytes(in []byte) []byte {
	rawLen := len(in)
	var ret []byte

	//try hex
	ret = make([]byte, hex.DecodedLen(rawLen))
	_, err := hex.Decode(ret, in)
	if err != nil {
		//try base64
		ret = make([]byte, base64.StdEncoding.DecodedLen(rawLen))
		_, err = base64.StdEncoding.Decode(ret, in)
		if err != nil {
			//return raw
			return in
		}
		//decoded base64
		return ret
	}
	//decoded hex
	return ret
}

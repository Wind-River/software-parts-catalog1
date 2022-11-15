package cli

import (
	"bufio"
	"fmt"
	"log"
	"os"
	"wrs/tk/packages/cryptography/aes"
)

func WriteAESKey(keyPath string, base int) int {
	key := aes.GenerateKey()
	payload := key.Encode(64)

	f, err := os.Create(keyPath)
	if err != nil {
		log.Fatal(err)
	}
	defer f.Close()

	w := bufio.NewWriter(f)
	num, err := fmt.Fprintf(w, string(payload))
	if err != nil {
		log.Fatal(err)
	}
	w.Flush()

	return num
}

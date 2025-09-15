package main

import (
	"flag"
	"fmt"
	"log"

	"golang.org/x/crypto/bcrypt"
)

func main() {
	// ambil password dari argument CLI
	pass := flag.String("p", "", "Password yang mau di-hash")
	flag.Parse()

	if *pass == "" {
		log.Fatal("Gunakan flag -p untuk isi password, contoh: go run genpass.go -p admin")
	}

	// generate hash bcrypt
	hash, err := bcrypt.GenerateFromPassword([]byte(*pass), bcrypt.DefaultCost)
	if err != nil {
		log.Fatal(err)
	}

	fmt.Println("Password:", *pass)
	fmt.Println("Hash    :", string(hash))
}

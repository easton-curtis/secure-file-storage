package main

import (
	"bytes"
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"io"
	"os"

	"golang.org/x/crypto/pbkdf2"
)

func main() {
	//checks if there are more arguments than just the name of the program
	if len(os.Args) < 2 {
		help()
		os.Exit(0)
	}

	function := os.Args[1]

	//tests the function variable to determine what the user wants to do
	switch function {
	case "--help":
		help()
	case "--encrypt":
		encrypt()
	case "--decrypt":
		decrypt()
	default:
		help()
		os.Exit(0)
	}
}

func help() { //prints out help page
	fmt.Println("\nSecure File Storage Help Page")
	fmt.Println()
	fmt.Println("Usage:")
	fmt.Println("\tsfs --encrypt /path/to/file")
	fmt.Println("\tsfs --decrypt /path/to/file")
	fmt.Println("\tsfs --encrypt /path/to/file --no-pass")
	fmt.Println("\tsfs --decrypt /path/to/file --no-pass")
	fmt.Println("\tsfs --help")
	fmt.Println()
	fmt.Println("Commands:")
	fmt.Println("\t--encrypt\t encrypts raw data of file provided")
	fmt.Println("\t--decrypt\t decrypts encrypted data of file provided")
	fmt.Println("\t--no-pass\t bypasses need for password")
	fmt.Println("\t--help\t         displays help page")
}

func preliminaryProcessing() (string, []byte) {
	//checks if there is a path to a file
	if len(os.Args) < 3 {
		fmt.Println("Missing path to file")
		fmt.Println()
		help()
		os.Exit(0)
	}

	filePath := os.Args[2]

	if !validateFile(filePath) { //checks if file exists in path
		panic("File Not Found.")
	}

	var password []byte

	if len(os.Args) > 3 && os.Args[3] == "--no-pass" {
		password = make([]byte, 0)
	} else {
		password = getPassword()
	}

	if os.Args[3] == "--no-pass" {
		password = make([]byte, 0)
	}

	return filePath, password
}

func getRawData(fP string) []byte {
	fileData, err := os.Open(fP)
	if err != nil {
		panic(err.Error())
	}
	defer fileData.Close()

	plaintext, err := io.ReadAll(fileData)
	if err != nil {
		panic(err.Error())
	}
	return plaintext
}

func encrypt() {
	filePath, password := preliminaryProcessing()
	fmt.Println("\nEncryption has begun...")

	//create hash of password
	key := password

	//get raw data of file
	rawData := getRawData(filePath)

	//create salt
	salt := make([]byte, 12)
	if _, err := io.ReadFull(rand.Reader, salt); err != nil {
		panic(err.Error())
	}

	//get full key
	derivedKey := pbkdf2.Key(key, salt, 4096, 32, sha256.New)

	//create block for cipher
	block, err := aes.NewCipher(derivedKey)
	if err != nil {
		panic(err.Error())
	}

	aesgcm, err := cipher.NewGCM(block)
	if err != nil {
		panic(err.Error())
	}

	//create ciphertext
	ciphertext := aesgcm.Seal(nil, salt, rawData, nil)
	ciphertext = append(ciphertext, salt...)

	//copy ciphertext over original file's raw data
	destFile, err := os.Create(filePath)
	if err != nil {
		panic(err.Error())
	}
	defer destFile.Close()

	_, err = destFile.Write(ciphertext)
	if err != nil {
		panic(err.Error())
	}

	fmt.Println("Encryption is done.")
}

func decrypt() {
	filePath, password := preliminaryProcessing()
	fmt.Println("\nDecryption has begun...")

	key := password

	//get raw data of file
	rawData := getRawData(filePath)

	//get salt from end of ciphertext
	salt1 := rawData[len(rawData)-12:]
	str := hex.EncodeToString(salt1)
	salt, err := hex.DecodeString(str)
	if err != nil {
		panic(err.Error())
	}

	//create derived key
	derivedKey := pbkdf2.Key(key, salt, 4096, 32, sha256.New)

	//create block
	block, err := aes.NewCipher(derivedKey)
	if err != nil {
		panic(err.Error())
	}

	aesgcm, err := cipher.NewGCM(block)
	if err != nil {
		panic(err.Error())
	}

	//decrypt ciphertext
	plaintext, err := aesgcm.Open(nil, salt, rawData[:len(rawData)-12], nil)
	if err != nil {
		panic(err.Error())
	}

	//copy plaintext over encrypted file's raw data
	destFile, err := os.Create(filePath)
	if err != nil {
		panic(err.Error())
	}
	defer destFile.Close()

	_, err = destFile.Write(plaintext)
	if err != nil {
		panic(err.Error())
	}

	fmt.Println("Decryption is done.")
}

func validateFile(file string) bool {
	if _, err := os.Stat(file); os.IsNotExist(err) {
		return false
	}
	return true
}

func getPassword() []byte {
	//gets password from user
	fmt.Print("Password: ")
	var password string
	fmt.Scan(&password)
	//validates password from user
	fmt.Print("Validate Password: ")
	var vPassword string
	fmt.Scan(&vPassword)

	//convernt strings to []byte

	password1 := []byte(password)
	password2 := []byte(vPassword)

	//checks if both passwords match before returning
	if bytes.Equal(password1, password2) {
		return password1
	}
	fmt.Println("\n\nPasswords do not match. Try Again.")
	getPassword()

	//just to avoid error of not having return
	return make([]byte, 0)
}

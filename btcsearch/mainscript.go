package main

import (
	"bufio"
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"sync"
	"time"

	"crypto/sha256"

	"github.com/btcsuite/btcutil/base58"
	"golang.org/x/crypto/ripemd160"
	"gopkg.in/yaml.v2"
)

type Config struct {
	Threads      int    `yaml:"threads"`
	OutputFile   string `yaml:"output_file"`
	BTCAddresses string `yaml:"btc_addresses"`
}

func readConfig(filePath string) (*Config, error) {
	file, err := os.Open(filePath)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	var config Config
	decoder := yaml.NewDecoder(file)
	if err := decoder.Decode(&config); err != nil {
		return nil, err
	}

	return &config, nil
}

func readAddresses(filePath string) (map[string]bool, error) {
	addresses := make(map[string]bool)

	file, err := os.Open(filePath)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		addresses[scanner.Text()] = true
	}

	if err := scanner.Err(); err != nil {
		return nil, err
	}

	return addresses, nil
}

func generateKeyAndAddress() (string, string, error) {
	privateKey, err := ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
	if err != nil {
		return "", "", err
	}

	publicKey := privateKey.PublicKey
	address, err := publicKeyToAddress(publicKey)
	if err != nil {
		return "", "", err
	}

	return hex.EncodeToString(privateKey.D.Bytes()), address, nil
}

func publicKeyToAddress(publicKey ecdsa.PublicKey) (string, error) {
	pubKeyBytes := append(publicKey.X.Bytes(), publicKey.Y.Bytes()...)

	sha256Hash := sha256.New()
	sha256Hash.Write(pubKeyBytes)
	sha256Result := sha256Hash.Sum(nil)

	ripemd160Hash := ripemd160.New()
	ripemd160Hash.Write(sha256Result)
	ripemd160Result := ripemd160Hash.Sum(nil)

	networkVersion := byte(0x00)
	addressBytes := append([]byte{networkVersion}, ripemd160Result...)
	checksum := sha256Checksum(addressBytes)
	fullAddress := append(addressBytes, checksum...)

	return base58.Encode(fullAddress), nil
}

func sha256Checksum(input []byte) []byte {
	firstSHA := sha256.New()
	firstSHA.Write(input)
	result := firstSHA.Sum(nil)

	secondSHA := sha256.New()
	secondSHA.Write(result)
	finalResult := secondSHA.Sum(nil)

	return finalResult[:4]
}

func worker(id int, wg *sync.WaitGroup, mutex *sync.Mutex, outputFile string, btcAddresses map[string]bool) {
	defer wg.Done()

	for {
		privateKey, publicAddress, err := generateKeyAndAddress()
		if err != nil {
			log.Printf("Worker %d: Failed to generate key and address: %s", id, err)
			continue
		}

		matchFound := "No"
		if _, exists := btcAddresses[publicAddress]; exists {
			matchFound = "Yes"
			fmt.Printf("Match Found! Privatekey: %s Publicaddress: %s\n", privateKey, publicAddress)

			mutex.Lock()
			file, err := os.OpenFile(outputFile, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
			if err != nil {
				log.Printf("Worker %d: Failed to open file: %s", id, err)
				mutex.Unlock()
				continue
			}

			if _, err := file.WriteString(fmt.Sprintf("%s:%s\n", privateKey, publicAddress)); err != nil {
				log.Printf("Worker %d: Failed to write to file: %s", id, err)
			}
			file.Close()
			mutex.Unlock()
		}

		fmt.Printf("Private Key: %s Public Address: %s Match: %s\n", privateKey, publicAddress, matchFound)
		time.Sleep(100 * time.Millisecond) // Add a small delay to reduce CPU usage
	}
}

func main() {
	if len(os.Args) != 2 {
		fmt.Println("Usage: ./golangscript <config-file.yaml>")
		os.Exit(1)
	}

	executablePath, err := os.Executable()
	if err != nil {
		log.Fatalf("Failed to get executable path: %s", err)
	}

	executableDir := filepath.Dir(executablePath)
	if err := os.Chdir(executableDir); err != nil {
		log.Fatalf("Failed to change directory to executable path: %s", err)
	}

	configFile := os.Args[1]
	config, err := readConfig(configFile)
	if err != nil {
		log.Fatalf("Failed to read config file: %s", err)
	}

	btcAddresses, err := readAddresses(config.BTCAddresses)
	if err != nil {
		log.Fatalf("Failed to read BTC addresses: %s", err)
	}

	fmt.Printf("Loaded %d BTC addresses\n", len(btcAddresses))

	file, err := os.OpenFile(config.OutputFile, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		log.Fatalf("Failed to open output file: %s", err)
	}
	defer file.Close()

	for i := 0; i < 10; i++ {
		if _, err := file.WriteString(fmt.Sprintf("69exampleprivatekey%d:69examplepublicaddress%d\n", i, i)); err != nil {
			log.Fatalf("Failed to write to output file: %s", err)
		}
	}

	var wg sync.WaitGroup
	var mutex sync.Mutex

	for i := 0; i < config.Threads; i++ {
		wg.Add(1)
		go worker(i, &wg, &mutex, config.OutputFile, btcAddresses)
	}

	wg.Wait()
}

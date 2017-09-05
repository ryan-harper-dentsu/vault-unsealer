package main

import (
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"strconv"
	"strings"
	"syscall"
	"time"

	"golang.org/x/sys/unix"

	vault "github.com/hashicorp/vault/api"
)

const (
	DefaultListenAddr      string = ":443"
	DefaultPollingInterval int    = 1
)

// borrowed from https://github.com/hashicorp/vault/blob/24d2f39a7fd9f637fe745a107a2580eb891f0fb1/helper/mlock/mlock_unix.go
func lockMemory() {
	// Mlockall prevents all current and future pages from being swapped out.
	err := unix.Mlockall(syscall.MCL_CURRENT | syscall.MCL_FUTURE)
	if err != nil {
		log.Printf("Failed to lock memory! Do not use for production as unseal keys could be written to swap space!")
	}
}

var unsealKeys []string
var unsealThreshold int

func handlerStatus(w http.ResponseWriter, r *http.Request) {
	if len(unsealKeys) < unsealThreshold {
		http.Error(w, fmt.Sprintf("%d of %d required unseal keys. Requirement not met", len(unsealKeys), unsealThreshold), http.StatusInternalServerError)
		return
	}

	fmt.Fprintf(w, fmt.Sprintf("%d of %d required unseal keys. Ready to unseal!", len(unsealKeys), unsealThreshold), http.StatusInternalServerError)
}

func handlerAddKey(w http.ResponseWriter, r *http.Request) {
	contents, err := ioutil.ReadAll(r.Body)
	if err != nil {
		log.Println(err)
	}
	unsealKey := string(contents[:])

	if len(unsealKeys) >= unsealThreshold {
		http.Error(w, fmt.Sprintf("unseal key threshold of %d already met", unsealThreshold), http.StatusInternalServerError)
		return
	}

	unsealKeys = append(unsealKeys, strings.TrimSpace(unsealKey))
	fmt.Fprintf(w, "%d of %d required unseal keys", len(unsealKeys), unsealThreshold)
}

func startServer(listenAddr string, certPath string, certKeyPath string, pollingInterval int) {
	lockMemory()
	client, err := vault.NewClient(vault.DefaultConfig())
	if err != nil {
		log.Fatal(err)
	}
	sys := client.Sys()
	status, err := sys.SealStatus()
	if err != nil {
		log.Fatal(err)
	}
	unsealThreshold = status.T

	http.HandleFunc("/add-key", handlerAddKey)
	http.HandleFunc("/status", handlerStatus)
	go http.ListenAndServeTLS(listenAddr, certPath, certKeyPath, nil)
	log.Printf("vault-unsealer up and running on %s", listenAddr)

	for {
		status, err := sys.SealStatus()
		if err == nil {
			if status.Sealed == true && len(unsealKeys) == unsealThreshold {
				for num, unsealKey := range unsealKeys {
					log.Printf("Unsealing vault w/ unseal key #%d", num+1)
					_, err := sys.Unseal(unsealKey)
					if err != nil {
						log.Println(err)
					}
				}
			}
		} else {
			log.Println(err)
		}
		time.Sleep(time.Duration(pollingInterval) * time.Second)
	}
}

func main() {
	// set up defaults
	address := os.Getenv("LISTEN_ADDR")
	if address == "" {
		address = DefaultListenAddr
	}

	certPath := os.Getenv("HTTPS_CERT")
	if certPath == "" {
		log.Fatal("HTTPS_CERT must be set")
	}

	certKeyPath := os.Getenv("HTTPS_CERT_KEY")
	if certKeyPath == "" {
		log.Fatal("HTTPS_CERT_KEY must be set")
	}

	pollingInterval := DefaultPollingInterval
	if os.Getenv("POLLING_INTERVAL") != "" {
		var err error
		pollingInterval, err = strconv.Atoi(os.Getenv("POLLING_INTERVAL"))

		if err != nil {
			log.Fatal(err)
		}
	}

	startServer(address, certPath, certKeyPath, pollingInterval)
}

package main

import (
	"bytes"
	"crypto/tls"
	"flag"
	"fmt"
	"io/ioutil"
	"log"
	"net"
	"net/http"
	"net/url"
	"os"
	"path"
	"strconv"
	"strings"
	"syscall"
	"time"

	"github.com/howeyc/gopass"
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

func pollVault(pollingInterval int) {
	client, err := vault.NewClient(vault.DefaultConfig())
	if err != nil {
		log.Fatal(err)
	}
	sys := client.Sys()
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

func server() {

	lockMemory()

	// REQUIRED ENV VARIABLES
	certPath := os.Getenv("HTTPS_CERT")
	if certPath == "" {
		log.Fatal("HTTPS_CERT must be set")
	}

	certKeyPath := os.Getenv("HTTPS_CERT_KEY")
	if certKeyPath == "" {
		log.Fatal("HTTPS_CERT_KEY must be set")
	}

	// OPTIONAL ENV VARIABLES
	listenAddr := os.Getenv("LISTEN_ADDR")
	if listenAddr == "" {
		listenAddr = DefaultListenAddr
	}

	pollingInterval := DefaultPollingInterval
	if os.Getenv("POLLING_INTERVAL") != "" {
		var err error
		pollingInterval, err = strconv.Atoi(os.Getenv("POLLING_INTERVAL"))

		if err != nil {
			log.Fatal(err)
		}
	}

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
	go startServer(listenAddr, certPath, certKeyPath)
	pollVault(pollingInterval)
}

func startServer(listenAddr string, certPath string, certKeyPath string) {

	http.HandleFunc("/add-key", handlerAddKey)
	http.HandleFunc("/status", handlerStatus)
	listener, err := net.Listen("tcp", listenAddr)
	if err != nil {
		log.Fatal(err)
	}

	log.Printf("vault-unsealer listening on %s", listenAddr)
	log.Fatal(http.ServeTLS(listener, nil, certPath, certKeyPath))
}

// client function to add an unseal key to a server
func addKey(skipHostVerification bool) {

	fmt.Printf("Enter Unseal Key: ")

	// Silent. For printing *'s use gopass.GetPasswdMasked()
	unsealKey, err := gopass.GetPasswd()
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}

	tr := &http.Transport{
		TLSClientConfig: &tls.Config{InsecureSkipVerify: skipHostVerification},
	}
	client := &http.Client{Transport: tr}

	url := getFullURL("/add-key")
	resp, err := client.Post(url, "text/plain", bytes.NewReader(unsealKey))

	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
	response, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
	fmt.Printf(string(response))
}

// client function to get status from a server
func status(skipHostVerification bool) {

	tr := &http.Transport{
		TLSClientConfig: &tls.Config{InsecureSkipVerify: skipHostVerification},
	}
	client := &http.Client{Transport: tr}

	resp, err := client.Get(getFullURL("/status"))
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
	response, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
	fmt.Printf(string(response))
}

// helper function to get VAULT_UNSEALER_ADDR and join w/ a relative URL
func getFullURL(relativePath string) string {
	serverAddr := os.Getenv("VAULT_UNSEALER_ADDR")
	if serverAddr == "" {
		fmt.Fprintln(os.Stderr, "VAULT_UNSEALER_ADDR must be set")
		os.Exit(1)
	}

	u, err := url.Parse(serverAddr)
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}

	u.Path = path.Join(u.Path, relativePath)
	return u.String()
}

func main() {

	serverModePointer := flag.Bool("server", false, "start a vault-unsealer server")
	addKeyPointer := flag.Bool("add-key", false, "securely send an unseal key to a vault-unsealer server")
	statusPointer := flag.Bool("status", false, "view status of a vault-unsealer server")
	skipHostVerificationPointer := flag.Bool("skip-host-verification", false, "disable certificate check for client commands (FOR TESTING ONLY)")
	flag.Parse()

	if *serverModePointer == true {
		server()
	} else if *addKeyPointer == true {
		addKey(*skipHostVerificationPointer)
	} else if *statusPointer == true {
		status(*skipHostVerificationPointer)
	} else {
		flag.Usage()
		os.Exit(1)
	}
}

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
	"time"

	"github.com/howeyc/gopass"

	uuid "github.com/hashicorp/go-uuid"
	vault "github.com/hashicorp/vault/api"
	"github.com/hashicorp/vault/helper/mlock"
	"github.com/hashicorp/vault/helper/xor"
)

const (
	Version                string = "v0.1-alpha"
	DefaultListenAddr      string = ":443"
	DefaultPollingInterval int    = 1
)

var unsealKeys []string
var unsealThreshold int
var rootGenerationTested bool

func Log(handler http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		log.Printf("%s %s %s", r.RemoteAddr, r.Method, r.URL)
		handler.ServeHTTP(w, r)
	})
}

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

func tryGenerateRootToken() error {
	client, err := vault.NewClient(vault.DefaultConfig())
	sys := client.Sys()

	defer sys.GenerateRootCancel()

	// we use a generic OTP since the user will never see the token anyway
	// the token will be shortly thrown away
	otp := "G7gXEUyq+mguoMF7vq/xJw=="
	status, err := sys.GenerateRootInit(otp, "")
	if err != nil {
		return err
	}

	for num, unsealKey := range unsealKeys {
		log.Printf("Testing vault root token generation w/ unseal key #%d", num+1)
		var err error
		status, err = sys.GenerateRootUpdate(unsealKey, status.Nonce)
		if err != nil {
			return err
		}
	}

	// we need to decode the token w/ the OTP
	// https://github.com/hashicorp/vault/blob/master/command/generate-root.go
	tokenBytes, err := xor.XORBase64(status.EncodedRootToken, otp)
	if err != nil {
		return err
	}

	rootToken, err := uuid.FormatUUID(tokenBytes)
	if err != nil {
		return fmt.Errorf("Error formatting base64 token value: %v", err)
	}

	// revoke the root token w/ revoke-self
	rootTokenClient, err := vault.NewClient(vault.DefaultConfig())
	if err != nil {
		return err
	}
	rootTokenClient.SetToken(rootToken)
	rootTokenClient.Auth().Token().RevokeSelf(rootToken)
	if err != nil {
		return err
	}

	rootGenerationTested = true
	log.Println("successfullly tested unseal keys by generating & destroying a root token")
	return nil
}

func pollVault(pollingInterval int, tryGenerateRoot bool) {

	rootGenerationTested = false

	client, err := vault.NewClient(vault.DefaultConfig())
	if err != nil {
		log.Fatal(err)
	}
	sys := client.Sys()

	for {
		// not enough unseal keys to do anything yet
		if len(unsealKeys) != unsealThreshold {
			time.Sleep(time.Duration(pollingInterval) * time.Second)
			continue
		}

		// error talking to vault, log and move on
		status, err := sys.SealStatus()
		if err != nil {
			log.Println(err)
			time.Sleep(time.Duration(pollingInterval) * time.Second)
			continue
		}

		if status.Sealed == false {
			// if vault isn't sealed, we can optionally check the unseal keys work by generating root token
			if tryGenerateRoot && !rootGenerationTested {
				err := tryGenerateRootToken()
				if err != nil {
					log.Fatalln(err)
				}
			}
		} else {
			// TODO have a max retry policy if this happens repeatedly
			log.Println("Attempting to unseal vault (if this happens repeatedly a key is incorrect or repeated")
			for num, unsealKey := range unsealKeys {
				log.Printf("Unsealing vault w/ unseal key #%d", num+1)
				_, err := sys.Unseal(unsealKey)
				if err != nil {
					log.Fatalln(err)
				}
			}

		}

		time.Sleep(time.Duration(pollingInterval) * time.Second)
	}
}

func server() {

	if mlock.Supported() {
		mlock.LockMemory()
	} else {
		log.Println("Unable to lock memory! Unseal keys could be swapped to disk!")
	}

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

	// option to generate a root token and destroy it immediately after to
	// test that the unseal keys work
	var tryGenerateRoot bool
	tryGenerateRootString := os.Getenv("ROOT_TOKEN_TEST")
	if tryGenerateRootString != "" {
		var err error
		tryGenerateRoot, err = strconv.ParseBool(tryGenerateRootString)
		if err != nil {
			log.Fatal(err)
		}
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
	pollVault(pollingInterval, tryGenerateRoot)
}

func startServer(listenAddr string, certPath string, certKeyPath string) {

	http.HandleFunc("/add-key", handlerAddKey)
	http.HandleFunc("/status", handlerStatus)
	listener, err := net.Listen("tcp", listenAddr)
	if err != nil {
		log.Fatal(err)
	}

	log.Printf("vault-unsealer listening on %s", listenAddr)
	log.Fatal(http.ServeTLS(listener, Log(http.DefaultServeMux), certPath, certKeyPath))
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

func version() {
	fmt.Println(Version)
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

	// logs: print line number
	log.SetFlags(log.LstdFlags | log.Lshortfile)

	serverModePointer := flag.Bool("server", false, "start a vault-unsealer server")
	addKeyPointer := flag.Bool("add-key", false, "securely send an unseal key to a vault-unsealer server")
	statusPointer := flag.Bool("status", false, "view status of a vault-unsealer server")
	versionPointer := flag.Bool("version", false, "show version")
	skipHostVerificationPointer := flag.Bool("skip-host-verification", false, "disable TLS certificate check for client commands (FOR TESTING PURPOSES ONLY)")
	flag.Parse()

	if *serverModePointer == true {
		server()
	} else if *addKeyPointer == true {
		addKey(*skipHostVerificationPointer)
	} else if *statusPointer == true {
		status(*skipHostVerificationPointer)
	} else if *versionPointer == true {
		version()
	} else {
		flag.Usage()
		os.Exit(1)
	}
}

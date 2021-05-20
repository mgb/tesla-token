package main

import (
	"bufio"
	"fmt"
	"log"
	"math/rand"
	"os"
	"strings"
	"syscall"
	"time"

	"github.com/mgb/tesla-token/pkg/tesla"
	"github.com/spf13/pflag"
	"golang.org/x/crypto/ssh/terminal"
)

func main() {
	username := pflag.StringP("username", "u", "", "email address for login")
	password := pflag.StringP("password", "p", "", "password for login (leave blank for prompt)")
	authcode := pflag.StringP("authcode", "c", "", "second factor auth code (leave blank for prompt, if one is detected)")

	saveHTML := pflag.Bool("save-html", false, "for debugging, will output HTML files the tool fetches")
	saveDir := pflag.String("save-html-dir", "", "the directory to store HTML outputs")

	pflag.Parse()

	p := tesla.Params{
		SaveHTML: *saveHTML,
		SaveDir:  *saveDir,
	}

	ownerAPI, err := tesla.New(p)
	if err != nil {
		log.Fatal(err)
	}

	// Allow for re-prompting if the user got it wrong
	usernameFn := func() string {
		if *username != "" {
			u := *username
			*username = ""
			return u
		}

		u := os.Getenv("TESLA_USERNAME")
		if u != "" {
			return u
		}

		return promptForString("Username/Email")
	}
	passwordFn := func() string {
		if *password != "" {
			p := *password
			*password = ""
			return p
		}

		p := os.Getenv("TESLA_PASSWORD")
		if p != "" {
			return p
		}

		return promptForPassword("Password")
	}
	authcodeFn := func() string {
		if *authcode != "" {
			c := *authcode
			*authcode = ""
			return c
		}

		return promptForString("MFA Auth Code")
	}

	accessToken, _, err := ownerAPI.Login(usernameFn, passwordFn, authcodeFn)
	if err != nil {
		log.Fatal(err)
	}

	ownerToken, err := ownerAPI.GetOwnerToken(accessToken)
	if err != nil {
		log.Fatal(err)
	}

	fmt.Println(ownerToken)
}

func init() {
	rand.Seed(time.Now().UnixNano())
}

func promptForString(action string) string {
	reader := bufio.NewReader(os.Stdin)

	fmt.Printf("Enter %s: ", action)
	result, err := reader.ReadString('\n')
	if err != nil {
		log.Fatal(err)
	}

	return strings.TrimSpace(result)
}

func promptForPassword(action string) string {
	fmt.Printf("Enter %s: ", action)
	bytePassword, err := terminal.ReadPassword(int(syscall.Stdin))
	if err != nil {
		fmt.Println()
		log.Fatal(err)
	}

	password := strings.TrimSpace(string(bytePassword))

	// ReadPassword consumes the newline, so print something showing we read something
	if password == "" {
		fmt.Println()
	} else {
		fmt.Println("********")
	}

	return password
}

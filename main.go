package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"

	"github.com/emersion/go-imap"
	"github.com/emersion/go-imap/client"
	"github.com/spf13/viper"
)

func main() {
	// コンフィグ
	viper.SetConfigName("config") // name of config file (without extension)
	viper.AddConfigPath(".")      // optionally look for config in the working directory
	err := viper.ReadInConfig()   // Find and read the config file
	if err != nil {               // Handle errors reading the config file
		panic(fmt.Errorf("Fatal error config file: %s \n", err))
	}

	http.HandleFunc("/fetch", fetchImap)
	http.ListenAndServe(":6993", nil)
}

func fetchImap(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Cache-Control", "no-cache")

	r.ParseForm()

	user := r.Form.Get("user")
	pass := r.Form.Get("pass")
	log.Println("user:", user)
	log.Println("pass:", pass)

	log.Println("Connecting to server...")

	// Connect to server
	// c, err := client.DialTLS("imap.gmail.com:993", nil)
	c, err := client.DialTLS(viper.GetString("imapServer"), nil)
	if err != nil {
		log.Fatal(err)
	}
	log.Println("Connected")

	// Don't forget to logout
	defer c.Logout()

	// Login
	if err := c.Login(user, pass); err != nil {
		log.Fatal(err)
	}
	log.Println("Logged in")

	// List mailboxes
	// mailboxes := make(chan *imap.MailboxInfo, 10)
	// done := make(chan error, 1)
	// go func() {
	// 	done <- c.List("", "*", mailboxes)
	// }()

	// log.Println("Mailboxes:")
	// for m := range mailboxes {
	// 	log.Println("* " + m.Name)
	// }

	// if err := <-done; err != nil {
	// 	log.Fatal(err)
	// }

	// Select INBOX
	mbox, err := c.Select("INBOX", false)
	if err != nil {
		log.Fatal(err)
	}
	log.Println("Flags for INBOX:", mbox.Flags)

	// Get the last 49 messages
	from := uint32(1)
	to := mbox.Messages
	if mbox.Messages > 49 {
		// We're using unsigned integers here, only substract if the result is > 0
		from = mbox.Messages - 49
	}
	seqset := new(imap.SeqSet)
	seqset.AddRange(from, to)

	messages := make(chan *imap.Message, 10)
	done := make(chan error, 1)
	go func() {
		done <- c.Fetch(seqset, []string{imap.EnvelopeMsgAttr}, messages)
	}()

	log.Println("Last 50 messages:")
	msgs := []imap.Envelope{}
	for msg := range messages {
		log.Println("* " + msg.Envelope.Subject)
		msgs = append(msgs, *msg.Envelope)
	}

	// json 形式に変換して返す
	jsonBytes, err := json.Marshal(msgs)
	if err != nil {
		fmt.Println("JSON Marshal error:", err)
		return
	}

	log.Println(string(jsonBytes))
	w.Write(jsonBytes)
	w.(http.Flusher).Flush()

	if err := <-done; err != nil {
		log.Fatal(err)
	}

	log.Println("Done!")
}

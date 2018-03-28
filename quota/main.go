package main
import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"

	"github.com/mxk/go-imap/imap"
	"github.com/spf13/viper"
)

var (
    c   *imap.Client
    cmd *imap.Command
	rsp *imap.Response
)

func main() {
	// コンフィグ
	viper.SetConfigName("config") // name of config file (without extension)
	viper.AddConfigPath(".")      // optionally look for config in the working directory
	err := viper.ReadInConfig()   // Find and read the config file
	if err != nil {               // Handle errors reading the config file
		panic(fmt.Errorf("Fatal error config file: %s", err))
	}

	http.HandleFunc("/quota", getQuota)
	http.ListenAndServe(":6993", nil)
}

func getQuota(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Cache-Control", "no-cache")

	r.ParseForm()

	user := r.Form.Get("user")
	pass := r.Form.Get("pass")
	log.Println("user:", user)
	log.Println("pass:", pass)

	log.Println("Connecting to server...")

	// Connect to server
	c, _ = imap.DialTLS(viper.GetString("imapServer"), nil)
	log.Println("Connected")
	defer c.Logout(3000)

	// Authenticate
	if c.State() == imap.Login {
		c.Login(user, pass)
	}

	cmd, _ = imap.Wait(c.GetQuotaRoot("INBOX"))
	fmt.Println("cmd:", cmd)
	fmt.Println("\ncmd data:", cmd.Data)
	for _, rsp = range cmd.Data {
		_, quota := rsp.Quota()
		if quota == nil { continue }
		jsonBytes, _ := json.Marshal(quota)
		log.Println(string(jsonBytes))
		w.Write(jsonBytes)
	}

	w.(http.Flusher).Flush()

	log.Println("Done!")


}


package gotunl

import (
	"bytes"
	"context"
	b64 "encoding/base64"
	"fmt"
	"io/ioutil"
	"log"
	"net"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"

	"github.com/howeyc/gopass"
	"github.com/tidwall/gjson"
	"github.com/tidwall/sjson"
)

type profile struct {
	Path string
	ID   int
	Conf string
}
type Gotunl struct {
	authKey    string
	profPath   string
	service    string
	unixSocket string
	Profiles   map[string]profile
}

func _getKey() string {
	keyPath := ""
	if runtime.GOOS == "windows" {
		keyPath = "c:\\ProgramData\\Pritunl\\auth"
	} else {
		keyPath = "/var/run/pritunl.auth"
	}
	if _, err := os.Stat(keyPath); !os.IsNotExist(err) {
		key, err := ioutil.ReadFile(keyPath)
		if err != nil {
			log.Fatalf("Error getting key: %s\n", err)
		}
		return string(key)
	}
	return ""
}

func _getProfilePath() string {
	home := ""
	profPath := ""
	switch oS := runtime.GOOS; oS {
	case "darwin":
		home = os.Getenv("HOME")
		profPath = home + "/Library/Application Support/pritunl/profiles"
	case "windows":
		home = os.Getenv("APPDATA")
		profPath = home + "\\pritunl\\profiles"
	case "linux":
		home = os.Getenv("HOME")
		profPath = home + "/.config/pritunl/profiles"
	}
	if _, err := os.Stat(profPath); !os.IsNotExist(err) {
		return profPath
	}
	return ""
}

func New() *Gotunl {
	service := ""
	if runtime.GOOS == "linux" || runtime.GOOS == "darwin" {
		service = "http://unix/"
	} else {
		service = "http://localhost:9770/"
	}
	g := Gotunl{_getKey(), _getProfilePath(), service, "/var/run/pritunl.sock", map[string]profile{}}
	g.loadProfiles()
	return &g
}

func (g Gotunl) makeReq(verb string, endpoint string, data string) string {
	url := g.service + endpoint
	req, err := http.NewRequest(verb, url, bytes.NewBuffer([]byte(data)))
	if err != nil {
		log.Fatalf("Error making request: %s\n", err)
	}
	req.Header.Set("User-Agent", "pritunl")
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Auth-Key", g.authKey)
	client := http.Client{}
	// pritunl now uses unix sockets for linux and MacOS.
	// ref: https://gist.github.com/teknoraver/5ffacb8757330715bcbcc90e6d46ac74
	if runtime.GOOS == "linux" || runtime.GOOS == "darwin" {
		client = http.Client{
			Transport: &http.Transport{
				DialContext: func(_ context.Context, _, _ string) (net.Conn, error) {
					return net.Dial("unix", g.unixSocket)
				},
			},
		}
	}
	res, err := client.Do(req)
	if err != nil {
		log.Fatalf("Error making request (do): %s\n", err)
	}
	if res.StatusCode == 200 {
		body, _ := ioutil.ReadAll(res.Body)
		res.Body.Close()
		return string(body)
	}
	return string(res.StatusCode)

}

func (g Gotunl) CheckStatus() string {
	s := g.makeReq("GET", "status", "")
	return gjson.Get(s, "status").String()
}

func (g Gotunl) Ping() bool {
	p := g.makeReq("GET", "ping", "")
	return p == ""
}

func (g Gotunl) GetConnections() string {
	cons := g.makeReq("GET", "profile", "")
	return cons
}

func (g Gotunl) StopConnections() {
	g.makeReq("POST", "stop", "")
}

func (g Gotunl) loadProfiles() {
	res, _ := filepath.Glob(g.profPath + "/*.conf")
	for i, f := range res {
		c := i + 1
		prof := strings.Split(filepath.Base(f), ".")[0]
		conf, err := ioutil.ReadFile(f)
		if err != nil {
			log.Fatalf("Error loading profiles: %s\n", err)
		}
		config := string(conf)                        // keep the whole config file to use later, instead of reading the file again.
		if gjson.Get(config, "name").String() == "" { // If "name": null it will set the name automatically.
			user := gjson.Get(config, "user").String()
			server := gjson.Get(config, "server").String()
			config, _ = sjson.Set(config, "name", fmt.Sprintf("%v (%v)", user, server))
		}
		g.Profiles[prof] = profile{f, c, config}
	}
}

func (g Gotunl) GetProfile(id string) (string, string) {
	auth := ""
	key := ""
	g.loadProfiles()
	prof := g.Profiles[id]
	ovpnFile := strings.Replace(prof.Path, id+".conf", id+".ovpn", 1)
	ovpn, err := ioutil.ReadFile(ovpnFile)
	if err != nil {
		log.Fatalf("Error getting profile: %s\n", err)
	}
	for _, l := range strings.Split(string(ovpn), "\n") {
		if strings.Contains(l, "auth-user-pass") && len(l) <= 17 { //check if it needs credentials and they are not provided as parameter
			auth = "creds"
		}
	}
	mode := gjson.Get(prof.Conf, "password_mode").String()
	if auth != "" && strings.Contains(prof.Conf, "password_mode") && mode != "" {
		auth = mode
	}
	if runtime.GOOS == "darwin" {
		command := "security find-generic-password -w -s pritunl -a " + id
		out, err := exec.Command("bash", "-c", command).Output()
		if err != nil {
			if strings.Contains("exit status 36", err.Error()) {
				log.Println("There was an error accessing the Keychain (probably connected through SSH)")
				log.Fatal("Run '/usr/bin/security unlock-keychain' to unlock the Keychain and try again")
			}
			log.Fatalf("Error getting profiles (find-generic-password): %s\n", err)
		}
		res, err := b64.StdEncoding.DecodeString(string(out))
		if err != nil {
			log.Fatalf("Error decoding base64: %s\n", err)
		}
		key = string(res)
	}
	vpn := string(ovpn) + "\n" + key
	return vpn, auth

}

func (g Gotunl) ConnectProfile(id string, user string, password string) {
	data := fmt.Sprintf(`{"id": "%v", "reconnect": true, "timeout": true}`, id)
	ovpn, auth := g.GetProfile(id)
	if (auth != "") && (user == "" || password == "") {
		auth_method := auth[len(auth)-3:]
		if auth_method == "otp" || auth_method == "pin" {
			var otp string
			user = "pritunl"
			if password == "" {
				fmt.Printf("Enter the PIN: ")
				pass, err := gopass.GetPasswdMasked()
				if err != nil {
					log.Fatalf("Error connecting to profile (GetPasswdMasked): %s\n", err)
				}
				if auth == "otp_pin" {
					fmt.Printf("Enter the OTP code: ")
					fmt.Scanln(&otp)
				}
				password = string(pass) + otp
			}
		}
		if user == "" {
			fmt.Printf("Enter the username: ")
			fmt.Scanln(&user)

		}
		if password == "" {
			fmt.Printf("Enter the password: ")
			pass, _ := gopass.GetPasswdMasked()
			password = string(pass)
		}
	}
	data, _ = sjson.Set(data, "username", user)
	data, _ = sjson.Set(data, "password", password)
	data, _ = sjson.Set(data, "data", ovpn)
	g.makeReq("POST", "profile", data)
}

func (g Gotunl) DisconnectProfile(id string) {
	g.makeReq("DELETE", "profile", `{"id": "`+id+`"}`)
}

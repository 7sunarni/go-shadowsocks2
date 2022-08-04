package main

import (
	"encoding/json"
	"io/ioutil"
	"log"
	"net"
	"net/http"
	"os"
	"os/exec"
	"sync"
)

var (
	lock       = sync.Mutex{}
	allowedIPs = map[string]string{}
)

func updateIPTables() error {
	var err error
	{
		c := exec.Command("sudo", "iptables", "-F")
		c.Stdout = os.Stdout
		c.Stderr = os.Stderr
		if err := c.Run(); err != nil {
			return err
		}
	}

	{
		c := exec.Command("sudo", "iptables", "-I", "INPUT", "-p", "tcp", "--dport", "58080", "-j", "REJECT")
		if err := c.Run(); err != nil {
			return err
		}
	}

	for name, ip := range allowedIPs {
		c := exec.Command("sudo", "iptables", "-I", "INPUT", "--source", ip, "-p", "tcp", "--dport", "58080", "-j", "ACCEPT")
		if err = c.Run(); err != nil {
			return err
		}
		log.Printf("add allowed %s ip %s", name, ip)
	}

	return nil
}

func Get(rw http.ResponseWriter, r *http.Request) {
	lock.Lock()
	defer lock.Unlock()
	rw.WriteHeader(http.StatusOK)
	json.NewEncoder(rw).Encode(allowedIPs)
}

type Payload struct {
	IP string
}

type ApiConfig struct {
	Items []struct {
		Path string `json:"path"`
		Name string `json:"name"`
	} `json:"items"`
}

func Put(rw http.ResponseWriter, r *http.Request) {
	lock.Lock()
	defer lock.Unlock()
	p := new(Payload)
	if err := json.NewDecoder(r.Body).Decode(p); err != nil {
		rw.WriteHeader(http.StatusBadRequest)
		rw.Write([]byte(err.Error()))
		return
	}
	ip := string(p.IP)
	allowedIPs["whitelist"] = ip
	updateIPTables()
	rw.WriteHeader(http.StatusOK)
	rw.Write([]byte(ip))

}

func StartHTTP(p string) {

	data, err := ioutil.ReadFile("config.json")
	if err != nil {
		log.Panic(err)
	}
	c := new(ApiConfig)
	if err := json.Unmarshal(data, c); err != nil {
		log.Panic(err)
	}

	for i := range c.Items {
		item := c.Items[i]
		http.HandleFunc(item.Path, func(rw http.ResponseWriter, r *http.Request) {
			lock.Lock()
			defer lock.Unlock()
			remoteIP, _, err := net.SplitHostPort(r.RemoteAddr)
			if err != nil {
				rw.WriteHeader(http.StatusInternalServerError)
				rw.Write([]byte(err.Error()))
				return
			}
			allowedIPs[item.Name] = remoteIP
			updateIPTables()
			rw.WriteHeader(http.StatusOK)
			rw.Write([]byte(remoteIP))
		})
	}

	http.HandleFunc("/api/v1/whitelist", func(rw http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case http.MethodGet:
			Get(rw, r)
			return
		case http.MethodPost:
			Put(rw, r)
			return
		default:
			rw.WriteHeader(http.StatusMethodNotAllowed)
			return
		}
	})
	resp, err := http.Get("https://ifconfig.me")
	if err != nil {
		return
	}
	respB, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return
	}
	allowedIPs["local"] = string(respB)
	http.ListenAndServe(p, nil)
}

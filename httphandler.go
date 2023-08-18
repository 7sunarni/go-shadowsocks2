package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net"
	"net/http"
	"os"
	"os/exec"
	"strings"
	"sync"
)

var (
	lock       = sync.Mutex{}
	allowedIPs = map[string]string{}
	port       = "58080"
)

func runIptables(command string) error {
	splitedCommands := strings.Split(command, " ")
	c := exec.Command("sudo", splitedCommands...)
	c.Stdout = os.Stdout
	c.Stderr = os.Stderr
	return c.Run()
}

func updateIPTables() error {
	if err := runIptables("iptabels -F"); err != nil {
		return err
	}
	/*
	iptable reject with tcp-reset option
	use tcpdump observe tcp connection behavior: 
 
        1. port is not listening
	client [S]
        client [S]
	server [R.]

        2. iptables -J DROP
	client [S]
        client [S]
	client [S]
        client [S]
	server [R.]

        3. iptables -J REJECT --reject-with icmp-port-unreachable
	client [S]
        client [S]
	client [S]
        client [S]
	server [R.]

        4. iptables -J REJECT --reject-with tcp-reset
 	client [S]
        client [S]
	server [R.]

        reject-with tcp-reset behave more like port is not listening.
	TODO: check tcp reset cause reason
 
        */
	if err := runIptables(fmt.Sprintf("iptables -I INPUT -p tcp --dport %s -j REJECT --reject-with tcp-reset", port)); err != nil {
		return err
	}

	for name, ip := range allowedIPs {
		if err := runIptables(fmt.Sprintf("iptables -I INPUT --source %s -p tcp --dport %s -j ACCEPT -m comment --comment %s", ip, port, name)); err != nil {
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
			// not update iptables
			if allowedIPs[item.Name] == remoteIP {
				rw.Write([]byte(remoteIP))
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

# !/bin/bash

# $1 password
curl 'https://dl.google.com/go/go1.17.6.linux-amd64.tar.gz' -o go1.17.6.linux-amd64.tar.gz

rm -rf /usr/local/go && tar -C /usr/local -xzf go1.17.6.linux-amd64.tar.gz

export PATH=$PATH:/usr/local/go/bin

curl 'https://codeload.github.com/EnnnOK/go-shadowsocks2/zip/refs/heads/master' -o temp.zip

unzip temp.zip 

cd go-shadowsocks2-master 

/usr/local/go/bin/go build -o go-ss .

export SHADOWSOCKS_SF_CAPACITY=1e6 SHADOWSOCKS_SF_FPR=1e-6 SHADOWSOCKS_SF_SLOT=10
nohup ./go-ss -s ss://CHACHA20-IETF-POLY1305:${PASSWORD}@:58080 -verbose &

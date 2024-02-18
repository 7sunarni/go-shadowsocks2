# !/bin/bash

# $1 password
curl 'https://dl.google.com/go/go1.17.6.linux-amd64.tar.gz' -o go1.17.6.linux-amd64.tar.gz

rm -rf /usr/local/go && tar -C /usr/local -xzf go1.17.6.linux-amd64.tar.gz

export PATH=$PATH:/usr/local/go/bin

curl 'https://codeload.github.com/EnnnOK/go-shadowsocks2/zip/refs/heads/master' -o temp.zip

unzip temp.zip

mv go-shadowsocks2-master go-shadowsocks2

rm temp.zip

cd go-shadowsocks2

/usr/local/go/bin/go build -o go-ss .

cat > /lib/systemd/system/go-ss.service << eof
[Unit]
Description=My Miscellaneous Service
After=network.target

[Service]
Type=simple
Environment="SHADOWSOCKS_SF_CAPACITY=1e6"
Environment="SHADOWSOCKS_SF_FPR=1e-6"
Environment="SHADOWSOCKS_SF_SLOT=10"
WorkingDirectory=/root/workspace/go-shadowsocks2
ExecStart=/root/workspace/go-shadowsocks2/go-ss -s ss://CHACHA20-IETF-POLY1305:hello-ssh@:58080 -verbose
Restart=on-failure # or always, on-abort, etc

[Install]
WantedBy=multi-user.target
eof

systemctl daemon-reload
systemctl start go-ss.service

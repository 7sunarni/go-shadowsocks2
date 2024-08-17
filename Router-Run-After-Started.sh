#!/bin/sh

### Custom user script
### Called after router started and network is ready

### Example - load ipset modules
#modprobe ip_set
#modprobe ip_set_hash_ip
#modprobe ip_set_hash_net
#modprobe ip_set_bitmap_ip
#modprobe ip_set_list_set
#modprobe xt_set
mkdir -p /tmp/pingtunnel;
server=""
curl -k https://$server:443/pingtunnel -o /tmp/pingtunnel/pingtunnel;
curl -k https://$server:443/GeoLite2-Country.mmdb -o /tmp/pingtunnel/GeoLite2-Country.mmdb;
chmod +x /tmp/pingtunnel/pingtunnel;
cd /tmp/pingtunnel;
pkill dnsmasq;
(./pingtunnel -type client -l :53 -s $server -t 8.8.8.8:53  -nolog 1 -noprint 1 )&
cd /tmp/pingtunnel;
(./pingtunnel -type client -l :4455 -s $server -sock5 1 -ltcp :4466 -tcp_bs 4096 -lhttp :4477 -s5ftfile /tmp/pingtunnel/GeoLite2-Country.mmdb -s5filter CN 2>1)&


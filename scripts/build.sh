#!/bin/bash

# clean
rm -rf output/ && mkdir -p output/bin/ && mkdir -p output/log/

# build servers
go build -v -o output/bin/net_reciever ./frameworks/net
go build -v -o output/bin/nbio_reciever ./frameworks/nbio
go build -v -o output/bin/gnet_reciever ./frameworks/gnet
go build -v -o output/bin/netpoll_reciever ./frameworks/netpoll
go build -v -o output/bin/client_bencher ./client


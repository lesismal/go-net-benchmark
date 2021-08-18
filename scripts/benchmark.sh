#!/bin/bash

. ./scripts/env.sh

repo=("net" "nbio" "gnet" "netpoll" "easygo")
ports=(8001 8002 8003 8004 8005)
rpcports=(9001 9002 9003 9004 9005)

. ./scripts/build.sh

c=$1

# benchmark
for b in ${body[@]}; do
  for ((i = 0; i < ${#repo[@]}; i++)); do
    rp=${repo[i]}
    addr="127.0.0.1:${ports[i]}"
    rpc="127.0.0.1:${rpcports[i]}"
    # server start
    nohup $taskset_server ./output/bin/${rp}_reciever >> output/log/nohup.log 2>&1 &
    sleep 1
    echo "server $rp running with $taskset_server"

    # run client
    echo "client $rp running with $taskset_client"
    $taskset_client ./output/bin/client_bencher -addr="$addr" -r=${rpc} -f=${rp} -b=$b -c=$c -n=$n

    # stop server
    pid=$(ps -ef | grep ${rp}_reciever | grep -v grep | awk '{print $2}')
    disown $pid
    kill -9 $pid
    sleep 1
  done
done

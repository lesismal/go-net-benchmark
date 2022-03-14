#!/bin/bash

. ./scripts/env.sh

repo=("nbio" "gnet" "gnetv2" "easygo" "netpoll" "net")
ports=(8001 8002 8003 8004 8005 8006)
rpcports=(9001 9002 9003 9004 9005 9006)

. ./scripts/build.sh

c=$1

# benchmark
for b in ${body[@]}; do
  for ((i = 0; i < ${#repo[@]}; i++)); do
    rp=${repo[i]}
    port=${ports[i]}
    rpcport=${rpcports[i]}
    # server start
    nohup $taskset_server ./output/bin/${rp}_reciever -p=${port} -r=${rpcport} >> output/log/nohup.log 2>&1 &
    sleep 1
    echo "server $rp running with $taskset_server"

    # run client
    echo "client $rp running with $taskset_client"
    $taskset_client ./output/bin/client_bencher -p=${port} -r=${rpcport} -f=${rp} -b=$b -c=$c -n=$n

    # stop server
    pid=$(ps -ef | grep ${rp}_reciever | grep -v grep | awk '{print $2}')
    disown $pid
    kill -9 $pid
    sleep 1
  done
done

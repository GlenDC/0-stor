#!/bin/bash
NUMBER_OF_SERVERS=$1
IYO_APP_ID=$2
IYO_APP_SECRET=$3
ORGANIZATION=$4
NAMESPACE=$5

check_arguments(){
    if [ "$#" != "5" ]; then
        echo """
This scirpt will install all zstordb dependencies and run #NUMBER_OF_SERVERS zstordb in the local host
starting from 8080 port, then upate the zstordb config file.

Usage:
    bash manual_deployment.sh <NUMBER_OF_SERVERS> <IYO_APP_ID> <IYO_APP_SECRET> <ORGANIZATION> <NAMESPACE>
         """
         exit 1
    fi
}

update_env(){
    apt-get update
	apt-get install -y curl net-tools git
}

install_go(){
	curl https://storage.googleapis.com/golang/go1.9.linux-amd64.tar.gz > go1.9.linux-amd64.tar.gz
	tar -C /usr/local -xzf go1.9.linux-amd64.tar.gz
	echo "export PATH=$PATH:/usr/local/go/bin" >> ~/.bash_profile
	echo "export GOROOT=/usr/local/go" >> ~/.bash_profile
 	echo "export GOPATH=/gopath" >> ~/.bash_profile
	mkdir /gopath
	source ~/.bash_profile
}

install_etcd(){
    ETCD_VER=v3.2.10

    # choose either URL
    GOOGLE_URL=https://storage.googleapis.com/etcd
    GITHUB_URL=https://github.com/coreos/etcd/releases/download
    DOWNLOAD_URL=${GOOGLE_URL}

    rm -f /tmp/etcd-${ETCD_VER}-linux-amd64.tar.gz
    rm -rf /tmp/etcd-download-test && mkdir -p /tmp/etcd-download-test

    curl -L ${DOWNLOAD_URL}/${ETCD_VER}/etcd-${ETCD_VER}-linux-amd64.tar.gz -o /tmp/etcd-${ETCD_VER}-linux-amd64.tar.gz
    tar xzvf /tmp/etcd-${ETCD_VER}-linux-amd64.tar.gz -C /tmp/etcd-download-test --strip-components=1
    rm -f /tmp/etcd-${ETCD_VER}-linux-amd64.tar.gz
}

run_etcd(){
    /tmp/etcd-download-test/etcd --advertise-client-urls  http://0.0.0.0:2379 --listen-client-urls http://0.0.0.0:2379 &
    echo " -------------------- ETCD CLUSTER -------------------- "
    echo "ETCD Cluster : http://0.0.0.0:2379"
}

install_zstor_server(){
    mkdir -p /gopath/src/github.com
    cp -ra /home/travis/build/zero-os /gopath/src/github.com
    cd /gopath/src/github.com/zero-os/0-stor
    git config remote.origin.fetch +refs/heads/*:refs/remotes/origin/*
    git fetch
    git checkout -f ${ZSTORDB_BRANCH}
    echo " [*] Install zerostor client from branch : ${ZSTORDB_BRANCH}"
    make .
    chmod 777 /gopath/src/github.com/zero-os/0-stor/bin
    ln -sf /gopath/src/github.com/zero-os/0-stor/bin/zstordb /bin/zstordb
    ln -sf /gopath/src/github.com/zero-os/0-stor/bin/zstor /bin/zstor    
    hash -r
    cd /home/travis/build/zero-os/0-stor
    git config remote.origin.fetch +refs/heads/*:refs/remotes/origin/*
    git fetch
    git checkout -f ${TESTCASE_BRANCH}    
    echo " [*] Execute test cases from branch : ${ZSTORDB_BRANCH}"
    rm -rf /zstor
    mkdir /zstor
}

run_zstor_server(){
    echo "data_shards:" > data_shards
    for ((i=0; i<$NUMBER_OF_SERVERS; i++)); do
        port=$((8080+$i))
        zstordb -L 0.0.0.0:$port --meta-dir /zstor/meta_$port --data-dir /zstor/data_$port &
        echo " -------------------- zstor -------------------- "
        echo "ZSTOR SERVER $i : 0.0.0.0:$port"
        echo -e "    - 127.0.0.1:$port" >> data_shards
    done    
}

update_zsrordb_config_file(){
    echo """organization: $ORGANIZATION
namespace: $NAMESPACE
iyo_app_id: $IYO_APP_ID
iyo_app_secret: $IYO_APP_SECRET
protocol: grpc

meta_shards:
    - http://127.0.0.1:2379

block_size: 4096
replication_nr: $NUMBER_OF_SERVERS
replication_max_size: 4096

distribution_data: $(($NUMBER_OF_SERVERS-1))
distribution_parity: 1

compress: true
encrypt: true
encrypt_key: ab345678901234567890123456789012
""" > /gopath/src/github.com/zero-os/0-stor/cmd/zstor/config.yaml
cat data_shards >> /gopath/src/github.com/zero-os/0-stor/cmd/zstor/config.yaml
rm -rf data_shards
}

check_arguments $1 $2 $3 $4 $5
update_env
install_go
install_etcd
install_zstor_server

run_etcd
run_zstor_server
update_zsrordb_config_file

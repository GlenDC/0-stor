#!/bin/bash
NUMBER_OF_SERVERS=$1

update_env(){
    apt-get update
	apt-get install -y curl vim tmux net-tools git
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
    go get -u github.com/zero-os/0-stor/cmd/zstordb
    ln -sf /gopath/src/github.com/zero-os/0-stor/bin/zstordb /bin/zstordb
    rm -rf /zstor
    mkdir /zstor
}

run_zstor_server(){
    for ((i=0; i<$1; i++)); do
        port=$((8080+$i))
        zstordb -L 0.0.0.0:$port --meta-dir /zstor/meta_$port --data-dir /zstor/data_$port &
        echo " -------------------- zstor -------------------- "
        echo "ZSTOR SERVER $i : 0.0.0.0:$port"
    done    
}

update_env
install_go
install_etcd
install_zstor_server

run_etcd
run_zstor_server $NUMBER_OF_SERVERS


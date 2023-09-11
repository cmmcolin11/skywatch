#!/bin/sh

build(){
	if ! docker images --format "{{.Repository}}:{{.Tag}}" | grep -q "msgpack-image:1.0"; then
		docker build -t msgpack-image:1.0 .
	fi
}

start(){
	if docker ps -a --format "{{.Names}}" | grep -q "msgpack"; then
		docker start -ai msgpack
	else
		docker run --name msgpack -it msgpack-image:1.0 sh
	fi
}

case "$1" in
	build)
		build
		start
		;;
	start)
		start
		;;
	*)
		echo "Usage: $0 {build|start}"
		exit
		;;
esac
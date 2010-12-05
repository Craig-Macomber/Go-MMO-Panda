#!/bin/bash

gotgo -o=tcp/connectedBag.go -package-name=tcp iterbag.got Connected


# 6g -o tcp.6 tcp/*.go
# 6g main.go
# 6l main.6
# ./6.out


DEPS="tcp"
for dep in ${DEPS}; do
	cd $dep ; gomake ; cd ..
done
# gomake clean
gomake

6l -L tcp/_obj _go_.6
./6.out
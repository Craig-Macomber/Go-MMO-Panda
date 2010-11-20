#!/bin/bash

gotgo -o=tcp/connectedBag.go -package-name=tcp iterbag.got Connected

6g -o tcp.6 tcp/*.go
6g main.go
6l main.6
./6.out
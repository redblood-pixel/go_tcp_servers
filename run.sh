#!/bin/bash

mkfifo /tmp/log_pipe

cd server1
go mod download
go build -o server1 .
sudo cp server1 /usr/local/bin/server1

cd ../server2
go mod download
go build -o server2 .
sudo cp server2 /usr/local/bin/server2

cd ../log_server
go mod download
go build -o log_server .
sudo cp log_server /usr/local/bin/log_server

cd ..
sudo cp server1.service /etc/systemd/system/server1.service
sudo cp server2.service /etc/systemd/system/server2.service
sudo cp log_server.service /etc/systemd/system/log_server.service


sudo systemctl daemon-reload
sudo systemctl start server1 server2 log_server
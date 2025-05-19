#!/bin/bash

cd server1
go build -o server1 .
sudo cp server1 /usr/local/bin/server1

cd ../server2
go build -o server2 .
sudo cp server2 /usr/local/bin/server2

cd ..
sudo cp server1.service /etc/systemd/system/server1.service
sudo cp server2.service /etc/systemd/system/server2.service


sudo systemctl daemon-reload
sudo systemctl start server1 server2
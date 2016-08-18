#!/bin/bash

#install go and git
sudo apt-get update
sudo apt-get -y upgrade
sudo apt-get -y install git
sudo $HOME
sudo curl -O https://storage.googleapis.com/golang/go1.6.linux-amd64.tar.gz
sudo tar -xvf go1.6.linux-amd64.tar.gz
rm go1.6.linux-amd64.tar.gzcd
mkdir $HOME/work
echo "export PATH=$PATH:$HOME/go/bin" >> ~/.profile
echo "export GOROOT=$HOME/go" >> ~/.profile
echo "export GOPATH=$HOME/work" >> ~/.profile
source ~/.profile

#install kanary
go get github.com/datajet-io/kanary
cd ~/work/src/github.com/datajet-io/kanary
go build


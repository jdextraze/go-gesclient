#! /bin/bash

echo ">>> Installing Go"

GO_VERSION=1.5.3
GO_ARCH=linux-amd64

# Install prerequisites
sudo apt-get -qq install wget

# Download
wget -q -O ~/go.tgz https://storage.googleapis.com/golang/go${GO_VERSION}.${GO_ARCH}.tar.gz

# Extract
tar -C $HOME -xzf ~/go.tgz

# Clean
rm ~/go.tgz

# Export paths
echo "export GOROOT=$HOME/go" >> ~/.profile
echo "export GOPATH=$HOME/code" >> ~/.profile
echo "export PATH=$HOME/code/bin:$HOME/go/bin:$PATH" >> ~/.profile
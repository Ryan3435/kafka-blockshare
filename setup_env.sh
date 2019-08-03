#!/bin/bash
set -e
if [ $EUID -ne 0 ]
	then echo "This script must be run as root" 
	exit 1
fi
apt-add-repository universe
apt-get update
apt-get install curl -y
apt-get install wget -y
apt-get install git -y
apt-get install make -y
apt-get install gcc -y


echo "Installing Golang..."
#Install Go
wget https://golang.org/dl/go1.10.4.linux-amd64.tar.gz
tar -xvzf go1.10.4.linux-amd64.tar.gz
mv go /usr/local
rm go1.10.4.linux-amd64.tar.gz

#Add Go to env variables
echo "#GO VARIABLES" >> ~/.bashrc
echo "export GOROOT=/usr/local/go" >> ~/.bashrc
echo "export GOPATH=$HOME/go" >> ~/.bashrc
echo "export PATH=$PATH:$GOROOT/bin:$GOPATH/bin" >> ~/.bashrc
echo "#END GO VARIABLES" >> ~/.bashrc

 export GOROOT=/usr/local/go
 export GOPATH=$HOME/go
 export PATH=$PATH:$GOROOT/bin:$GOPATH/bin

go get -u golang.org/x/crypto/...

echo "Installing docker..."
#Install Docker
curl  -fsSL  https://download.docker.com/linux/ubuntu/gpg  |  sudo  apt-key  add  -
add-apt-repository  "deb  [arch=amd64]  https://download.docker.com/linux/ubuntu  $(lsb_release  -cs)  stable"
apt-get update
apt-get install -y docker-ce
service docker start
systemctl enable docker

#Install Pip

echo "Installing pip..."
apt-get install python-pip -y


echo "Installing docker compose..."
#Install Docker Compose
pip install docker-compose
usermod -aG docker ${USER}
chmod +x /usr/local/bin/docker-compose

echo "Installing nodejs..."
#Install nodejs and npm
curl  -sL https://deb.nodesource.com/setup_8.x | sudo -E bash -
apt-get update
apt-get install nodejs -y
npm install npm@5.6.0 -g

#Install neccessary files for Hyperledger Fabric network
echo "Installing tools..."
git clone https://github.com/skcript/Kafka-Fabric-Network
cd ./Kafka-Fabric-Network
git checkout 0d1efa77ace5eac6bbc3195777188d9d0b1850d9
cd ..
mv ./Kafka-Fabric-Network/bin .
mv ./Kafka-Fabric-Network/.env .
rm -r Kafka-Fabric-Network
mkdir channel-artifacts
./misc/hlf_script.sh 1.2.0 1.2.0 0.4.10
mv ./fabric-samples/config .
rm -r fabric-samples

#Add the binaries needed to the path
echo "export PATH=$PATH:~/hlf/kafka-blockshare/bin" >> ~/.bashrc

#Get a copy of fabric and move it to the appropriate directory
echo "Cloning fabric..."
git clone https://github.com/hyperledger/fabric.git
cd ./fabric
git checkout 306640d399bbea46d25bfa1673f35a3fa8187b49
cd ..
mkdir -p $GOPATH/src/github.com/hyperledger
mv -f fabric $GOPATH/src/github.com/hyperledger


echo "Cloning fabric-sdk-go..."
git clone https://github.com/hyperledger/fabric-sdk-go.git
cd ./fabric-sdk-go
git checkout 9efe90fcb75414fddddac02c653b169e46d3c33c
cd ..
mv -f fabric-sdk-go $GOPATH/src/github.com/hyperledger
/usr/local/go/bin/go get -u github.com/golang/dep/cmd/dep

echo "Building fabric..."
make -C $HOME/go/src/github.com/hyperledger/fabric-sdk-go version depend-noforce license
make -C $HOME/go/src/github.com/hyperledger/fabric native docker

#Install qt5
echo "Installing Qt..."
apt-get install build-essential -y
apt-get install qtcreator -y
apt-get install build-essential -y
apt-get install libfontconfig1 -y
apt-get install mesa-common-dev -y
apt-get install libglu1-mesa-dev -y
apt-get install qt5-default -y
go get -u -v github.com/therecipe/qt/cmd/...
$(go env GOPATH)/bin/qtsetup prep
$(go env GOPATH)/bin/qtsetup check
$(go env GOPATH)/bin/qtsetup generate
$(go env GOPATH)/bin/qtsetup install


echo "export LD_LIBRARY_PATH=$HOME/go/src/github.com/therecipe/env_linux_amd64_513/5.13.0/gcc_64/lib/" >> ~/.bashrc
export LD_LIBRARY_PATH=$HOME/go/src/github.com/therecipe/env_linux_amd64_513/5.13.0/gcc_64/lib/
cd $HOME/hlf/kafka-blockshare/client_applications
echo "Building client application..."
qtdeploy build

echo "Starting blockchain network..."
cd $HOME/hlf/kafka-blockshare
./manage.sh all

echo "Project installed and blockchain network started. To run client application run $HOME/hlf/kafka-blockshare/client_applications/deploy/linux/client_applications.  You must export LD_LIBRARY_PATH first, refer to this script or the README for path or use source ~/.bashrc"

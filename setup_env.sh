#!/bin/sh
apt-add-repository universe
apt-get update
apt-get install curl -y
apt-get install wget -y
apt-get install git -y
apt-get install make -y
apt-get install gcc -y

#Install Go
wget https://golang.org/dl/go1.10.4.linux-amd64.tar.gz
tar -xvzf go1.10.4.linux-amd64.tar.gz
mv go /usr/local
rm go1.10.4.linux-amd64.tar.gz

#Add Go to env variables
echo "\n#GO VARIABLES" >> ~/.bashrc
echo "export GOROOT=/usr/local/go" >> ~/.bashrc
echo "export GOPATH=$HOME/go" >> ~/.bashrc
echo "export PATH=$PATH:$GOROOT/bin:$GOPATH/bin" >> ~/.bashrc
echo "#END GO VARIABLES\n" >> ~/.bashrc

export GOROOT=/usr/local/go
export GOPATH=$HOME/go
export PATH=$PATH:$GOROOT/bin:$GOPATH/bin

go get -u golang.org/x/crypto/

#Install Docker
curl  -fsSL  https://download.docker.com/linux/ubuntu/gpg  |  sudo  apt-key  add  -
add-apt-repository  "deb  [arch=amd64]  https://download.docker.com/linux/ubuntu  $(lsb_release  -cs)  stable"
apt-get update
apt-get install -y docker-ce
service docker start
systemctl enable docker

#Install Pip

apt-get install python-pip -y

#Install Docker Compose
pip install docker-compose
usermod -aG docker ${USER}
chmod +x /usr/local/bin/docker-compose

#Install nodejs and npm
curl  -sL https://deb.nodesource.com/setup_8.x | sudo -E bash -
apt-get update
apt-get install nodejs -y
npm install npm@5.6.0 -g

#Install neccessary files for Hyperledger Fabric network
git clone https://github.com/skcript/Kafka-Fabric-Network
mv ./Kafka-Fabric-Network/bin .
mv ./Kafka-Fabric-Network/.env .
rm -r Kafka-Fabric-Network
mkdir channel-artifacts
curl -sSL http://bit.ly/2ysbOFE| bash -s 1.2.0 1.2.0 0.4.10
mv ./fabric-samples/config .
rm -r fabric-samples

#Add the binaries needed to the path
echo "\nexport PATH=$PATH:~/hlf/kafka-blockshare/bin" >> ~/.bashrc


#Get a copy of fabric and move it to the appropriate directory
git clone https://github.com/hyperledger/fabric.git
mkdir -p $GOPATH/src/github.com/hyperledger
mv -f fabric $GOPATH/src/github.com/hyperledger

git clone https://github.com/hyperledger/fabric-sdk-go.git
mv -f fabric-sdk-go $GOPATH/src/github.com/hyperledger
/usr/local/go/bin/go get -u github.com/golang/dep/cmd/dep
/usr/local/go/bin/go get -u github.com/hyperledger/fabric-sdk-go
make -C $HOME/go/src/github.com/hyperledger/fabric-sdk-go
make -C $HOME/go/src/github.com/hyperledger/fabric peer orderer release docker

apt-get -y install build-essential libglu1-mesa-dev libpulse-dev libglib2.0-dev -y
go get -u -v github.com/therecipe/qt/cmd/... && $(go env GOPATH)/bin/qtsetup test && $(go env GOPATH)/bin/qtsetup

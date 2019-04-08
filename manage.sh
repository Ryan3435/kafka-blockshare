#!/bin/bash

if [[ $UID != 0 ]]; then
    echo "Please run this script with sudo:"
    echo "sudo $0 $*"
    exit 1
fi

if [ $# -lt 1 ] || [ $# -gt 1 ]
then
	echo "Accepts one argument. [generate|start|clean|install|client|update|all]"

elif [ $1 = all ]
then
	./$0 clean
	./$0 generate
	./$0 start
	./$0 install
	
elif [ $1 = generate ]
then
	# don't rewrite paths for Windows Git Bash users
	export MSYS_NO_PATHCONV=1
	export GOROOT=/usr/local/go
	export GOPATH=$HOME/go
	export PATH=$GOROOT/src/github.com/hyperledger/fabric/build/bin:${PWD}/bin:${PWD}:$PATH:$GOROOT/bin:$GOPATH/bin
	export FABRIC_CFG_PATH=${PWD}

	chmod 777 -R $GOPATH
	chmod 777 -R $GOROOT
	chmod 777 -R .
	# Exit on first error, print all commands.
	set -ev


	CHANNEL_NAME=mychannel
	
	mkdir -p channel-artifacts
	mkdir -p crypto-config

	./bin/cryptogen generate --config=./crypto-config.yaml

	./bin/configtxgen -profile TwoOrgsOrdererGenesis -outputBlock ./channel-artifacts/genesis.block

	export CHANNEL_NAME=mychannel  && ./bin/configtxgen -profile TwoOrgsChannel -outputCreateChannelTx ./channel-artifacts/channel.tx -channelID $CHANNEL_NAME

	./bin/configtxgen -profile TwoOrgsChannel -outputAnchorPeersUpdate ./channel-artifacts/Org1PeerOrgMSPanchors.tx -channelID $CHANNEL_NAME -asOrg Org1PeerOrg
	./bin/configtxgen -profile TwoOrgsChannel -outputAnchorPeersUpdate ./channel-artifacts/Org2PeerOrgMSPanchors.tx -channelID $CHANNEL_NAME -asOrg Org2PeerOrgMSP
	sudo chmod -R 777 ../kafka-blockshare

elif [ $1 = start ]
then
	# Exit on first error, print all commands.
	set -ev

	# don't rewrite paths for Windows Git Bash users
	export MSYS_NO_PATHCONV=1

	export GOROOT=/usr/local/go
	export GOPATH=$HOME/go
	export PATH=$GOROOT/src/github.com/hyperledger/fabric/build/bin:${PWD}/bin:${PWD}:$PATH:$GOROOT/bin:$GOPATH/bin

	echo -e "-------------------------------\n\n\nSTARTING DOCKER IMAGES\n\n\n------------------------------------------"

	
	docker-compose -f ./docker-images/docker-compose-kafka.yml up -d
	# wait for Hyperledger Fabric to start
	# incase of errors when running later commands, issue export FABRIC_START_TIMEOUT=<larger number>
	export FABRIC_START_TIMEOUT=10
	echo ${FABRIC_START_TIMEOUT}
	sleep ${FABRIC_START_TIMEOUT}

	echo -e "-------------------------------\n\n\nCREATING CHANNEL\n\n\n------------------------------------------"
	
	# # # Create the channel
	docker exec -e "CORE_PEER_LOCALMSPID=Org1PeerOrgMSP" -e "CORE_PEER_MSPCONFIGPATH=/var/hyperledger/users/Admin@peer.org1.com/msp" peer0.peer.org1.com peer channel create -o orderer0.org1.com:7050 -c mychannel -f /var/hyperledger/custom-configs/channel.tx -t 120s

	echo -e "-------------------------------\n\n\nJoining peer0.peer.org1.com\n\n\n------------------------------------------"
	
	# # # Join peer0.org1.example.com to the channel.
	docker exec -e "CORE_PEER_LOCALMSPID=Org1PeerOrgMSP" -e "CORE_PEER_MSPCONFIGPATH=/var/hyperledger/users/Admin@peer.org1.com/msp" peer0.peer.org1.com peer channel join -b mychannel.block

	echo -e "-------------------------------\n\n\nJoining peer0.peer.org2.com\n\n\n------------------------------------------"

	docker exec -e "CORE_PEER_LOCALMSPID=Org2PeerOrgMSP" -e "CORE_PEER_MSPCONFIGPATH=/var/hyperledger/users/Admin@peer.org2.com/msp" peer0.peer.org2.com peer channel fetch config /var/hyperledger/custom-configs/mychannel.block -c mychannel -o orderer0.org2.com:7050

	docker exec -e "CORE_PEER_LOCALMSPID=Org2PeerOrgMSP" -e "CORE_PEER_MSPCONFIGPATH=/var/hyperledger/users/Admin@peer.org2.com/msp" peer0.peer.org2.com peer channel join -b /var/hyperledger/custom-configs/mychannel.block

	echo -e "-------------------------------\n\n\nJoining peer1.peer.org2.com\n\n\n------------------------------------------"

	docker exec -e "CORE_PEER_LOCALMSPID=Org2PeerOrgMSP" -e "CORE_PEER_MSPCONFIGPATH=/var/hyperledger/users/Admin@peer.org2.com/msp" peer1.peer.org2.com peer channel fetch config /var/hyperledger/custom-configs/mychannel.block -c mychannel -o orderer0.org2.com:7050

	docker exec -e "CORE_PEER_LOCALMSPID=Org2PeerOrgMSP" -e "CORE_PEER_MSPCONFIGPATH=/var/hyperledger/users/Admin@peer.org2.com/msp" peer1.peer.org2.com peer channel join -b /var/hyperledger/custom-configs/mychannel.block

	echo -e "-------------------------------\n\n\nJoining peer1.peer.org1.com\n\n\n------------------------------------------"
	
	docker exec -e "CORE_PEER_LOCALMSPID=Org1PeerOrgMSP" -e "CORE_PEER_MSPCONFIGPATH=/var/hyperledger/users/Admin@peer.org1.com/msp" peer1.peer.org1.com peer channel fetch config /var/hyperledger/custom-configs/mychannel.block -c mychannel -o orderer0.org1.com:7050

	docker exec -e "CORE_PEER_LOCALMSPID=Org1PeerOrgMSP" -e "CORE_PEER_MSPCONFIGPATH=/var/hyperledger/users/Admin@peer.org1.com/msp" peer1.peer.org1.com peer channel join -b /var/hyperledger/custom-configs/mychannel.block
	
	#Changing this line to be the new latest peer cp, old line below
	docker cp peer1.peer.org1.com:/var/hyperledger/custom-configs/mychannel.block /home/ryan/hlf/kafka-blockshare/channel-artifacts/
	#docker cp peer3.peer.org2.com:/var/hyperledger/custom-configs/mychannel.block /home/ryan/hlf/kafka-blockshare/channel-artifacts/
	
	
elif [ $1 = install ]
then
	echo -e "-------------------------------\n\n\n\nInstalling chaincode on peer0.peer.org1.com\n\n\n\n------------------------------------------"
	docker exec -e "CORE_PEER_MSPCONFIGPATH=/opt/gopath/src/github.com/hyperledger/crypto-config/peerOrganizations/peer.org1.com/users/Admin@peer.org1.com/msp" -e "CORE_PEER_ADDRESS=peer0.peer.org1.com:7051" -e "CORE_PEER_LOCALMSPID=Org1PeerOrgMSP" cli peer chaincode install -n usermgmt -p chaincode -v v0 -l "golang"
	#-e "CORE_PEER_MSPCONFIGPATH=/var/hyperledger/users/Admin@peer.org1.com/msp" 
	echo -e "-------------------------------\n\n\n\nInstantiating chaincode\n\n\n\n------------------------------------------"
	docker exec -e "CORE_PEER_MSPCONFIGPATH=/opt/gopath/src/github.com/hyperledger/crypto-config/peerOrganizations/peer.org1.com/users/Admin@peer.org1.com/msp" -e "CORE_PEER_ADDRESS=peer0.peer.org1.com:7051" -e "CORE_PEER_LOCALMSPID=Org1PeerOrgMSP" cli peer chaincode instantiate -o orderer0.org1.com:7050 -C mychannel -n usermgmt -v v0 -c '{"Args": ["init"]}' 
	echo -e "-------------------------------\n\n\n\nInstalling chaincode on peer0.peer.org2.com\n\n\n\n------------------------------------------"

	docker exec -e "CORE_PEER_MSPCONFIGPATH=/opt/gopath/src/github.com/hyperledger/crypto-config/peerOrganizations/peer.org2.com/users/Admin@peer.org2.com/msp" -e "CORE_PEER_ADDRESS=peer0.peer.org2.com:7051" -e "CORE_PEER_LOCALMSPID=Org2PeerOrgMSP" cli peer chaincode install -n usermgmt -p chaincode -v v0 -l "golang"

	echo -e "-------------------------------\n\n\n\nInstalling chaincode on peer1.peer.org1.com\n\n\n\n------------------------------------------"
	
	docker exec -e "CORE_PEER_MSPCONFIGPATH=/opt/gopath/src/github.com/hyperledger/crypto-config/peerOrganizations/peer.org1.com/users/Admin@peer.org1.com/msp" -e "CORE_PEER_ADDRESS=peer1.peer.org1.com:7051" -e "CORE_PEER_LOCALMSPID=Org1PeerOrgMSP" cli peer chaincode install -n usermgmt -p chaincode -v v0 -l "golang"

	echo -e "-------------------------------\n\n\n\nInstalling chaincode on peer1.peer.org2.com\n\n\n\n------------------------------------------"

	docker exec -e "CORE_PEER_MSPCONFIGPATH=/opt/gopath/src/github.com/hyperledger/crypto-config/peerOrganizations/peer.org2.com/users/Admin@peer.org2.com/msp" -e "CORE_PEER_ADDRESS=peer1.peer.org2.com:7051" -e "CORE_PEER_LOCALMSPID=Org2PeerOrgMSP" cli peer chaincode install -n usermgmt -p chaincode -v v0 -l "golang"
	
	

elif [ $1 = stop ]
then
	export GOROOT=/usr/local/go
	docker-compose -f ./docker-images/docker-compose-kafka.yml down --remove-orphans
	docker-compose -f ./docker-images/docker-compose-cli.yml down

elif [ $1 = clean ]
then
	rm -rf ./crypto-config
	rm ./channel-artifacts/*
	docker stop $(docker ps -aq)
	docker rm $(docker ps -aq)
	docker network prune -f
	docker rmi $(docker images | grep 'usermgmt')

elif [ $1 = update ]
then
	./bin/cryptogen extend --config=./crypto-config.yaml --input="./crypto-config"
	chmod 777 -R ./crypto-config

elif [ $1 = client ]
then
	docker exec -it client /bin/bash
fi

Steps to install and run base-build version of kafka-blockshare hyperledger fabric network

0. These steps apply only to a 64-bit machine running Ubuntu 18.04, other machines have not been tested

1. Download base-build files from https://www.dropbox.com/sh/hkhtm9x99v83538/AAAl6yQDQND7DEufoLjd2_U4a?dl=0

2. Create a directory in your home folder named "hlf" and a subdirectory within hlf named "kafka-blockshare" (mkdir -p ~/hlf/kafka-blockshare)

3. Move the downloaded base build files to ~/hlf/kafka-blockshare (mv ~/Downloads/base-build.zip ~/hlf/kafka-blockshare/)

4. Unzip the base-build files (unzip base-build.zip)

5. Run the setup_env.sh script as sudo (sudo ./setup_env.sh)
	-This is the last step needed to have the environment set up, to start the network and install chaincode continue to step 6

6. Run the management script with the command "all" to start the network and install chaincode, must be run as sudo (sudo ./manage.sh all)


![alt text](https://github.com/Ryan3435/kafka-blockshare/blob/master/hlfFrameworkProtocol.png?raw=true)



# Introduction

This project is the source code for the decentralized ride-hailing platform detailed in the paper listed in the title of this github.  The project utilizes Hyperledger Fabric to run several nodes in docker and facilitate the ride-hailing messages using our custom protocol.  A client application that works with real map data using Open Street Map is built using Qt for golang.  Instructions to build and run this project on a fresh installation of Ubuntu 18.04 are listed below.



### Supported Platforms

This code has only been fully tested on Ubuntu Desktop 18.04.  An early build was run on Ubuntu Server but we cannot guarantee the automated install will work.  Installation on Windows or Mac could possibly be done by following the setup script and translating to native commands.





# Installation

1. Create a directory titled "hlf" in your home directory. 

   `mkdir $HOME/hlf && cd $HOME/hlf`

2. Use git to clone this repository to this new directory. 

   `git clone https://github.com/Ryan3435/kafka-blockshare`

3. Start the installation script. Installation of tools takes a long time, be patient. This script will also build the client application and start the blockchain network.

   `./setup_env.sh`

4. After the environment is installed it is a good idea to run `source ~/.bashrc` to update the environment variables of your shell.



# Using the client application

The same client application is used for both riders and drivers.  When you register as a user you must use a specific username that was pre-certified during the install process.  If you desire to reconfigure the network or change the certifications of users you must edit the YAML files in this project and use the manage script to re-generate the require certifications.



### Starting the client application

To run the client application you should export LD_LIBRARY_PATH to ensure Qt can find the required shared libraries.

`export LD_LIBRARY_PATH=$HOME/go/src/github.com/therecipe/env_linux_amd64_513/5.13.0/gcc_64/lib/`



After that the client application can be started by running: 

`$HOME/hlf/kafka-blockshare/client_applications/deploy/linux/client_applications`



### Driver

1. Start the client application as described above
2. Choose the register menu option
3. **Your username must be the word User followed by a number 1 through 1000.**  For example: both User1 and User1000 are valid usernames
4. Your organization must either be **peer.org1.com** or **peer.org2.com**.  This is the demo setup but more organizations can be added by reconfiguring.
5. Choose any password.
6. Navigate back to the login menu button. 
7. Re-enter your registered information and sign in as a **Rider**. All users start as riders and upgrade to being a driver through the application.
8. Navigate to the **Upgrade to Driver** menu option and enter additional information about your user. Note: This is not a production build so information will be saved but does not have to be accurate.
9. Logout
10. Click the login button but this time select **Driver**
11. Click the Start Driving option and enter a real address you would like the driver to begin at.
12. Once a ride request is transmitted the driver will receive a prompt asking if he would like to accept.



### Rider

1. Start the client application as described above
2. Choose the register menu option
3. **Your username must be the word User followed by a number 1 through 1000.**  For example: both User1 and User1000 are valid usernames
4. Your organization must either be **peer.org1.com** or **peer.org2.com**.  This is the demo setup but more organizations can be added by reconfiguring.
5. Choose any password.
6. Navigate back to the login menu button. 
7. Re-enter your registered information and sign in as a **Rider**.
8. Navigate to the Update Profile menu option and enter your information.
9. Select Request Ride and enter your pickup and dropoff locations.
10. When a driver accepts your ride will begin.



# Application Purpose

This implementation aims to show how blockchain technology could be utilized to provide a decentralized ride-hailing application.  Blockchain is useful in this context because it provides trust between competing entities. In this case competing organizations of drivers. A single user could own and operate an organization or many drivers could create an organization.  All organizations can work together however to provide the ride-hailing service such that a single user organization could provide service without having to join a centralized ride-hailing corporation and without having to develop a critical mass of users.

In all environments user privacy is critical.  In a decentralized environment individual organizations cannot be trusted to not act maliciously and cannot be responsible for user privacy as it is much more difficult to standardize good cyber-security practices.  Hyperledger Fabric allows for allowable interactions to be defined through smart contracts called "Chaincode" which helps protect against malicious behavior. Other blockchains support this feature as well but Hyperledger Fabric allows for a finer grain of control over user permissions and certifications.

### Implementation Security

This implementation uses a custom chaincode protocol to further ensure user information security. All reads/writes that are issued to the Hyperledger Fabric blockchain must be issued through a chaincode transaction.  These are invokable programs stored on the blockchain that specifiy exactly how interactions take place.  User certificates are passed implicitly to these functions and are used as part of the key during reads / writes. This ensures that no user can access information from another without having both their certificate and their password. A depiction of the transaction flow for a single ride is shown below:

![alt text](C:\Users\rmshivers42\Downloads\kafka-blockshare-master\misc\hlfFrameworkProtocol.png)





### Transaction Flow

In Hyperledger Fabric all transactions are first certified by endorsing peers the submitted to the ordering service to be aggregated into a block which is then transmitted to all peers in the network. This process is shown below.



![alt text](C:\Users\rmshivers42\Downloads\kafka-blockshare-master\misc\hlfTransactionFlow.png)


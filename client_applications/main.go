package main

import (
	"log"
	"fmt"
	"time"
	"strings"
	"errors"
	"strconv"
	"math/rand"
	"crypto/sha256"
	"encoding/base64"
	"golang.org/x/crypto/pbkdf2"
	resmgmt "github.com/hyperledger/fabric-sdk-go/pkg/client/resmgmt"
	"github.com/hyperledger/fabric-sdk-go/pkg/core/config"
	"github.com/hyperledger/fabric-sdk-go/pkg/fabsdk"
	"github.com/hyperledger/fabric-sdk-go/pkg/client/channel"
	"github.com/hyperledger/fabric-sdk-go/pkg/common/errors/retry"
	"github.com/hyperledger/fabric-sdk-go/pkg/client/event"
	contextAPI "github.com/hyperledger/fabric-sdk-go/pkg/common/providers/context"
	
	
	//Qt stuff
	"os"

	"github.com/therecipe/qt/core"
	"github.com/therecipe/qt/quick"
	"github.com/therecipe/qt/quickcontrols2"
	"github.com/therecipe/qt/widgets"
)

// FabricSetup implementation
type FabricSetup struct {
	ConfigFile      string
	ChannelID       string
	initialized     bool
	ChannelConfig   string
	OrgAdmin        string
	OrgName         string
	admin           resmgmt.Client
	sdk             *fabsdk.FabricSDK
}

type ClientConnection struct {
	orgChannelClientContext contextAPI.ChannelProvider
	chClientUser channel.Client
}

//Committed to the ledger when a ride is requested, driver uses this to get to the rider
type RideRequest struct {
  RiderID string `json:"riderID"`
  DriverID string `json:"driverID"`
  Status string `json:"status"`
  PickupLocationLat string `json:"pickupLocationLat:"`
  PickupLocationLong string `json:"pickupLocationLong:"`
  DropoffLocationX string `json:"dropoffLocationLat:"`
  DropoffLocationY string `json:"dropoffLocationLong:"`
  Distance string `json:"distance"`
  PickupTime string `json:"pickupTime"`
  DropoffTime string `json:"dropoffTime"`
}

type Rider struct {
	rideID string
	inCar bool
	pickupLocationLat float64
	pickupLocationLong float64
	dropoffLocationLat float64
	dropoffLocationLong float64
}

type Driver struct {
	currentLat float64
	currentLong float64
	numRiders int
	riders []Rider
	conn ClientConnection
}

var (
	continueUpdating bool	= false
	qmlObjects = make(map[string]*core.QObject) //qt stuff
	qmlBridge          *QmlBridge
)


// Initialize reads the configuration file and sets up the client, chain and event hub
func (setup *FabricSetup) Initialize() error {

	// Add parameters for the initialization
	if setup.initialized {
		return fmt.Errorf("sdk already initialized")
	}

	// Initialize the SDK with the configuration file
	sdk, err := fabsdk.New(config.FromFile(setup.ConfigFile))
	if err != nil {
		return fmt.Errorf("failed to create sdk: %v", err)
	}
	setup.sdk = sdk

	setup.initialized = true
	return nil
}

//
//
//START DRIVER FUNCTIONS------------------------------------
//
//

func (driver *Driver) popRider() error {
	if driver.numRiders == 0 {
		return errors.New("No riders currently")
	}
	
	driver.riders = driver.riders[1:]
	driver.numRiders = driver.numRiders - 1
	return nil
}

func (driver *Driver) peekRider() (Rider, error) {
	if driver.numRiders == 0 {
		return Rider{},errors.New("No riders currently")
	}
	return driver.riders[0], nil
}

func (driver *Driver) getRider(index int) (Rider, error) {
	if driver.numRiders < index + 1 {
		return Rider{},errors.New("No riders currently")
	}
	return driver.riders[index], nil
}

func (driver *Driver) addRider(newRider Rider) error {
	driver.riders = append(driver.riders, newRider)
	driver.numRiders += 1
	return nil
}

func (fSetup *FabricSetup) pickupRider (rideID string, userName string, orgName string) error {
	// Initialization of the Fabric SDK from the previously set properties
	err := fSetup.Initialize()
	if err != nil {
		fmt.Printf("Unable to initialize the Fabric SDK: %v\n", err)
	}
	fmt.Printf("\n\nSDK Initialized\n")

	//Try to connect the client to the channel
	orgChannelClientContext := fSetup.sdk.ChannelContext("mychannel", fabsdk.WithUser(userName), fabsdk.WithOrg(orgName))
	chClientUser, err := channel.New(orgChannelClientContext)
	_ = chClientUser
	if err != nil {
		fmt.Printf("Failed to create new channel client for Org1 user: %s", err)
		return err
	}
	
	t := time.Now()
	//Call the pickup rider method
	_, err = invoke(*chClientUser,"pickupRider",[][]byte{[]byte(rideID),[]byte(t.Format("2006-01-02 15:04:05"))},"usermgmt")
	if err != nil {
		fmt.Printf("Unable to pickup rider. Error: %s", err)
		return err
	}
	
	r, rideRequestErr := query(*chClientUser,"getRideRequest",[][]byte{[]byte(rideID)},"usermgmt")
	if rideRequestErr != nil {
		fmt.Printf("Could not get value: %s", rideRequestErr)
		return rideRequestErr
	}
	payload := string(r.Payload)
	fmt.Printf("\n\n" + payload + "\n\n")
	dropoffLocationLat := strings.Split(strings.Split(payload,"dropoffLocationLat\":\"")[1],"\"")[0]
	dropoffLocationLong := strings.Split(strings.Split(payload,"dropoffLocationLong\":\"")[1],"\"")[0]
	
	fmt.Printf("dropoff location lat: %s\n", dropoffLocationLat)
	fmt.Printf("dropoff location long: %s\n", dropoffLocationLong)
				
	floatDropoffLocationLat, _ := strconv.ParseFloat(dropoffLocationLat, 64)
	floatDropoffLocationLong, _ := strconv.ParseFloat(dropoffLocationLong, 64)
	
	qmlBridge.RiderPickup(floatDropoffLocationLat, floatDropoffLocationLong, rideID)
	return nil
}


func (fSetup *FabricSetup) dropoffRider (rideID string, userName string, orgName string) error {

	// Initialization of the Fabric SDK from the previously set properties
	err := fSetup.Initialize()
	if err != nil {
		fmt.Printf("Unable to initialize the Fabric SDK: %v\n", err)
	}
	fmt.Printf("\n\nSDK Initialized\n")

	//Try to connect the client to the channel
	orgChannelClientContext := fSetup.sdk.ChannelContext("mychannel", fabsdk.WithUser(userName), fabsdk.WithOrg(orgName))
	chClientUser, err := channel.New(orgChannelClientContext)
	_ = chClientUser
	if err != nil {
		fmt.Printf("Failed to create new channel client for Org1 user: %s", err)
		return err
	}
	
	t := time.Now()
	//Dropoff the rider
	_, err = invoke(*chClientUser,"dropoffRider",[][]byte{[]byte(rideID),[]byte(t.Format("2006-01-02 15:04:05"))},"usermgmt")
	if err != nil {
		fmt.Printf("Unable to dropoff rider. Error: %s", err)
		return err
	}
	qmlBridge.DropoffRider(rideID)
	return nil
}

func (fSetup *FabricSetup) updateCoriderPickup (my_rideID string, co_rideID string, coPickupLat float64, coPickupLong float64, userName string, orgName string) error {

	// Initialization of the Fabric SDK from the previously set properties
	err := fSetup.Initialize()
	if err != nil {
		fmt.Printf("Unable to initialize the Fabric SDK: %v\n", err)
	}
	fmt.Printf("\n\nSDK Initialized\n")

	//Try to connect the client to the channel
	orgChannelClientContext := fSetup.sdk.ChannelContext("mychannel", fabsdk.WithUser(userName), fabsdk.WithOrg(orgName))
	chClientUser, err := channel.New(orgChannelClientContext)
	_ = chClientUser
	if err != nil {
		fmt.Printf("Failed to create new channel client for Org1 user: %s", err)
		return err
	}
	
	stringPickupLat := fmt.Sprintf("%f", coPickupLat)
	stringPickupLong := fmt.Sprintf("%f", coPickupLong)
	
	_, err = invoke(*chClientUser,"setCoriderPickup",[][]byte{[]byte(my_rideID),[]byte(co_rideID),[]byte(stringPickupLat), []byte(stringPickupLong)},"usermgmt")
	if err != nil {
		fmt.Printf("Unable to dropoff rider. Error: %s", err)
		return err
	}
	return nil
}

func (fSetup *FabricSetup) updateCoriderDropoff (my_rideID string, co_rideID string, coDropoffLat float64, coDropoffLong float64, userName string, orgName string) error {

	// Initialization of the Fabric SDK from the previously set properties
	err := fSetup.Initialize()
	if err != nil {
		fmt.Printf("Unable to initialize the Fabric SDK: %v\n", err)
	}
	fmt.Printf("\n\nSDK Initialized\n")

	//Try to connect the client to the channel
	orgChannelClientContext := fSetup.sdk.ChannelContext("mychannel", fabsdk.WithUser(userName), fabsdk.WithOrg(orgName))
	chClientUser, err := channel.New(orgChannelClientContext)
	_ = chClientUser
	if err != nil {
		fmt.Printf("Failed to create new channel client for Org1 user: %s", err)
		return err
	}
	
	stringDropoffLat := fmt.Sprintf("%f", coDropoffLat)
	stringDropoffLong := fmt.Sprintf("%f", coDropoffLong)
	
	_, err = invoke(*chClientUser,"setCoriderDropoff",[][]byte{[]byte(my_rideID),[]byte(co_rideID),[]byte(stringDropoffLat), []byte(stringDropoffLong)},"usermgmt")
	if err != nil {
		fmt.Printf("Unable to dropoff rider. Error: %s", err)
		return err
	}
	return nil
}

func (driver *Driver) blockingCheckForRider() error {
	//Register to receive first ride request
	eventClient, _ := event.New(driver.conn.orgChannelClientContext, event.WithBlockEvents())
	registration, notifier, err := eventClient.RegisterChaincodeEvent("usermgmt", "rideRequest")

	if err != nil {
		fmt.Println("failed to register rideRequest event")
		return err
	}

	fmt.Println("Created chaincode event")
	fmt.Println("Waiting for a user to request a ride...")

	var receivedRequest bool = false
	//Waiting for a ride request event
	var rideRequestID string = ""
	//TODO make the event stuff a loop eventually so that it can keep looking for rides
	select {
		//Received a ride request event
		case ccEvent := <-notifier:
			fmt.Printf("Received ride request.\n")
			rideRequestID = string(ccEvent.Payload)
			receivedRequest = true
			eventClient.Unregister(registration)
		//Still waiting on a ride request event TODO handle when an event is never received by adding a bool check and returning
		case <-time.After(time.Second * 3000000):
			fmt.Println("timeout while waiting for chaincode event")
	}
	if receivedRequest != true {
		fmt.Printf("Did not receive a ride request in the time allotment.")
		return errors.New("Did not receive a ride request in the time allotment.")
	}
	driver.acceptAndAddRider(rideRequestID)
	
	return nil
}

func (driver *Driver) acceptAndAddRider(rideID string) error {

	//Get the rider's location
	r, rideRequestErr := query(driver.conn.chClientUser,"getRideRequest",[][]byte{[]byte(rideID)},"usermgmt")
	if rideRequestErr != nil {
		fmt.Printf("Could not get value: %s", rideRequestErr)
		return rideRequestErr
	}
	payload := string(r.Payload)
	pickupLocationLat := strings.Split(strings.Split(payload,"pickupLocationLat\":\"")[1],"\"")[0]
	pickupLocationLong := strings.Split(strings.Split(payload,"pickupLocationLong\":\"")[1],"\"")[0]
	
	
	//get the integers separately
	floatPickupLocationLat, _ := strconv.ParseFloat(pickupLocationLat, 64)
	floatPickupLocationLong, _ := strconv.ParseFloat(pickupLocationLong, 64)
	
	var rider Rider = Rider {
		rideID:		rideID,
		inCar:		false,
		pickupLocationLat:	floatPickupLocationLat,
		pickupLocationLong:	floatPickupLocationLong,
		dropoffLocationLat:	-1,
		dropoffLocationLong:	-1,
	}
	driver.addRider(rider)
	return nil
}

func (driver *Driver) nonBlockingCheckForRider() error {
	//Register to receive first ride request
	eventClient, _ := event.New(driver.conn.orgChannelClientContext, event.WithBlockEvents())
	registration, notifier, err := eventClient.RegisterChaincodeEvent("usermgmt", "rideRequest")

	if err != nil {
		fmt.Println("failed to register rideRequest event")
		return err
	}

	var receivedRequest bool = false
	var rideRequestID string = ""
	
	//Waiting for a ride request event
	select {
		//Received a ride request event
		case ccEvent := <-notifier:
			fmt.Printf("\n\nReceived new ride request.\nAttempting to accept ride...\n")
			receivedRequest = true
			rideRequestID = string(ccEvent.Payload)
		//Still waiting on a ride request event TODO handle when an event is never received by adding a bool check and returning
		case <-time.After(time.Second * 2):
	}
	eventClient.Unregister(registration)
	if receivedRequest != true {
		return errors.New("Did not receive a ride request in the time allotment.")
	}
	err = driver.acceptAndAddRider(rideRequestID)
	if err != nil {
		return err
	}
	
	return nil
}

func (driver *Driver) getDropoffLocation (riderIndex int) error {
	//Get the dropoffLocation
	r, rideRequestErr := query(driver.conn.chClientUser,"getRideRequest",[][]byte{[]byte(driver.riders[riderIndex].rideID)},"usermgmt")
	if rideRequestErr != nil {
		fmt.Printf("Could not get value: %s", rideRequestErr)
		return rideRequestErr
	}
	payload := string(r.Payload)
	dropoffLocationLat := strings.Split(strings.Split(payload,"dropoffLocationLat\":\"")[1],"\"")[0]
	dropoffLocationLong := strings.Split(strings.Split(payload,"dropoffLocationLong\":\"")[1],"\"")[0]
				
	floatDropoffLocationLat, _ := strconv.ParseFloat(dropoffLocationLat, 64)
	floatDropoffLocationLong, _ := strconv.ParseFloat(dropoffLocationLong, 64)
	driver.riders[riderIndex].dropoffLocationLat = floatDropoffLocationLat
	driver.riders[riderIndex].dropoffLocationLong = floatDropoffLocationLong
	return nil
}


func (fsetup *FabricSetup) startDriving(chClientUser *channel.Client, orgChannelClientContext contextAPI.ChannelProvider,ccargs [][]byte) error {
	
	fmt.Printf("Adding driver to the pool\n")
	_, err := invoke(*chClientUser,"addDriverToPool",ccargs,"usermgmt")
	if err != nil {
		fmt.Printf("Could not add driver to the pool. Error: %s", err)
		return err
	}
	fmt.Printf("Added driver to the pool.\n")
	
	//Pull out locations from ccargs
	locationSlice := strings.Split(string(ccargs[0]), ",")
	driverLocationLat := locationSlice[0]
	driverLocationLong := locationSlice[1]
	fmt.Printf("Driver starting location: (%s,%s)\n",driverLocationLat, driverLocationLong)
	
	floatDriverLocationLat, _ := strconv.ParseFloat(driverLocationLat, 64)
	floatDriverLocationLong, _ := strconv.ParseFloat(driverLocationLong, 64)
	qmlBridge.UpdateDriver(floatDriverLocationLat, floatDriverLocationLong) //qt
	
	var driver Driver = Driver{
		currentLat:		floatDriverLocationLat,
		currentLong:	floatDriverLocationLong,
		numRiders:		0,
		riders:			[]Rider{},
		conn:			ClientConnection{orgChannelClientContext, *chClientUser},
	}
	
	
	for {
		err = driver.blockingCheckForRider()
		if err != nil {
			fmt.Printf("Error in rideRequest process: %s", err)
			return err
		}
		
		rider, _ := driver.peekRider()
		fmt.Printf("\n\nSending new ride request to QML\n\n")
		fmt.Printf("\n\nLatitude: %s\n\n", rider.pickupLocationLat)
		fmt.Printf("\n\nLongitude: %s\n\n", rider.pickupLocationLong)
		qmlBridge.NewRideRequest(rider.pickupLocationLat, rider.pickupLocationLong, rider.rideID) //qt
		driver.popRider()
	}
	return nil
}



//
//
//END DRIVER FUNCTIONS------------------------------------
//
//
//}



/*
This function:
1. Invokes the request ride chaincode function (which sends an event to listening drivers and creates the temporary ride
request key.
2. Waits for a driver to accept the ride request and then calls the "setRideDest" chaincode function which updates the 
temporary key in the kedger to include the destination location for the ride
3. Waits to get picked up, when picked up it then waits to be dropped off
4. When dropped off it creates a permanent key in the ledger for the ride and updates all of the information specifically
for this ride from this riders perspective
*/
func (fsetup *FabricSetup) requestRide(chClientUser *channel.Client, orgChannelClientContext contextAPI.ChannelProvider,ccargs [][]byte) error {
	file, err := os.Create("riding.txt")
	if err != nil {
		log.Fatal("Cannot create file", err)
	}
	defer file.Close()
	
	// Call the requestRide chaincode function using the invoke helper function
	r, err := invoke(*chClientUser,"requestRide",[][]byte{ccargs[0]},"usermgmt")
	_ = r
	if err != nil {
		fmt.Printf("Unable to request ride. Error: %s", err)
		return err
	}
	fmt.Printf("Response payload: %s\n\n\n", r.Payload)
	fmt.Printf("Successfully requested ride")
	
	// Create an event to wait for a "rideAccept" event to indicate a driver has accepted the ride
	eventClient, _ := event.New(orgChannelClientContext, event.WithBlockEvents())
	registration, notifier, err := eventClient.RegisterChaincodeEvent("usermgmt", "rideAccept")
	if err != nil {
		fmt.Printf("failed to register rideAccept event")
		return err
	}
	
	fmt.Printf("Created chaincode event")
	fmt.Printf("Waiting for a driver to accept the ride request...")
	
	//Get the rider's ride request id
	fmt.Printf("before get user info")
	r, _ = query(*chClientUser,"getUserInfo",[][]byte{},"usermgmt")
	payload := string(r.Payload)
	fmt.Printf("\n\n" + payload + "\n\n")
	rideRequestID := "RideRequest-" + strings.Split(strings.Split(payload,"userID\":\"")[1],"\"")[0]
	
	//Set to true if a ride is accepted by a driver
	var receivedAccept bool = false
	
	//Waiting for a ride accept event
	select {
	
		//Received a rideAccept event
		case ccEvent := <-notifier:
			_ = ccEvent
			fmt.Printf("Received ride accept.\nAttempting to set dest..\n")
			receivedAccept = true
			eventClient.Unregister(registration)
			
		//Still waiting on a rideAccept event 
		case <-time.After(time.Second * 3000000):
			fmt.Printf("timeout while waiting for a driver to send a rideAccept event")
	}
	// Return with error if a rideAccept event is never received
	if receivedAccept != true {
		fmt.Printf("Did not receive a ride accept in the time allotment after requesting ride.")
		return errors.New("Did not receive a ride accept in the time allotment after requesting ride.")
	}

	//When accepted, set destination in ledger
	fmt.Printf("Setting the destination in the ledger\n")
	r, err = invoke(*chClientUser,"setRideDest",[][]byte{ccargs[1]},"usermgmt")
	_ = r
	if err != nil {
		fmt.Printf("Unable to set ride destination after ride was accepted by driver. Error: %s\n", err)
		return err
	}
	
	fmt.Printf("Waiting for driver to arrive...\n")
	
	// Set an event to start waiting for the driver to arrive
	eventClient, _ = event.New(orgChannelClientContext, event.WithBlockEvents())
	registration, notifier, err = eventClient.RegisterChaincodeEvent("usermgmt", "ridePickup")
	if err != nil {
		fmt.Printf("failed to register ridePickup event")
	}

	var receivedPickup bool = false
	var stopListening bool = false
	
	for stopListening == false {
		// Waiting for a ridePickup event to show the driver arrived
		select {
			// Received a ridePickup event
			case ccEvent := <-notifier:
				if string(ccEvent.Payload) == rideRequestID {
					fmt.Printf("Driver has arrived.\nJoined driver.\nWaiting for driver to reach my dropoff location\n")
					receivedPickup = true
					eventClient.Unregister(registration)
					stopListening = true
				}
			//Still waiting on a ride request event TODO handle when an event is never received by adding a bool check and returning
			case <-time.After(time.Second * 3000000):
				fmt.Printf("timeout while waiting for pickup")
				stopListening = true
		}
	}
	if receivedPickup != true {
		fmt.Printf("Did not receive a ride pickup in the time allotment.")
		return errors.New("Did not receive a ride pickup in the time allotment.")
	}
	
	// Set an event to start waiting for the ride to end
	eventClient, _ = event.New(orgChannelClientContext, event.WithBlockEvents())
	registration, notifier, err = eventClient.RegisterChaincodeEvent("usermgmt", "rideDropoff")
	if err != nil {
		fmt.Printf("failed to register rideDropoff event")
	}


	var receivedDropoff bool = false
	stopListening = false
	
	for stopListening == false {
	// Waiting for a rideDropoff event
		select {
			//Received a ride requestdropoff notifier:
			case ccEvent := <-notifier:
				if string(ccEvent.Payload) == rideRequestID {
					fmt.Printf("Arrived at destination.\nLeaving driver...\n")
					receivedDropoff = true
					eventClient.Unregister(registration)
					stopListening = true
				}
			//Still waiting on a ride request event TODO handle when an event is never received by adding a bool check and returning
			case <-time.After(time.Second * 3000000):
				fmt.Printf("timeout while waiting for dropoff")
				stopListening = true
		}
	}
	if receivedDropoff != true {
		fmt.Printf("Did not receive a ride dropoff in the time allotment.")
		return errors.New("Did not receive a ride dropoff in the time allotment.")
	}
	time.Sleep(time.Second * 1)
	
	//Call the leaveDriver chaincode
	_, err = invoke(*chClientUser,"leaveDriver",[][]byte{},"usermgmt")
	if err != nil {
		fmt.Printf("Unable to leave driver. Error: %s", err)
		return err
	}
	
	return nil
}

func (fSetup *FabricSetup) setupAndInvoke (userName string, orgName string, command string, ccargs [][]byte) error {

	// Initialization of the Fabric SDK from the previously set properties
	err := fSetup.Initialize()
	if err != nil {
		fmt.Printf("Unable to initialize the Fabric SDK: %v\n", err)
	}
	fmt.Printf("\n\nSDK Initialized\n")

	//Try to connect the client to the channel
	orgChannelClientContext := fSetup.sdk.ChannelContext("mychannel", fabsdk.WithUser(userName), fabsdk.WithOrg(orgName))
	chClientUser, err := channel.New(orgChannelClientContext)
	_ = chClientUser
	if err != nil {
		fmt.Printf("Failed to create new channel client for Org1 user: %s", err)
		return err
	}

	fmt.Printf("\nClient connected to mychannel as %s of %s\n\n", userName, orgName)

	switch command {
		case "startDriving":
			fSetup.startDriving(chClientUser, orgChannelClientContext, ccargs)
		case "getDriverPool":
			r, err := query(*chClientUser,"getDriverPool",ccargs,"usermgmt")
			if err != nil {
				fmt.Printf("Could not get value: %s", err)
				return err
			}
			fmt.Printf("\n\n%s\n", r.Payload)
		case "requestRide":
			fSetup.requestRide(chClientUser, orgChannelClientContext, ccargs)
		case "registerUser":
			r, err := invoke(*chClientUser,command,ccargs,"usermgmt")
			if err != nil {
				fmt.Printf("Unable to register user. Error: %s", err)
				return err
			}
			fmt.Printf("Response payload: %s\n", r.Payload)
			return nil
		case "updateUserName":
			r, err := invoke(*chClientUser,command,ccargs,"usermgmt")
			if err != nil {
				fmt.Printf("Unable to update user profile. Error: %s", err)
				return err
			}
			fmt.Printf("Response payload: %s\n", r.Payload)
			return nil
		case "upgradeToDriver":
			r, err := invoke(*chClientUser,command,ccargs,"usermgmt")
			if err != nil {
				fmt.Printf("Unable to upgrade user to driver. Error: %s", err)
				return err
			}
			fmt.Printf("Response payload: %s\n", r.Payload)
			return nil
		case "getDriverInfo":
			r, err := query(*chClientUser,command,ccargs,"usermgmt")
			_ = r
			if err != nil {
				fmt.Printf("Unable to get driver info. Error: %s", err)
				return err
			}
			fmt.Printf("\n\n%s\n", r.Payload)
			return nil
		case "getUserInfo":
			fmt.Printf("Trying to issue command: %s", command)
			r, err := query(*chClientUser,command,ccargs,"usermgmt")
			_ = r
			if err != nil {
				fmt.Printf("Unable to get rider info. Error: %s", err)
				return err
			}
			fmt.Printf("\n\n%s\n", r.Payload)
			return nil
		case "loginRider":
			fmt.Printf("Trying to issue command: %s", command)
			r, err := query(*chClientUser,"getUserInfo",[][]byte{},"usermgmt")
			if err != nil {
				return err
			}
			payload := string(r.Payload)
			hash := strings.Split(strings.Split(payload,"hash\":\"")[1],"\"")[0]
			salt := strings.Split(strings.Split(payload,"salt\":\"")[1],"\"")[0]
			pw := string(ccargs[0])
			
			dk := pbkdf2.Key([]byte(pw), []byte(salt), 20000, 32, sha256.New)
			new_hash := base64.URLEncoding.EncodeToString(dk)
			if new_hash != hash {
				fmt.Printf("\nCould not log in. Incorrect password.\n")
				return errors.New("Incorrect password")
			} else {
				fmt.Printf("\nSuccessfully logged in\n")
				return nil
			}
		case "loginDriver":
			fmt.Printf("Trying to issue command: %s\n", command)
			r, err := query(*chClientUser,"getUserInfo",[][]byte{},"usermgmt")
			if err != nil {
				return err
			}
			hash := strings.Split(strings.Split(string(r.Payload),"hash\":\"")[1],"\"")[0]
			salt := strings.Split(strings.Split(string(r.Payload),"salt\":\"")[1],"\"")[0]
			pw := string(ccargs[0])
			
			dk := pbkdf2.Key([]byte(pw), []byte(salt), 20000, 32, sha256.New)
			new_hash := base64.URLEncoding.EncodeToString(dk)
			
			if new_hash != hash {
				fmt.Printf("\nCould not log in. Incorrect password.\n")
				return errors.New("Incorrect password")
			} else {
				fmt.Printf("\nSuccessfully logged in\n")
				return nil
			}
		case "acceptRide":
			//Accept the ride
			_, err := invoke(*chClientUser,"acceptRide",ccargs,"usermgmt")
			if err != nil {
				fmt.Printf("Unable to accept ride. Error: %s", err)
				return err
			}
			fmt.Printf("Accepted ride.\n\n")
		default:
			fmt.Printf("Function name not recognized")
	}
	return nil
}

func invoke (chClientUser channel.Client, command string, ccargs [][]byte, chaincodeID string) (channel.Response, error) {
			r, err := chClientUser.Execute(channel.Request{ChaincodeID: chaincodeID, Fcn: command, Args: ccargs},channel.WithTargetEndpoints("peer0.peer.org1.com", "peer1.peer.org1.com","peer0.peer.org2.com", "peer1.peer.org2.com"),channel.WithRetry(retry.DefaultChannelOpts))
			return r, err
}

func query (chClientUser channel.Client, command string, ccargs [][]byte, chaincodeID string) (channel.Response, error) {
			r, err := chClientUser.Query(channel.Request{ChaincodeID: chaincodeID, Fcn: command, Args: ccargs},channel.WithTargetEndpoints("peer0.peer.org1.com", "peer1.peer.org1.com","peer0.peer.org2.com", "peer1.peer.org2.com"),channel.WithRetry(retry.DefaultChannelOpts))
			return r, err
}

func RandomString() string {
    bytes := make([]byte, 32)
    for i := 0; i < 32; i++ {
        bytes[i] = byte(65 + rand.Intn(25))  //A=65 and Z = 65+25
    }
    return string(bytes)
}

//Qt stuff
type QmlBridge struct {
	core.QObject
	_ func() `signal:"driverLogin"`
	_ func() `signal:"riderLogin"`
	_ func(rideID string) `signal:"dropoffRider"`
	_ func(latitude float64, longitude float64) `signal:"updateDriver"`
	_ func(latitude float64, longitude float64, rideID string) `signal:"newRideRequest"`
	_ func(latitude float64, longitude float64, rideID string) `signal:"riderPickup"`
	
	_ func(latitude float64, longitude float64, rideID string) `signal:"driverAcceptedRideRequestSignal"`
	_ func(latitude float64, longitude float64, rideID string) `slot:"driverAcceptedRideRequestSlot"`
	_ func(channel chan bool) `slot:"stopDriving"`
	
	_ func(userID string, orgID string, funcName string, arg1 string, arg2 string, arg3 string, arg4 string, arg5 string, arg6 string, arg7 string) `slot:"goFunction"`
	_ func(latitude float64, longitude float64) `slot:"updateDriverVisual"`
	_ func(rideID string, userName string, orgName string) `slot:"pickupRiderCC"`
	_ func(rideID string, userName string, orgName string) `slot:"dropoffRiderCC"`
	_ func(my_rideID string, co_rideID string, coPickupLat float64, coPickupLong float64, userName string, orgName string) `slot:"updateCoriderPickupCC"`
	_ func(my_rideID string, co_rideID string, coDropoffLat float64, coDropoffLong float64, userName string, orgName string) `slot:"updateCoriderDropoffCC"`
	
}

func main() {

//-----------------QT Stuff--------------------

	// enable high dpi scaling
	// useful for devices with high pixel density displays
	// such as smartphones, retina displays, ...
	core.QCoreApplication_SetAttribute(core.Qt__AA_EnableHighDpiScaling, true)
	
	// needs to be called once before you can start using QML/Quick
	widgets.NewQApplication(len(os.Args), os.Args)

	// use the material style
	// the other inbuild styles are:
	// Default, Fusion, Imagine, Universal
	quickcontrols2.QQuickStyle_SetStyle("Material")

	// create the quick view
	// with a minimum size of 250*200
	// set the window title to "Hello QML/Quick Example"
	// and let the root item of the view resize itself to the size of the view automatically
	view := quick.NewQQuickView(nil)
	view.SetMinimumSize(core.NewQSize2(250, 200))
	view.SetResizeMode(quick.QQuickView__SizeRootObjectToView)
	view.SetTitle("BlockShare")
	
	var quickWidget = quick.NewQQuickWidget(nil)
	quickWidget.SetResizeMode(quick.QQuickWidget__SizeRootObjectToView)

	
	
	
	//New Qt stuff
	qmlBridge = NewQmlBridge(nil)
	view.RootContext().SetContextProperty("qmlBridge", qmlBridge)		
	
	qmlBridge.ConnectDriverAcceptedRideRequestSlot(func(latitude float64, longitude float64, rideID string) {
		qmlBridge.DriverAcceptedRideRequestSignal(latitude, longitude, rideID)
	})
	
	qmlBridge.ConnectUpdateDriverVisual(func(latitude float64, longitude float64) {
		time.Sleep(500 * time.Millisecond)
		qmlBridge.UpdateDriver(latitude, longitude)
	})
	
	qmlBridge.ConnectPickupRiderCC(func(rideID string, userName string, orgName string) {
		//quitDriving := make(chan bool)
		// Definition of the Fabric SDK properties
		var fSetup FabricSetup = FabricSetup{
			OrgAdmin:        "Admin",
			OrgName:         orgName,
			ConfigFile:      "/home/ryan/hlf/kafka-blockshare/client_applications/dock_config.yaml", //qT
			// Channel parameters 
			ChannelID:       "mychannel",
			ChannelConfig:   "/home/ryan/hlf/kafka-blockshare/channel-artifacts/mychannel.block", //qT
		}
		go fSetup.pickupRider(rideID, userName, orgName)
	})
	
	qmlBridge.ConnectDropoffRiderCC(func(rideID string, userName string, orgName string) {
		// Definition of the Fabric SDK properties
		//quitDriving := make(chan bool)
		var fSetup FabricSetup = FabricSetup{
			OrgAdmin:        "Admin",
			OrgName:         orgName,
			ConfigFile:      "/home/ryan/hlf/kafka-blockshare/client_applications/dock_config.yaml", //qT
			// Channel parameters 
			ChannelID:       "mychannel",
			ChannelConfig:   "/home/ryan/hlf/kafka-blockshare/channel-artifacts/mychannel.block", //qT
		}
		go fSetup.dropoffRider(rideID, userName, orgName)
	})
	
	qmlBridge.ConnectUpdateCoriderPickupCC(func(my_rideID string, co_rideID string, coPickupLat float64, coPickupLong float64, userName string, orgName string) {
		//quitDriving := make(chan bool)
		// Definition of the Fabric SDK properties
		var fSetup FabricSetup = FabricSetup{
			OrgAdmin:        "Admin",
			OrgName:         orgName,
			ConfigFile:      "/home/ryan/hlf/kafka-blockshare/client_applications/dock_config.yaml", //qT
			// Channel parameters 
			ChannelID:       "mychannel",
			ChannelConfig:   "/home/ryan/hlf/kafka-blockshare/channel-artifacts/mychannel.block", //qT
		}
		go fSetup.updateCoriderPickup(my_rideID, co_rideID, coPickupLat, coPickupLong, userName, orgName)
	})
	
	qmlBridge.ConnectUpdateCoriderDropoffCC(func(my_rideID string, co_rideID string, coDropoffLat float64, coDropoffLong float64, userName string, orgName string) {
		//quitDriving := make(chan bool)
		// Definition of the Fabric SDK properties
		var fSetup FabricSetup = FabricSetup{
			OrgAdmin:        "Admin",
			OrgName:         orgName,
			ConfigFile:      "/home/ryan/hlf/kafka-blockshare/client_applications/dock_config.yaml", //qT
			// Channel parameters 
			ChannelID:       "mychannel",
			ChannelConfig:   "/home/ryan/hlf/kafka-blockshare/channel-artifacts/mychannel.block", //qT
		}
		go fSetup.updateCoriderDropoff(my_rideID, co_rideID, coDropoffLat, coDropoffLong, userName, orgName)
	})
	
	qmlBridge.ConnectGoFunction(func(userID string, orgID string, funcName string, arg1 string, arg2 string, arg3 string, arg4 string, arg5 string, arg6 string, arg7 string) {
		
		// Definition of the Fabric SDK properties
		var fSetup FabricSetup = FabricSetup{
			OrgAdmin:        "Admin",
			OrgName:         orgID,
			ConfigFile:      "/home/ryan/hlf/kafka-blockshare/client_applications/dock_config.yaml", //qT
			// Channel parameters 
			ChannelID:       "mychannel",
			ChannelConfig:   "/home/ryan/hlf/kafka-blockshare/channel-artifacts/mychannel.block", //qT
		}
		
		if funcName == "registerUser" {
				//If the passwords match
				if arg1 == arg2{
					salt := RandomString()
					dk := pbkdf2.Key([]byte(arg1), []byte(salt), 20000, 32, sha256.New)
					hash := base64.URLEncoding.EncodeToString(dk)
					err := fSetup.setupAndInvoke(userID,orgID,funcName,[][]byte{[]byte(hash),[]byte(salt)})
					if err != nil {
						fmt.Printf( "Failed to register driver")
					} else {
						fmt.Printf( "Successfully registered driver")
					}
				} else {
					fmt.Printf( "Passwords do not match")
				}
		}
		
		if funcName == "updateUserName" {
			err := fSetup.setupAndInvoke(userID,orgID,funcName,[][]byte{[]byte(arg1),[]byte(arg2)})
			if err != nil {
				fmt.Printf( "Failed to update user profile")
			} else {
				fmt.Printf( "Successfully updated user profile")
			}
		}
		if funcName == "upgradeToDriver" {
			err := fSetup.setupAndInvoke(userID,orgID,funcName,[][]byte{[]byte(arg1),[]byte(arg2),[]byte(arg3),[]byte(arg4),[]byte(arg5)})
			if err != nil {
				fmt.Printf( "Failed to upgrade to driver")
			} else {
				fmt.Printf( "Successfully upgraded user to driver")
			}
		}

		if funcName == "loginRider" {
			err := fSetup.setupAndInvoke(userID,orgID,funcName,[][]byte{[]byte(arg1)})
			if err != nil {
				fmt.Printf( "Failed to login as rider")
			} else {
				fmt.Printf( "Successfully logged in as rider")
				qmlBridge.RiderLogin()
			}
		}
		if funcName == "loginDriver" {
			err := fSetup.setupAndInvoke(userID,orgID,funcName,[][]byte{[]byte(arg1)})
			if err != nil {
				fmt.Printf( "Failed to login as driver")
			} else {
				fmt.Printf( "Successfully logged in as driver")
				qmlBridge.DriverLogin()
			}
		}
		if funcName == "startDriving" {
			//TODO - Check user has upgraded to driver
			//qmlBridge.DrivingStarted()
			go fSetup.setupAndInvoke(userID,orgID,funcName,[][]byte{[]byte(arg1 + "," + arg2)})
			/*
			if err != nil {
				fmt.Printf( "Failed to start driving")
			} else {
				fmt.Printf( "Successfully started driving")
			}
			*/
		}
		if funcName == "requestRide" {
			err := fSetup.setupAndInvoke(userID,orgID,funcName,[][]byte{[]byte(arg1 + "," + arg2),[]byte(arg3 + "," + arg4)})
			if err != nil {
				fmt.Printf( "Failed to request ride")
			} else {
				fmt.Printf( "Successfully requested ride")
			}
		}
		if funcName == "acceptRide" {
			err := fSetup.setupAndInvoke(userID,orgID,funcName,[][]byte{[]byte(arg1)})
			if err != nil {
				fmt.Printf( "Failed to accept ride")
			} else {
				fmt.Printf( "Successfully accepted ride")
			}
		}
	})

	
	// load the embeeded qml file
	// created by either qtrcc or qtdeploy
	view.SetSource(core.NewQUrl3("qrc:/qml/main.qml", 0))
	// you can also load a local file like this instead:
	//view.SetSource(core.QUrl_FromLocalFile("./qml/main.qml"))

	// make the view visible
	view.Show()

	// start the main Qt event loop
	// and block until app.Exit() is called
	// or the window is closed by the user
	widgets.QApplication_Exec()

	
//---------------------End qt stuff----------------------------------	

}

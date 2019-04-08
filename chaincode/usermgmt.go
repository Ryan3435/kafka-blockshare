package main

import (
	"encoding/json"
	"fmt"
	"errors"
	"strings"
	"github.com/hyperledger/fabric/core/chaincode/shim"
	"github.com/hyperledger/fabric/core/chaincode/lib/cid"
	sc "github.com/hyperledger/fabric/protos/peer"
)

type SmartContract struct {
}

//Struct for storing both rides given and taken that are attached to the User object in the ledger
type Ride struct {
  RideID string `json:"rideID"`
  RiderID string `json:"riderID"`
  DriverID string `json:"driverID"`
  PickupTime string `json:"pickupTime"`
  DropoffTime string `json:"dropoffTime"`
  StartLocation string `json:"startLocation"`
  DropoffLocation string `json:"dropoffLocation"`
  Distance string `json:"distance"`
  Price string `json:"price"`
  CoriderID string `json:"coriderID"`
  CoriderPickupLocation string `json:"coriderPickupLocation"`
  CordiderDropoffLocation string `json:"coriderDropoffLocation"`
}

//TODO - Probably delete because most likely unneeded
type PoolEntry struct {
  DriverID string `json:"driverID"`
  DriverLocation string `json:"location"`
}
type Pool []PoolEntry

//User object in the ledger with all of the user's information + their previous rides
type User struct {
  UserID string `json:"userID"`
  FirstName string `json:"firstName"`
  LastName string `json:"lastName"`
  IsDriver bool `json:"isDriver"`
  Rides []Ride
  VehicleMake string `json:"vehicleMake"`
  VehicleModel string `json:"vehicleModel"`
  VehicleYear string `json:"vehicleYear"`
  Hash string `json:"hash"`
  Salt string `json:"salt"`
}

//Committed to the ledger when a ride is requested, controls driver / rider communications and storage.
//Deleted upon ride finish
type RideRequest struct {
  RiderID string `json:"riderID"`
  DriverID string `json:"driverID"`
  Status string `json:"status"`
  PickupLocationLat string `json:"pickupLocationLat"`
  PickupLocationLong string `json:"pickupLocationLong"`
  DropoffLocationLat string `json:"dropoffLocationLat"`
  DropoffLocationLong string `json:"dropoffLocationLong"`
  Distance string `json:"distance"`
  PickupTime string `json:"pickupTime"`
  DropoffTime string `json:"dropoffTime"`
  CoriderID string `json:"coriderID"`
  CoriderPickupLocationLat string `json:"coriderPickupLocationLat"`
  CoriderPickupLocationLong string `json:"coriderPickupLocationLong"`
  CoriderDropoffLocationLat string `json:"coriderDropoffLocationLat"`
  CoriderDropoffLocationLong string `json:"coriderDropoffLocationLong"`
}

// Init function
func (s *SmartContract) Init(APIstub shim.ChaincodeStubInterface) sc.Response {
  return shim.Success(nil)
}

//Function called when chaincode is invoked to determine functionality being called
func (s *SmartContract) Invoke(APIstub shim.ChaincodeStubInterface) sc.Response {
    fn, args := APIstub.GetFunctionAndParameters()

    var result string
    var err error
	
	if fn == "addDriverToPool" {
        result, err = s.addDriverToPool(APIstub, args)
    } else if fn == "getDriverPool" {
        result, err = s.getDriverPool(APIstub)
    } else if fn == "removeDriverFromPool" {
		result, err = s.removeDriverFromPool(APIstub)
    } else if fn == "requestRide" {
		result, err = s.requestRide(APIstub, args)
    } else if fn == "registerUser" {
		result, err = s.registerUser(APIstub, args)
	} else if fn == "unregisterUser" {
		result, err = s.unregisterUser(APIstub, args)
	} else if fn == "upgradeToDriver" {
		result, err = s.upgradeToDriver(APIstub, args)
	} else if fn == "updateUserName" {
		result, err = s.updateUserName(APIstub, args)
    } else if fn == "getUserInfo" {
		result, err = s.getUserInfo(APIstub)
	} else if fn == "acceptRide" {
		result, err = s.acceptRide(APIstub,args)
	} else if fn == "getRideRequest" {
		result, err = s.getRideRequest(APIstub,args)
	} else if fn == "setRideDest" {
		result, err = s.setRideDest(APIstub,args)
	} else if fn == "pickupRider" {
		result, err = s.pickupRider(APIstub,args)
	} else if fn == "dropoffRider" {
		result, err = s.dropoffRider(APIstub,args)
	} else if fn == "leaveDriver" {
		result, err = s.leaveDriver(APIstub,args)
	} else if fn == "setCoriderPickup" {
		result, err = s.setCoriderPickup(APIstub,args)
	} else if fn == "setCoriderDropoff" {
		result, err = s.setCoriderDropoff(APIstub,args)
    } else {
		return shim.Error(err.Error())
    }
    if err != nil {
		fmt.Println(result)
        return shim.Error(err.Error())
    }
    // If function executed successfully return the result as success payload
    return shim.Success([]byte(result))
}

//Create an entry in the ledger for the driver's information
func (s *SmartContract) registerUser(APIstub shim.ChaincodeStubInterface, args []string) (string, error) {
  
  //Validate the amount of arguments is correct
  if len(args) != 2 {
    return "Incorrect number of arguments provided!", errors.New("wrong arg amount")
  }

  //Get the specific userID by combining the MSP userID and the mspID
  userID, err := cid.GetID(APIstub)
  if err != nil {
    return "Error with cid.GetID", err
  }
  mspID, err := cid.GetMSPID(APIstub)
  if err != nil {
    return "Error with cid.GetMSPID", err
  }
  //Fill local variables with the User's information
  Id := mspID + "-" + userID
  
  //Get password hash and salt from provided arguments
  hash := args[0]
  salt := args[1]

  //Make sure that the user is not already registered
  r, err := APIstub.GetState(Id)
  if len(r) != 0 {
    return "Error: user already registered", errors.New("User already registered")
  }
  
  //Create a blank array for the user's rides
  var rides []Ride
  
  //Create the user registration and commit it to the ledger
  var userRegistration = User{Id, "", "", false, rides, "", "", "", hash, salt}
  Bytes, err := json.Marshal(userRegistration)
  if err != nil {
	return "Could not marshal user data provided.", errors.New("Could not marshal user data provided.")
  }
  err = APIstub.PutState(Id, Bytes)
  if err != nil {
    return "Could not register user. ", errors.New("Could not register user. ")
  }
  return "User successfully registered", nil
}

//Create an entry in the ledger for the driver's information
func (s *SmartContract) upgradeToDriver(APIstub shim.ChaincodeStubInterface, args []string) (string, error) {
  //Validate the amount of arguments is correct
  if len(args) != 5 {
    return "Incorrect number of arguments provided!", errors.New("wrong arg amount")
  }

  //Get the specific userID by combining the MSP userID and the mspID
  userID, err := cid.GetID(APIstub)
  if err != nil {
    return "Error with cid.GetID", err
  }
  mspID, err := cid.GetMSPID(APIstub)
  if err != nil {
    return "Error with cid.GetMSPID", err
  }

  //Fill local variables with the Driver's information
  Id := mspID + "-" + userID
  firstName := args[0]
  lastName := args[1]
  vehicleMake := args[2]
  vehicleModel := args[3]
  vehicleYear := args[4]

  //Make sure that the user is already registered
  Bytes, err := APIstub.GetState(Id)
  if err != nil {
    return "Error: user not previously registered", errors.New("User not already registered")
  }
  
  //Retrieve ledger version of user and store in local variable
  var user = User{}
  err = json.Unmarshal(Bytes,&user)
  if err != nil {
    return "Error: could not unmarshall user bytes locally", errors.New("User data could not be retrieved from the ledger")
  }
  
  //Update locally stored user information with arguments
  user.FirstName = firstName
  user.LastName = lastName
  user.IsDriver = true
  user.VehicleMake = vehicleMake
  user.VehicleModel = vehicleModel
  user.VehicleYear = vehicleYear
  
  //Marshall the local user variable back into bytes
  Bytes, err = json.Marshal(user)
  if err != nil {
    return "Error: could not marshall user bytes locally", errors.New("User data could not be sent to the ledger")
  }
  
  //Update the ledger with the new user bytes
  err = APIstub.PutState(Id, Bytes)
  if err != nil {
    return "Unable to update user info on the ledger when upgrading to driver. ", errors.New("Unable to PutState new user info to upgrade to driver")
  }

  return "User successfully upgraded to driver", nil
}

//Update the user's name in the ledger
func (s *SmartContract) updateUserName(APIstub shim.ChaincodeStubInterface, args []string) (string, error) {

  //Validate the amount of arguments is correct
  if len(args) != 2 {
    return "Incorrect number of arguments provided!", errors.New("wrong arg amount")
  }

  //Get the specific userID by combining the MSP userID and the mspID
  userID, err := cid.GetID(APIstub)
  if err != nil {
    return "Error with cid.GetID", err
  }
  mspID, err := cid.GetMSPID(APIstub)
  if err != nil {
    return "Error with cid.GetMSPID", err
  }
  Id := mspID + "-" + userID

  //Fill local variables with the user's information
  firstName := args[0]
  lastName := args[1]

  //Retrieve ledger version of user and store in local variable
  Bytes, err := APIstub.GetState(Id)
  if err != nil {
    return "Error: user not previously registered", errors.New("User not already registered")
  }
  var user = User{}
  err = json.Unmarshal(Bytes,&user)
  if err != nil {
    return "Error: could not unmarshall user bytes locally", errors.New("User data could not be retrieved from the ledger")
  }
  
  //Update locally stored user information with arguments
  user.FirstName = firstName
  user.LastName = lastName
  
  //Marshall the local user variable back into bytes
  Bytes, err = json.Marshal(user)
  if err != nil {
    return "Error: could not marshall user bytes locally while calling update user", errors.New("Could not marshall user bytes while calling update user name")
  }
  
  //Update the ledger with the new user bytes
  err = APIstub.PutState(Id, Bytes)
  if err != nil {
    return "Unable to update user info on the ledger when calling update user. ", errors.New("Unable to PutState new user info when calling update user")
  }

  return "User name successfully updated", nil
}

//Delete a user from the ledger
func (s *SmartContract) unregisterUser(APIstub shim.ChaincodeStubInterface, args []string) (string, error) {
  //Validate the amount of arguments is correct
  if len(args) != 1 {
    return "Incorrect number of arguments provided!", errors.New("wrong arg amount")
  }

  //Get the specific userID by combining the MSP userID and the mspID
  userID, err := cid.GetID(APIstub)
  if err != nil {
    return "Error with cid.GetID", err
  }
  mspID, err := cid.GetMSPID(APIstub)
  if err != nil {
    return "Error with cid.GetMSPID", err
  }
  //Fill local variables with the Driver's information
  Id := mspID + "-" + userID
  
  //Make sure that the user is already registered
  r, err := APIstub.GetState(Id)
  if err != nil {
    return "Error: Cannot unregister non-existent user", errors.New("Cannot unregister non-existent user")
  }
  
  //Store the user's ledger bytes in a local variable unmarshalled
  var user = User{}
  err = json.Unmarshal(r,&user)
  
  //Check the user's password is valid before unregistering
  if args[0] != user.Hash {
    return "Invalid password", errors.New("Invalid password: could not unregister user")
  }
  
  //Delete the user from the ledger
  err = APIstub.DelState(Id)
  if err != nil {
	return "Could not unregister user.", errors.New("Could not delete user from the ledger")
  }
  
  return "Driver successfully unregistered", nil
}

//Sends an event to all drivers subscribed to this chaincode that this rider is requesting a ride
//Inputs
//args[0] = pickup location in the format "latitude,longitude" with no quotation marks
func (s *SmartContract) requestRide(APIstub shim.ChaincodeStubInterface, args []string) (string, error) {
  //Check that the number of arguments is correct
  if len(args) != 1 {
    return "Error with args", errors.New("wrong arg amount")
  }
  //Get the global ID by combining the local MSP userID with the mspID
  userID, err := cid.GetID(APIstub)
  if err != nil {
    return "Error getting local MSP userID", err
  }
  mspID, err := cid.GetMSPID(APIstub)
  if err != nil {
    return "Error getting global MSP ID", err
  }
  Id := mspID + "-" + userID
  
  //Check to ensure that the user is registered
  r, err := APIstub.GetState(Id)
  fmt.Printf("length of r: %s", r)
  if len(r) == 0 {
    return "Error: rider not registered", errors.New("Rider not registered")
  }
  
  //Store user's ledger bytes locally for comparison
  var user = User{}
  err = json.Unmarshal(r, &user)
  if err != nil {
	return "Could not unmarshall user data locally", errors.New("Could not unmarshall user data locally while requesting ride")
  }
  
  //Ensure user has filled out needed information
  if user.FirstName == "" || user.LastName == "" {
	return "User has not filled out needed profile information", errors.New("User could not request ride because needed profile information not completed")
  }  
  
  //Check that the user does not have an active ride request
  Bytes, err := APIstub.GetState("RideRequest-" + Id)
  if len(Bytes) != 0 {
    return "Error: user already has a ride active", errors.New("User already has an active ride request")
  }
  
  //Pull the location from the passed in argument
  locationSlice := strings.Split(args[0], ",")
  locationLat := locationSlice[0]
  locationLong := locationSlice[1]
  rideRequest := RideRequest{Id, "", "requested", locationLat, locationLong, "", "", "", "", "", "", "", "", "", ""}
  
  Bytes, marshallErr := json.Marshal(rideRequest)
  if marshallErr != nil {
    return "Error marshalling rideRequest variable: ", errors.New("Error marshalling rideRequest variable")
  }
  
  //Commit new ride request to the ledger
  err = APIstub.PutState("RideRequest-" + Id, Bytes)
  if err != nil {
    return "Error committing the new ride request to the ledger", errors.New("Error committing the new ride request to the ledger")
  }
  
  //Send an event to driver's subscribed to this chaincode that a ride has been requested
  //Eventually we will want to do some filtering to only send it to drivers that are in the area
  payloadStr := "RideRequest-" + Id
  
  APIstub.SetEvent("rideRequest", []byte(payloadStr))
  return "Requested ride as user: " + string(Bytes), nil
}

//Returns the specified user's bytes that are stored in the ledger
func (s *SmartContract) getUserInfo(APIstub shim.ChaincodeStubInterface) (string,error) {
  //Get the global ID by combining the local MSP userID with the mspID
  userID, err := cid.GetID(APIstub)
  if err != nil {
    return "Error getting MSP specific user id: ", err
  }
  mspID, err := cid.GetMSPID(APIstub)
  if err != nil {
    return "Error getting MSP ID: ", err
  }
  Id := mspID + "-" + userID
  Bytes, err := APIstub.GetState(Id)
  if err != nil {
    return "Error getting user's information",  errors.New("Error retrieving user's ledger bytes")
  }
  return string(Bytes), nil
}

//TODO - Remove this function if is not actually needed
//Add driver to the pool of drivers who are currently looking to give a ride
func (s *SmartContract) addDriverToPool(APIstub shim.ChaincodeStubInterface, args []string) (string, error) {
  //Ensure the location of the driver is the only argument given
  if len(args) != 1 {
    return "Error with args", errors.New("wrong arg amount")
  }
  //Get the global ID by combining the local MSP userID with the mspID
  userID, err := cid.GetID(APIstub)
  if err != nil {
    return "Error getting MSP specific user id: ", err
  }
  mspID, err := cid.GetMSPID(APIstub)
  if err != nil {
    return "Error getting MSP ID: ", err
  }
  Id := mspID + "-" + userID
  r, err := APIstub.GetState(Id)
  if len(r) == 0 {
    return "Error, driver not registered", err
  }
  //Create a new entry with the driver's id and location to be added to the pool
  var entry = PoolEntry{Id, args[0]}

  //Try to get the driver pool from the ledger to ensure it exists
  query := "DriverPool"
  Bytes, err := APIstub.GetState(query)

  //Create the driver pool if it does not exist, otherwise append to it
  if err != nil {
    var pool = Pool{entry}
    Bytes, _ = json.Marshal(pool)
    err = APIstub.PutState(query, Bytes)
  } else {
    var pool = Pool{entry}
    err = json.Unmarshal(Bytes,&pool)
    pool = append(pool, entry)
    Bytes, _ = json.Marshal(pool)
    err = APIstub.PutState(query, Bytes)
  }
  response := "Successfully added " + Id + " to the Driver Pool"
  if err != nil {
	  return "Unable to add driver to the pool. Error: ", err
  }
  return response, nil
}


//TODO - remove this function if it not actually needed
//Returns the current pool of drivers including their ID's and locations (Will make internal only after testing)
func (s *SmartContract) getDriverPool(APIstub shim.ChaincodeStubInterface) (string,error) {
  query := "DriverPool"
  Bytes, err := APIstub.GetState(query)
  if err != nil {
    return "Error querying", err
  }
  return string(Bytes), nil
}

//TODO - Remove this function if unneeded
//Deletes a driver's entry from the global pool
func (s *SmartContract) removeDriverFromPool(APIstub shim.ChaincodeStubInterface) (string,error) {
  query := "DriverPool"

  //Get the global ID by combining the local MSP userID with the mspID
  userID, err := cid.GetID(APIstub)
  if err != nil {
    return "Error getting local MSP userID", err
  }
  mspID, err := cid.GetMSPID(APIstub)
  if err != nil {
    return "Error getting global MSP ID", err
  }
  Id := mspID + "-" + userID

  //Query the ledger for the driver pool state
  Bytes, err := APIstub.GetState(query)

  //Create a pool and fill it with the pool in the ledger state
  var pool = Pool{}
  err = json.Unmarshal(Bytes,&pool)

  //Create a new pool that will replace the ledger state pool
  var newPool = Pool{}

  //Iterate through entries and if the user is found within the pool do not add him to the new pool
  for _, entry := range pool {
    if entry.DriverID != Id {
      newPool = append(newPool, entry)
    }
  }
  //Commit the new pool to the ledger
  Bytes, _ = json.Marshal(newPool)
  err = APIstub.PutState(query, Bytes)
  if err != nil {
    return "Error committing the updated pool to the ledger", err
  }
  return "Successfully removed " + Id + " from the Driver Pool", nil
}

//Updates a rideRequest object on the ledger so status is accepted
//Also sends event to rider to notify driver is coming
//Inputs
//args[0] = ride request id
func (s *SmartContract) acceptRide(APIstub shim.ChaincodeStubInterface, args []string) (string, error) {

  //Check that the number of arguments is correct
  if len(args) != 1 {
    return "Error with args", errors.New("wrong arg amount")
  }

  //Get the global ID by combining the local MSP userID with the mspID
  userID, err := cid.GetID(APIstub)
  if err != nil {
    return "Error getting local MSP userID", err
  }
  mspID, err := cid.GetMSPID(APIstub)
  if err != nil {
    return "Error getting global MSP ID", err
  }
  Id := mspID + "-" + userID

  //Retrieve the ride request from the ledger and store locally
  Bytes, err := APIstub.GetState(args[0])
  if err != nil {
	return "Ride request was not found while trying to accept", errors.New("Ride request was not found while trying to accept") 
  }
  var rideRequest = RideRequest{}
  err = json.Unmarshal(Bytes,&rideRequest)
  if err != nil {
	return "Error unmarshalling ride request while trying to accept", errors.New("Error unmarshalling ride request while trying to accept")
  }
  
  //Check the ride was not alreay accepted
  if rideRequest.Status != "requested" {
    return "Ride has already been accepted." + string(Bytes) + " :  " + rideRequest.Status + "  : " + rideRequest.RiderID, errors.New("Ride has already been accepted." + string(Bytes) + " :  " + rideRequest.Status + "  : " + rideRequest.RiderID)
  }
  
  //Update local variable to accepted and add driverID
  rideRequest.Status = "accepted"
  rideRequest.DriverID = Id
  
  //Marshal local variables into bytes
  Bytes, _ = json.Marshal(rideRequest)
 
  //Commit new ride request to the ledger
  err = APIstub.PutState(args[0], Bytes)
  if err != nil {
	return "Ride request was not found while trying to accept", errors.New("Ride request was not found while trying to accept")
  }
  
  //Send event to notify rider that ride has been accepted
  APIstub.SetEvent("rideAccept", []byte("accepted"))
  return "Accepted ride", nil
}

///Retrieves ledger bytes of a provided ride request
//Inputs
//args[0] = ride request id - optional
func (s *SmartContract) getRideRequest(APIstub shim.ChaincodeStubInterface, args []string) (string, error) {

  //If ride request id not supplied get the invoking user's active ride request
  if len(args) == 0 {
  
    //Get the global ID by combining the local MSP userID with the mspID
	userID, err := cid.GetID(APIstub)
	if err != nil {
		return "Error getting local MSP userID", err
	}
	mspID, err := cid.GetMSPID(APIstub)
	if err != nil {
		return "Error getting global MSP ID", err
	}
	Id := mspID + "-" + userID
	
	Bytes, err := APIstub.GetState("RideRequest-" + Id)
	if err != nil {
		return "Ride request id not found while trying to return ride request bytes", errors.New("Ride request id not found while trying to return ride request bytes")
	}
	return string(Bytes), nil
  }
  Bytes, err := APIstub.GetState(args[0])
  if err != nil {
	return "Ride request id not found while trying to return ride request bytes", errors.New("Ride request id not found while trying to return ride request bytes")
  }
  return string(Bytes), nil
}

///Updates an existing ride request with the destination location
//Inputs
//args[0] = destination location in form "latitude,longitude" with no quotes
func (s *SmartContract) setRideDest(APIstub shim.ChaincodeStubInterface, args []string) (string, error) {
  if len(args) != 1 {
    return "Error with args", errors.New("wrong arg amount")
  }
  //Get the global ID by combining the local MSP userID with the mspID
  userID, err := cid.GetID(APIstub)
  if err != nil {
    return "Error getting local MSP userID", err
  }
  mspID, err := cid.GetMSPID(APIstub)
  if err != nil {
    return "Error getting global MSP ID", err
  }
  Id := mspID + "-" + userID
  
  //Retrieve ride request bytes from the ledger and store locally
  Bytes, err := APIstub.GetState("RideRequest-" + Id)
  if err != nil {
	return "Ride request id not found while trying to set destination location", errors.New("Ride request id not found while trying to set destination location")
  }
  var rideRequest = RideRequest{}
  err = json.Unmarshal(Bytes,&rideRequest)
  if err != nil {
	return "Error unmarshalling ride request id bytes while setting destination location", errors.New("Error unmarshalling ride request id bytes while setting destination location")
  }
  
  //Pull the location from the passed in argument
  locationSlice := strings.Split(args[0], ",")
  locationLat := locationSlice[0]
  locationLong := locationSlice[1]
  
  //Store location in local ride request variable
  rideRequest.DropoffLocationLat = locationLat
  rideRequest.DropoffLocationLong = locationLong
 
  //Marshall updated ride request into bytes and commit to ledger
  Bytes, _ = json.Marshal(rideRequest)
  err = APIstub.PutState("RideRequest-" + Id, Bytes)
  if err != nil {
	return "Ride id was not found: ", err
  }
  
  return "Added destination to rideRequest", nil
}

//Sets corider pickup location and the corider's ride id in an active ride request
//Inputs
//args[0] = ride request id
//args[1] = corider id
//args[2] = corider pickup location latitude
//args[3] = corider pickup location longitude
func (s *SmartContract) setCoriderPickup(APIstub shim.ChaincodeStubInterface, args []string) (string, error) {
  
  //Check argument length is correct
  if len(args) != 4 {
    return "Error with args", errors.New("wrong arg amount")
  }  
 
  //Get the active ride request from the ledger and store it locally
  Bytes, err := APIstub.GetState(string(args[0]))
  if err != nil {
	return "Ride request id not found in ledger while updating corider pickup location", errors.New("Ride request id not found in ledger while updating corider pickup location")
  }
  var rideRequest = RideRequest{}
  err = json.Unmarshal(Bytes,&rideRequest)
  if err != nil {
    return "Error unmarshalling active ride request while updating corider pickup location", errors.New("Error unmarshalling active ride request while updating corider pickup location")
  }
  
  //Update local ride request with inputs
  rideRequest.CoriderID = string(args[1])
  rideRequest.CoriderPickupLocationLat = string(args[2])
  rideRequest.CoriderPickupLocationLong = string(args[3])
  
  //Marshall local variable and commit to ledger
  Bytes, err = json.Marshal(rideRequest)
  if err != nil {
    return "Error committing ride request to ledger after updating corider pickup location", errors.New("Error committing ride request to ledger after updating corider pickup location")
  }
  err = APIstub.PutState(string(args[0]), Bytes)
  if err != nil {
	return "Error committing ride request to ledger after updating corider pickup location", errors.New("Error committing ride request to ledger after updating corider pickup location")
  }

  return "Added corider ID and pickup to rideRequest", nil
}

//Sets corider dropoff location and the corider's ride id in an active ride request
//Inputs
//args[0] = ride request id
//args[1] = corider id
//args[2] = corider dropoff location latitude
//args[3] = corider dropoff location longitude
func (s *SmartContract) setCoriderDropoff(APIstub shim.ChaincodeStubInterface, args []string) (string, error) {
  
  //Check argument length is correct
  if len(args) != 4 {
    return "Error with args", errors.New("wrong arg amount")
  }  
 
  //Get the active ride request from the ledger and store it locally
  Bytes, err := APIstub.GetState(string(args[0]))
  if err != nil {
	return "Ride id was not found: ", err
  }
  var rideRequest = RideRequest{}
  err = json.Unmarshal(Bytes,&rideRequest)
  
  //Update loca ride request with inputs
  rideRequest.CoriderID = string(args[1])
  rideRequest.CoriderDropoffLocationLat = string(args[2])
  rideRequest.CoriderDropoffLocationLong = string(args[3])
  
  //Marshall local variable and commit to ledger
  Bytes, err = json.Marshal(rideRequest)
  if err != nil {
    return "Error marshalling ride request while updating corider dropoff location", errors.New("Error marshalling ride request while updating corider dropoff location")
  }
  err = APIstub.PutState(string(args[0]), Bytes)
  if err != nil {
	return "Error committing marshalled ride request while updating corider dropoff location", errors.New("Error committing marshalled ride request while updating corider dropoff location")
  }
  
  return "Added corider ID and pickup to rideRequest", nil
}

//Updates status and pickup time of active ride request
//Inputs
//args[0] = ride request id
//args[1] = pickup time
func (s *SmartContract) pickupRider(APIstub shim.ChaincodeStubInterface, args []string) (string, error) {

  //Check argument length is correct
  if len(args) != 2 {
    return "Error with args", errors.New("wrong arg amount")
  }  

  //Get the active ride request from the ledger and store it locally
  Bytes, err := APIstub.GetState(args[0])
  if err != nil {
	return "Ride request id not found in ledger while calling pickup rider", errors.New("Ride request id not found in ledger while calling pickupRider")
  }
  var rideRequest = RideRequest{}
  err = json.Unmarshal(Bytes,&rideRequest)
  
  //Update local ride request
  rideRequest.Status = "ongoing"
  rideRequest.PickupTime = args[1]
  
  //Marshall local variable and commit to ledger
  Bytes, err = json.Marshal(rideRequest)
  if err != nil {
    return "Error marshalling ride request while calling pickup rider", errors.New("Error marshalling ride request while calling pickup rider")
  }
  err = APIstub.PutState(args[0], Bytes)
  if err != nil {
	return "Error committing ride request to ledger while calling pickup rider", errors.New("Error committing ride request to ledger while calling pickup rider")
  }
  
  //Send event to rider to notify them of pickup
  APIstub.SetEvent("ridePickup", []byte(args[0]))
  return "Picked up rider", nil
}

//Commits ride request to permanent storage on ledger for rider and deletes ride request
//
//No inputs
func (s *SmartContract) leaveDriver(APIstub shim.ChaincodeStubInterface, args []string) (string, error) {
  
  //Get the global ID by combining the local MSP userID with the mspID
  userID, err := cid.GetID(APIstub)
  if err != nil {
    return "Error getting local MSP userID", err
  }
  mspID, err := cid.GetMSPID(APIstub)
  if err != nil {
    return "Error getting global MSP ID", err
  }
  Id := mspID + "-" + userID

  //Get the active ride request from the ledger and store it locally
  Bytes, err := APIstub.GetState("RideRequest-" + Id)
  if err != nil {
    return "Ride request id not found in ledger while calling leave driver", errors.New("Ride request id not found in ledger while calling leave driver")
  }
  var rideRequest = RideRequest{}
  err = json.Unmarshal(Bytes,&rideRequest)
  if err != nil {
    return "Error unmarshalling active ride request while calling leave driver", errors.New("Error unmarshalling active ride request while calling leave driver")
  }
  
  //Create a new ride taken object with the active ride requests information
  var ride = Ride{"RideRequest-" + Id, Id, rideRequest.DriverID, rideRequest.PickupTime, rideRequest.DropoffTime, rideRequest.PickupLocationLat + "," + rideRequest.PickupLocationLong, rideRequest.DropoffLocationLat + "," + rideRequest.DropoffLocationLong, rideRequest.Distance, "No prices set yet", rideRequest.CoriderID, rideRequest.CoriderPickupLocationLat + "," + rideRequest.CoriderPickupLocationLong, rideRequest.CoriderDropoffLocationLat + "," + rideRequest.CoriderDropoffLocationLong}
  Bytes, err = json.Marshal(ride)
  if err != nil {
    return "Error marshalling ride request while calling leave driver", errors.New("Error marshalling ride request while calling leave driver")
  }
  
  //Commit new ride taken to the ledger
  err = APIstub.PutState(Id + "-rideTaken-1", Bytes)
  if err != nil {
    return "Error committing permanent ride taken to ledger while calling leave driver", errors.New("Error committing permanent ride taken to ledger while calling leave driver")
  }
  
  //Delete the temporal ride request from the ledger
  err = APIstub.DelState("RideRequest-" + Id)
  if err != nil {
    return "Error deleting temporal ride request while calling leave driver", errors.New("Error deleting temporal ride request while calling leave driver")
  }
  
  return "Successfully left driver and updated ride in ledger", nil
}

//Commits ride request to permanent storage on ledger for the driver
//Also notifies rider that the ride is ending
//Inputs
//Args[0] - ride request id
//args[1] - dropoff time
func (s *SmartContract) dropoffRider(APIstub shim.ChaincodeStubInterface, args []string) (string, error) {

  //Check argument length is correct
  if len(args) != 2 {
    return "Error with args", errors.New("wrong arg amount")
  } 

  //Get the global ID by combining the local MSP userID with the mspID
  userID, err := cid.GetID(APIstub)
  if err != nil {
    return "Error getting local MSP userID", err
  }
  mspID, err := cid.GetMSPID(APIstub)
  if err != nil {
    return "Error getting global MSP ID", err
  }
  Id := mspID + "-" + userID
	
  //Get the active ride request from the ledger and store it locally
  Bytes, err := APIstub.GetState(args[0])
  if err != nil {
	return "Ride request id not found in ledger while calling dropoff rider", errors.New("Ride request id not found in ledger calling dropoff rider")
  }
  var rideRequest = RideRequest{}
  err = json.Unmarshal(Bytes,&rideRequest)
  if err != nil {
    return "Error unmarshalling active ride request while calling dropoff rider", errors.New("Error unmarshalling active ride request while calling dropoff rider")
  }
  
  //Update local ride request with inputs
  rideRequest.Status = "completed"
  rideRequest.DropoffTime = args[1]
  rideRequest.Distance = "0"
  
  //Marshall local variable and commit to ledger
  Bytes, _ = json.Marshal(rideRequest)
  err = APIstub.PutState(args[0], Bytes)
  if err != nil {
	return "Ride id was not found: ", err
  }
  
  
  //Create new ride and commit to ledger
  var ride = Ride{args[0],rideRequest.RiderID, Id, rideRequest.PickupTime, rideRequest.DropoffTime, rideRequest.PickupLocationLat + "," + rideRequest.PickupLocationLong, rideRequest.DropoffLocationLat + "," + rideRequest.DropoffLocationLong, rideRequest.Distance, "No prices set yet", rideRequest.CoriderID, rideRequest.CoriderPickupLocationLat + "," + rideRequest.CoriderPickupLocationLong, rideRequest.CoriderDropoffLocationLat + "," + rideRequest.CoriderDropoffLocationLong}
  
  Bytes, err = json.Marshal(ride)
  if err != nil {
    return "Error marshalling ride request while calling dropoff rider", errors.New("Error marshalling ride request while calling dropoff rider")
  }
  err = APIstub.PutState(Id + "-rideGiven-1", Bytes)
  if err != nil {
    return "Error committing ride to ledger while calling dropoff rider", errors.New("Error committing ride to ledger while calling dropoff rider")
  }

  //Notify rider that rider is ending
  APIstub.SetEvent("rideDropoff", []byte(args[0]))
  
  return "Dropped off rider", nil
}


func main() {
  err := shim.Start(new(SmartContract))
  if err != nil {
    fmt.Println("Could not start SmartContract")
  } else {
    fmt.Println("SmartContract successfully started")
  }
}
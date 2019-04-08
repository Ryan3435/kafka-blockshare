import QtQuick 2.7			//Item
//import QtQuick.Controls 1.4 as QQC1	//ButtonStyle
import QtQuick.Controls 2.1	//Dialog
import QtQuick.Layouts 1.3
import QtLocation 5.6
import QtPositioning 5.9
import QtQml 2.11 //timer
import QtGraphicalEffects 1.0

Item {
	id: window
	
	height: 500
	width: 300
	
	Item
	{
		id: background
        anchors.fill: parent
		Image
		{
			id: backgroundPic
			anchors.fill: parent
			source: "pics/phoneRoad.jpg"
		}
		FastBlur {
			id: backgroundBlur
			anchors.fill: backgroundPic
			source: backgroundPic
			radius: 0
		}
		
	}
	
	Plugin {
        id: mapPlugin
        name: "esri" // "mapboxgl", "esri", ...
    }
	Plugin {
        id: mapPlugin2
        name: "osm" // "mapboxgl", "esri", ...
    }
	
	Column {
		anchors.centerIn: parent
		
		
		Item {
			width: 1 // dummy value != 0
			height: 50
		}
		Text {
			anchors.horizontalCenter: parent.horizontalCenter
			text: "HyFRA"
			font.family: "Verdana"
			style: Text.Outline 
			styleColor: "black" 
			font.pointSize: 26
			color: "#FFFFFF"
		}
		Item {
			width: 1 // dummy value != 0
			height: 15
		}
		Text {
			anchors.horizontalCenter: parent.horizontalCenter
			text: "Hyperledger Fabric"
			font.family: "Verdana"
			style: Text.Outline 
			styleColor: "black" 
			font.pointSize: 14
			color: "#FFFFFF"
		}
		Text {
			anchors.horizontalCenter: parent.horizontalCenter
			text: "Ride-sharing Application"
			font.family: "Verdana"
			style: Text.Outline 
			styleColor: "black" 
			font.pointSize: 14
			color: "#FFFFFF"
		}
		Item {
			width: 1 // dummy value != 0
			height: 210
		}

		Button {
			id: registerButton
			anchors.topMargin: 200
			anchors.horizontalCenter: parent.horizontalCenter
			text: "Register"
			onClicked: 
			{
				backgroundBlur.radius = 32
				registerUserDialog.open()
			}
			width: 200
		}
		Button {
			id: loginButton
			anchors.horizontalCenter: parent.horizontalCenter
			text: "Login"
			onClicked: 
			{
				backgroundBlur.radius = 32
				loginDialog.open()
			}
			width: 200
		}

		Item {
			width: 1 // dummy value != 0
			height: 20
		}
		
		Text {
			anchors.horizontalCenter: parent.horizontalCenter
			text: "Developed at Tennessee Tech University"
			font.family: "Verdana"
			font.pointSize: 10
			color: "#FFFFFF"
			style: Text.Outline 
			styleColor: "black"
		}
		
		Connections
		{
			target: qmlBridge
			onDriverLogin: driverFunctionsDialog.open()
			onRiderLogin: riderFunctionsDialog.open()
		}
	}
	
	Dialog {
		id: driverFunctionsDialog

		x: 0
		y: 0

		contentWidth: window.width
		contentHeight: window.height
		property var startLat: 36.174970
		property var startLong: -85.516170
		
		
		contentItem: Rectangle {
			x: 0
			y: 0
			color: "#4F2984"
			anchors.fill: parent
		}
		
		Column {
			anchors.centerIn: parent

			Button {
				anchors.horizontalCenter: parent.horizontalCenter
				text: "Start Driving"
				width: window.width * 0.75
				height: window.height * 0.33
				onClicked: 
				{
					startDrivingDialog.open()
				}
			}
			Button {
				width: window.width * 0.75
				height: window.height * 0.33
				anchors.horizontalCenter: parent.horizontalCenter
				text: "Logout"
				onClicked: driverFunctionsDialog.close()
			}
		}
	}
	
	Dialog {
		id: startDrivingDialog

		x: (window.width - width) * 0.5
		y: (window.height - height) * 0.5

		contentWidth: window.width * 0.7
		contentHeight: window.height * 0.45
		
		property var startLat: 0
		property var startLong: 0
		
		standardButtons: Dialog.Cancel | Dialog.Ok
		onAccepted: 
		{
			map.addMapItem(mapItemView)
			aQuery.clearWaypoints()
			fromAddress.street = startDrivingStreet.text
			fromAddress.city = startDrivingCity.text
			fromAddress.country = startDrivingCountry.text
			fromAddress.state = startDrivingState.text
			fromAddress.postalCode = startDrivingPostalCode.text
			geocodeModel.query = fromAddress
			geocodeModel.update()
		}

		Column {
			anchors.centerIn: parent
			Text {
				anchors.horizontalCenter: parent.horizontalCenter
				text: "Enter your current address"
				width: window.width * 0.75
			}
			TextField {
				id: startDrivingStreet
		
				anchors.horizontalCenter: parent.horizontalCenter
				placeholderText: "Street address"
				width: window.width * 0.75
			}
			TextField {
				id: startDrivingCity
		
				anchors.horizontalCenter: parent.horizontalCenter
				placeholderText: "City"
				width: window.width * 0.75
			}
			TextField {
				id: startDrivingCountry
		
				anchors.horizontalCenter: parent.horizontalCenter
				placeholderText: "Country"
				width: window.width * 0.75
			}
			TextField {
				id: startDrivingState
		
				anchors.horizontalCenter: parent.horizontalCenter
				placeholderText: "State"
				width: window.width * 0.75
			}
			TextField {
				id: startDrivingPostalCode
		
				anchors.horizontalCenter: parent.horizontalCenter
				placeholderText: "Postal Code"
				width: window.width * 0.75
			}
		}
		
		GeocodeModel {
			id: geocodeModel
			plugin: map.plugin
			onStatusChanged: {
				if (status == GeocodeModel.Ready) {
					startDrivingDialog.startLat = get(0).coordinate.latitude
					startDrivingDialog.startLong = get(0).coordinate.longitude
					map.currentLat = startDrivingDialog.startLat
					map.currentLong = startDrivingDialog.startLong
					drivingDialog.open()
					qmlBridge.goFunction(loginUserInput.text, loginOrgInput.text, "startDriving",startDrivingDialog.startLat, startDrivingDialog.startLong, "", "", "", "", "")			
				}
			}
		}
		
		Address {
			id :fromAddress
			street: ""
			city: ""
			country: ""
			state : ""
			postalCode: ""
		}
	}
	Dialog {
		id: drivingDialog

		x: 0
		y: 0
		
		width: window.width
		height: window.height
		
		contentItem: Map {
			property var rideIDs: []
			property var pLatArray: []
			property var pLongArray: []
			property var dLatArray: []
			property var dLongArray: []	
			property var inCarArray: []	
			property var numRides: 0
			property var driveType: "none"
			property var currentRide: []
			property var tempWaypoint: 0
			property var currentLat: 0
			property var currentLong: 0
			property var stopDriving: []
			property int currentRideIndex: 0
		
			RouteQuery {
				id: aQuery
				travelModes: RouteQuery.CarTravel
			}

			RouteModel {
				id: routeModel
				plugin: mapPlugin2
				query: aQuery
				autoUpdate: false
				property var lats: [];
				property var longs: [];
				onStatusChanged: {
					console.log(RouteModel.Ready)
					if (status == RouteModel.Ready) {
						lats = []
						longs = []
						for(var i = 0; i < routeModel.get(0).path.length; i++) {
							lats[i] = routeModel.get(0).path[i].latitude
							longs[i] = routeModel.get(0).path[i].longitude
						}
						map.stopDriving[map.currentRideIndex] = true
						if (map.driveType == "dropoff") {
							map.currentRideIndex = 0
						} else {
							map.currentRideIndex = map.numRides - 1
						}
						map.stopDriving[map.currentRideIndex] = false
						console.log(map.currentRideIndex)
						var msg = {'lats': lats, 'longs': longs, 'driveType': map.driveType, 'user': loginUserInput.text, 'org': loginOrgInput.text, 'rideID': map.rideIDs[map.currentRideIndex], 'index': map.currentRideIndex};	
						console.log("asdf")
						driveWorker.sendMessage(msg)
						console.log("ff")
					}
				}
			}
			
			WorkerScript {
				id: driveWorker
				source: "driveWorker.js"
				
				onMessage:
				{
					if (map.stopDriving[messageObject.index] == false) {
						qmlBridge.updateDriverVisual(messageObject.lat, messageObject.long)
						map.currentLat = messageObject.lat
						map.currentLong = messageObject.long
						
						if (messageObject.driveType == "pickup" && messageObject.last == 'yes') {
							qmlBridge.pickupRiderCC(messageObject.rideID, messageObject.user, messageObject.org)
						}
						else if (messageObject.driveType == "dropoff" && messageObject.last == 'yes') {
							qmlBridge.dropoffRiderCC(messageObject.rideID, messageObject.user, messageObject.org)
						}
					}
				}
			}
			
			MapItemView {
				id: mapItemView
				model: routeModel
				delegate: routeDelegate
			}
			
			Component {
				id: routeDelegate
				
				MapRoute {
					id: mapRoute
					route: routeData
					line.color: "blue"
					line.width: 5
					smooth: true
					opacity: 0.8
				}
			}
		
		
			id: map
			plugin: mapPlugin
			anchors.fill: parent
			center: QtPositioning.coordinate(map.currentLat, map.currentLong) // Oslo
			zoomLevel: 16
			width: window.width
			height: window.height
			
			MapCircle {
				id: mapCircle
				center {
					latitude: map.currentLat
					longitude: map.currentLong
				}
				radius: 15.0
				color: 'green'
				border.width: 1
			}
			
			Dialog {
				id: newRideRequestDialog

				x: (window.width - width) * 0.3
				y: (window.height - height) * 0.3

				contentWidth: window.width * 0.7
				contentHeight: window.height * 0.7
				
				standardButtons: Dialog.Cancel | Dialog.Ok
				onAccepted: 
				{
					qmlBridge.driverAcceptedRideRequestSlot(map.pLatArray[map.numRides], map.pLongArray[map.numRides], map.rideIDs[map.numRides])
				}
				onRejected:
				{
					map.pLatArray.slice(map.numRides,1)
					map.pLongArray.slice(map.numRides,1)
					map.rideIDs.slice(map.numRides,1)
					map.inCarArray.slice(map.numRides,1)
					map.stopDriving[map.currentRideIndex] = false
				}
				
				Column {
					anchors.centerIn: parent
					Text {
						text: "New Ride Request. "
					}
					Text {
						id: latitudeText
						text: "Would you like to accept?"
					}
				}
			}
			
		}
		Connections
		{
			target: qmlBridge
			onUpdateDriver: 
			{
				mapCircle.center = QtPositioning.coordinate(latitude, longitude)
				map.center = QtPositioning.coordinate(latitude, longitude)
				map.currentLat = latitude
				map.currentLong = longitude
			}
			onNewRideRequest:
			{

				if (map.numRides < 3) {
					map.rideIDs.push(rideID)
					map.pLatArray.push(latitude)
					map.pLongArray.push(longitude)
					map.inCarArray.push(false)
					map.stopDriving[map.currentRideIndex] = true
					newRideRequestDialog.open()
				}
			}
			onDriverAcceptedRideRequestSignal:
			{		
				qmlBridge.goFunction(loginUserInput.text, loginOrgInput.text, "acceptRide", rideID, "", "", "", "", "", "")	
				console.log(rideID)
				map.numRides += 1
				//map.stopDriving.push(true)
				//map.driveType = "pickup"
				aQuery.clearWaypoints()
				routeModel.reset()
				map.driveType = "pickup"
				aQuery.addWaypoint(QtPositioning.coordinate(map.currentLat, map.currentLong))
				aQuery.addWaypoint(QtPositioning.coordinate(latitude, longitude))
				routeModel.update()
			}
			onRiderPickup:
			{
				var i;
				for (i = 0; i < map.numRides; i++) {
					if (map.currentRideIndex != i && map.inCarArray[i] == true) {
						qmlBridge.updateCoriderPickupCC(map.rideIDs[i], map.rideIDs[map.currentRideIndex],map.pLatArray[map.currentRideIndex],map.pLongArray[map.currentRideIndex],loginUserInput.text, loginOrgInput.text)
					}
				}
				map.dLatArray[map.currentRideIndex] = latitude
				map.dLongArray[map.currentRideIndex] = longitude
				map.inCarArray[map.currentRideIndex] = true
				aQuery.clearWaypoints()
				routeModel.reset()
				map.driveType = "dropoff"
				aQuery.addWaypoint(QtPositioning.coordinate(map.currentLat, map.currentLong))
				aQuery.addWaypoint(QtPositioning.coordinate(map.dLatArray[0], map.dLongArray[0]))
				routeModel.update()
			}
			onDropoffRider:
			{
				var finishedRideIndex = map.rideIDs.indexOf(rideID)
				console.log("Dropping off rider ", finishedRideIndex)
				var i;
				for (i = 0; i < map.numRides; i++) {
					console.log("Should I update corider for driver ", i)
					if (finishedRideIndex != i && map.inCarArray[i] == true) {
						console.log("Yes! Doing so now")
						qmlBridge.updateCoriderDropoffCC(map.rideIDs[i], map.rideIDs[finishedRideIndex],map.dLatArray[finishedRideIndex],map.dLongArray[finishedRideIndex],loginUserInput.text, loginOrgInput.text)
					}
				}
				map.numRides -= 1
				map.pLatArray.shift() //= map.pLatArray[1:] //.slice(finishedRideIndex,1)
				map.pLongArray.shift() //= map.pLongArray[1:] //.slice(finishedRideIndex,1)
				map.rideIDs.shift() //= map.rideIDs[1:] //.slice(finishedRideIndex,1)
				map.inCarArray.shift() //= map.inCarArray[1:] //.slice(finishedRideIndex,1)
				map.dLatArray.shift() //= map.dLatArray[1:] //.slice(finishedRideIndex,1)
				map.dLongArray.shift() //= map.dLongArray[1:] //.slice(finishedRideIndex,1)
				if (map.numRides == 0) {
					aQuery.clearWaypoints()
					routeModel.reset()
					doneDrivingDialog.open()
				}
				else {
					aQuery.clearWaypoints()
					routeModel.reset()
					map.driveType = "dropoff"
					aQuery.addWaypoint(QtPositioning.coordinate(map.currentLat, map.currentLong))
					aQuery.addWaypoint(QtPositioning.coordinate(map.dLatArray[0], map.dLongArray[0]))
					routeModel.update()
				}
			}
		}
		
	}	
	
	Dialog {
		id: riderFunctionsDialog

		x: 0
		y: 0

		contentWidth: window.width
		contentHeight: window.height
		
		contentItem: Rectangle {
			x: 0
			y: 0
			color: "#1e91bf"
			anchors.fill: parent
		}
		

		Column {
			anchors.centerIn: parent

			Button {
				anchors.horizontalCenter: parent.horizontalCenter
				text: "Update Profile"
				width: window.width * 0.75
				height: window.height * 0.2
				onClicked: updateProfileDialog.open()
			}
			Button {
				anchors.horizontalCenter: parent.horizontalCenter
				width: window.width * 0.75
				height: window.height * 0.2
				text: "Upgrade to Driver"
				onClicked: upgradeToDriverDialog.open()
			}
			Button {
				anchors.horizontalCenter: parent.horizontalCenter
				text: "Request Ride"
				width: window.width * 0.75
				height: window.height * 0.2
				onClicked: requestRideDialog.open()
			}
			Button {
				anchors.horizontalCenter: parent.horizontalCenter
				text: "Logout"
				width: window.width * 0.75
				height: window.height * 0.2
				onClicked: riderFunctionsDialog.close()
			}
		}
	}
	
	Dialog {
		id: requestRideDialog

		x: (window.width - width) * 0.5
		y: (window.height - height) * 0.5

		contentWidth: window.width * 0.7
		contentHeight: window.height * 0.55
		
		property var startLat: 0
		property var startLong: 0
		property var endLat: 0
		property var endLong: 0
		
		standardButtons: Dialog.Cancel | Dialog.Ok
		onAccepted: 
		{
			fromAddressRR.street = startStreet.text
			fromAddressRR.city = startCity.text
			fromAddressRR.country = "United States"
			fromAddressRR.state = "TN"
			fromAddressRR.postalCode = startPostalCode.text
			geocodeModelFrom.query = fromAddressRR
			geocodeModelFrom.update()
			
			toAddressRR.street = endStreet.text
			toAddressRR.city = endCity.text
			toAddressRR.country = "United States"
			toAddressRR.state = "TN"
			toAddressRR.postalCode = endPostalCode.text
			geocodeModelTo.query = toAddressRR
			geocodeModelTo.update()
		}

		Column {
			anchors.centerIn: parent
			Text {
				anchors.horizontalCenter: parent.horizontalCenter
				text: "Enter your current address"
				width: window.width * 0.75
			}
			TextField {
				id: startStreet
		
				anchors.horizontalCenter: parent.horizontalCenter
				placeholderText: "Street address"
				width: window.width * 0.75
			}
			TextField {
				id: startCity
		
				anchors.horizontalCenter: parent.horizontalCenter
				placeholderText: "City"
				width: window.width * 0.75
			}
			TextField {
				id: startPostalCode
		
				anchors.horizontalCenter: parent.horizontalCenter
				placeholderText: "Postal Code"
				width: window.width * 0.75
			}
			Text {
				anchors.horizontalCenter: parent.horizontalCenter
				text: "Enter your destination address"
			}
			TextField {
				id: endStreet
		
				anchors.horizontalCenter: parent.horizontalCenter
				placeholderText: "Street address"
				width: window.width * 0.75
			}
			TextField {
				id: endCity
		
				anchors.horizontalCenter: parent.horizontalCenter
				placeholderText: "City"
				width: window.width * 0.75
			}
			TextField {
				id: endPostalCode
		
				anchors.horizontalCenter: parent.horizontalCenter
				placeholderText: "Postal Code"
				width: window.width * 0.75
			}
		}
		
		GeocodeModel {
			id: geocodeModelFrom
			plugin: map.plugin
			onStatusChanged: {
				if (status == GeocodeModel.Ready) {
					requestRideDialog.startLat = get(0).coordinate.latitude
					requestRideDialog.startLong = get(0).coordinate.longitude
					
					if (requestRideDialog.endLat != 0 && requestRideDialog.endLong != 0) {
						qmlBridge.goFunction(loginUserInput.text, loginOrgInput.text, "requestRide", requestRideDialog.startLat, requestRideDialog.startLong, requestRideDialog.endLat, requestRideDialog.endLong, "", "", "")//requestRideDialog.open()
					}
				}
			}
		}
		
		GeocodeModel {
			id: geocodeModelTo
			plugin: map.plugin
			onStatusChanged: {
				if (status == GeocodeModel.Ready) {
					requestRideDialog.endLat = get(0).coordinate.latitude
					requestRideDialog.endLong = get(0).coordinate.longitude

					console.log(requestRideDialog.startLat)
					console.log(requestRideDialog.startLong)
					console.log(requestRideDialog.endLat)
					console.log(requestRideDialog.endLong)
					if (requestRideDialog.startLat != 0 && requestRideDialog.startLong != 0) {
						qmlBridge.goFunction(loginUserInput.text, loginOrgInput.text, "requestRide", requestRideDialog.startLat, requestRideDialog.startLong, requestRideDialog.endLat, requestRideDialog.endLong, "", "", "")//requestRideDialog.open()
					}					
				}
			}
		}
		
		Address {
			id :fromAddressRR
			street: ""
			city: ""
			country: ""
			state : ""
			postalCode: ""
		}
		
		Address {
			id :toAddressRR
			street: ""
			city: ""
			country: ""
			state : ""
			postalCode: ""
		}
	}
	
	Dialog {
		id: doneDrivingDialog

		x: (window.width - width) * 0.5
		y: (window.height - height) * 0.5

		contentWidth: window.width * 0.7
		contentHeight: window.height * 0.7
		standardButtons: Dialog.Ok

		Column {
			anchors.centerIn: parent

			Text {
				text: "Waiting for a new ride request"
			}
		}
	}
	
	Dialog {
		id: pickupRiderDialog

		x: (window.width - width) * 0.5
		y: (window.height - height) * 0.5

		contentWidth: window.width * 0.7
		contentHeight: window.height * 0.7

		Column {
			anchors.centerIn: parent

			Text {
				text: "Picking up Rider"
			}
		}
	}

	Dialog {
		id: registerUserDialog

		x: (window.width - width) * 0.5
		y: (window.height - height) * 0.5

		contentWidth: window.width * 0.7
		contentHeight: window.height * 0.33
		standardButtons: Dialog.Cancel | Dialog.Ok
		onRejected:
		{
			backgroundBlur.radius = 0
		}
		onAccepted:
		{
			qmlBridge.goFunction(userRegisterUserInput.text, userRegisterOrgInput.text, "registerUser", userRegisterPasswordInput.text, userRegisterReenterPasswordInput.text, "", "", "", "", "")
			backgroundBlur.radius = 0
		}

		Column {
			anchors.left: parent.left
			anchors.top: parent.top
			width: parent.width
			TextField {
				id: userRegisterUserInput
				width: parent.width
				anchors.left: parent.left
				placeholderText: "User ID"
			}
			TextField {
				id: userRegisterOrgInput
				
				anchors.left: parent.left
				placeholderText: "Organization"
				width: parent.width
			}
			TextField {
				id: userRegisterPasswordInput
		
				anchors.left: parent.left
				placeholderText: "Enter password"
				width: parent.width
			}
			TextField {
				id: userRegisterReenterPasswordInput
		
				anchors.left: parent.left
				placeholderText: "Re-enter password"
				width: parent.width
			}
		}
	}
	Dialog {
		id: loginDialog

		x: (window.width - width) * 0.5
		y: (window.height - height) * 0.5

		contentWidth: window.width * 0.7
		contentHeight: window.height * 0.3
		standardButtons: Dialog.Cancel | Dialog.Ok
		onRejected:
		{
			backgroundBlur.radius = 0
		}
		onAccepted: 
		{
			backgroundBlur.radius = 0
			if (driverRadio.checked == true) {
				qmlBridge.goFunction(loginUserInput.text, loginOrgInput.text, "loginDriver",loginPasswordInput.text, "", "", "", "", "", "")
			} else {
				qmlBridge.goFunction(loginUserInput.text, loginOrgInput.text, "loginRider", loginPasswordInput.text, "", "", "", "", "", "")
			}
		}

		Column {
			anchors.left: parent.left
			anchors.top: parent.top
			width: parent.width
			TextField {
				id: loginUserInput
		
				anchors.horizontalCenter: parent.horizontalCenter
				placeholderText: "User ID"
				width: parent.width
				anchors.left: parent.left
			}
			TextField {
				id: loginOrgInput
		
				anchors.horizontalCenter: parent.horizontalCenter
				placeholderText: "Organization"
				width: parent.width
				anchors.left: parent.left
			}
			TextField {
				id: loginPasswordInput
		
				anchors.horizontalCenter: parent.horizontalCenter
				placeholderText: "Password"
				width: parent.width
				anchors.left: parent.left
			}
			RowLayout {
				anchors.horizontalCenter: parent.horizontalCenter
				RadioButton {
					id: riderRadio
					checked: true
					text: qsTr("Rider")
				}
				RadioButton {
					id: driverRadio
					text: qsTr("Driver")
				}
			}
		}
	}
	
	Dialog {
		id: updateProfileDialog

		x: (window.width - width) * 0.5
		y: (window.height - height) * 0.5

		contentWidth: window.width * 0.7
		contentHeight: window.height * 0.15
		standardButtons: Dialog.Cancel | Dialog.Ok
		onAccepted: qmlBridge.goFunction(loginUserInput.text, loginOrgInput.text, "updateUserName", updateProfileFirstName.text, updateProfileLastName.text, "", "", "", "", "", "", "")

		Column {
			anchors.left: parent.left
			anchors.top: parent.top
			width: parent.width
			TextField {
				id: updateProfileFirstName
		
				anchors.horizontalCenter: parent.horizontalCenter
				placeholderText: "First Name"
				width: parent.width
				anchors.left: parent.left
			}
			TextField {
				id: updateProfileLastName
		
				anchors.horizontalCenter: parent.horizontalCenter
				placeholderText: "Last Name"
				width: parent.width
				anchors.left: parent.left
			}
		}
	}
	
	Dialog {
		id: upgradeToDriverDialog

		x: (window.width - width) * 0.5
		y: (window.height - height) * 0.5

		contentWidth: window.width * 0.7
		contentHeight: window.height * 0.37
		standardButtons: Dialog.Cancel | Dialog.Ok
		onAccepted: qmlBridge.goFunction(loginUserInput.text, loginOrgInput.text, "upgradeToDriver", upgradeToDriverFirstName.text, upgradeToDriverLastName.text, upgradeToDriverVehicleMake.text, upgradeToDriverVehicleModel.text, upgradeToDriverVehicleYear.text, "", "", "", "")

		Column {
			anchors.left: parent.left
			anchors.top: parent.top
			width: parent.width
			TextField {
				id: upgradeToDriverFirstName
		
				anchors.horizontalCenter: parent.horizontalCenter
				placeholderText: "First Name"
				width: parent.width
				anchors.left: parent.left
			}
			TextField {
				id: upgradeToDriverLastName
		
				anchors.horizontalCenter: parent.horizontalCenter
				placeholderText: "Last Name"
				width: parent.width
				anchors.left: parent.left
			}
			TextField {
				id: upgradeToDriverVehicleMake
		
				anchors.horizontalCenter: parent.horizontalCenter
				placeholderText: "Vehicle Make"
				width: parent.width
				anchors.left: parent.left
			}
			TextField {
				id: upgradeToDriverVehicleModel
		
				anchors.horizontalCenter: parent.horizontalCenter
				placeholderText: "Vehicle Model"
				width: parent.width
				anchors.left: parent.left
			}
			TextField {
				id: upgradeToDriverVehicleYear
		
				anchors.horizontalCenter: parent.horizontalCenter
				placeholderText: "Vehicle Year"
				width: parent.width
				anchors.left: parent.left
			}
		}
	}
}
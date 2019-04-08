function sleep(milliseconds) {
  var start = new Date().getTime();
  for (var i = 0; i < 1e7; i++) {
    if ((new Date().getTime() - start) > milliseconds){
      break;
    }
  }
}

WorkerScript.onMessage = function(msg) {
    // ... long-running operations and calculations are done here
	var i;
	for (i = 0; i < msg.lats.length; i++) {
		sleep(500)
		if (i == msg.lats.length - 1) {
			WorkerScript.sendMessage({ 'lat': msg.lats[i], 'long': msg.longs[i], 'last': 'yes', 'driveType': msg.driveType, 'index': msg.index, 'rideID': msg.rideID, 'user': msg.user, 'org': msg.org })
		}
		else {
			WorkerScript.sendMessage({ 'lat': msg.lats[i], 'long': msg.longs[i], 'last': 'no', 'driveType': msg.driveType, 'index': msg.index, 'rideID': msg.rideID, 'user': msg.user, 'org': msg.org  })
		}
	}
}
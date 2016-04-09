package cl
// === This is a list of response-headers that the system uses to communicate
// Commands from master 

var Stop string = "STOP"
var Up string = "UP"
var Down string = "DOWN"

//Commands from slave 
var InnerOrder string = "INNER"
var OuterOrder string = "OUTER"
var Floor string = "FLOOR"
var DoorClosed string = "DOORCLOSED"


// === System status
var Master string = "MASTER"
var Slave string = "SLAVE"
var Ping string = "PING"

// Commands from slave 
var SetMaster string = "SETMASTER"
var Startup string = "STARTUP"
var Timeout string = "TIMEOUT"
var Unknown string = "UNKNOWN"

// Commands from master
var JoinMaster string = "JOIN"
var Backup string = "BACKUP"


// Sender-address
var All string = "ALL"

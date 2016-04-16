package cl

// === This is a list of response-headers that the system uses to communicate
// Commands from master

var Stop string = "STOP"
var Up string = "UP"
var Down string = "DOWN"
var LightOffInner = "LOFFINNER"
var LightOffOuterUp = "LOFFOUP"
var LightOffOuterDown = "LOFFODOWN"
var LightOnInner = "LONINNER"
var LightOnOuterUp = "LONOUP"
var LightOnOuterDown = "LONODOWN"

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
var Unknown string = "UNKNOWN"

// Commands from master
var JoinMaster string = "JOIN"
var Backup string = "BACKUP"

// Sender-address
var All string = "ALL"

// Communication package
// Response
var Connection string = "CONNECTION"
// Content
var OK string = "OK"
var Timeout string = "TIMEOUT"
var Sent string = "SENT"
var Failed string = "Shame! Shame! Shame!"

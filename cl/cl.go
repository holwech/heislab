package cl

// === This is a list of response-headers that the system uses to communicate

// === Orders
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
// = Response
var System string = "SYSTEM"

// = Content
// Commands from both
var Master string = "MASTER"
var Slave string = "SLAVE"
var Ping string = "PING"

// Commands from slave
var SetMaster string = "SETMASTER"
var Startup string = "STARTUP"
var Unknown string = "UNKNOWN"
var EngineFail string = "ENGINEFAILED"
var EngineOK string = "ENGINEOK"

// Commands from master
var JoinMaster string = "JOIN"
var Backup string = "BACKUP"
var All string = "ALL"

// === Communication package
// = Response
var Connection string = "CONNECTION"

// = Content
var OK string = "OK"
var Timeout string = "TIMEOUT"
var Sent string = "SENT"
var Failed string = "FAILED"

// === Ports
var SReadPort string = ":25101"
var SWritePort string = ":25010"

var MReadPort string = ":25010"
var MWritePort string = ":25101"

var MtoMPort string = ":26010"

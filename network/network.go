package network

type Message struct {
	Sender, Receiver, ID, Response, Content string
}

func InitNetwork() (<-chan Message, <-chan Message){
	
}
	//messageChan, statusChan := InitNetwork()
	//sendChan := network get send chan
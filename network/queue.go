package network

type OrderList struct {
	list []Message
}

func (ol *OrderList) Next() (Message) {
	var message Message
	if (len(ol.list) > 0) {
		message = ol.list[0]
		ol.remove(0)
		return message
	} else {
		return false
	}
}

func (ol *OrderList) Add(order Message) {
	append(ol.list, order)
}

func (ol *OrderList) Done(id string) {
	for val, i := range ol.list {
		if val[i].ID == id {
			ol.remove(i)
		}
	}
}

func (ol *OrderList) remove(i int) {
	copy(ol.list[i:], ol.list[i + 1])
	ol.list[len(ol.list) - 1] = nil
	ol.list = ol.list[:len(arr) - 1]
}

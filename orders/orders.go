package orders

/*WANT: 
Possibility to add new orders

handle orders:
-needs to chose which elevator is going to handle an order
-needs to know which elevator is close and idle or going same direction
-elevator states should be kept in elevator package?
-Do we even need an elevator package? Part of slave?
-pass elevator states to order handling?
-in that case, should be read only

remove orders when finished

*/

struct Orders{
	InnerOrders = 
	OuterOrdersUp
	OuterOrdersDown
}
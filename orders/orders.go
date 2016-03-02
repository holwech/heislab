package orders

/*WANT: 


handle orders:
-needs to chose which elevator is going to handle an order
-needs to know which elevator is close and idle or going same direction
-elevator states should be kept in elevator package?
-Do we even need an elevator package? Part of slave?
-pass elevator states to order handling?
-in that case, should be read only


*/

/*
After some thinking-consulting:
orders should be kept in elevator module as inner orders could
be considered part of elevator state
and
outer orders is global
*/
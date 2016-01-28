// Wrapper for libComedi Elevator control.
// These functions provide an interface to the elevators in the real time lab


// Number of floors. Hardware-dependent, do not modify.
#define N_FLOORS 4

// Number of buttons (and corresponding lamps) on a per-floor basis
#define N_BUTTONS 3

void elev_init(void);

void elev_set_motor_direction(int dirn);
void elev_set_button_lamp(int button, int floor, int value);
void elev_set_floor_indicator(int floor);
void elev_set_door_open_lamp(int value);
void elev_set_stop_lamp(int value);

int elev_get_button_signal(int button, int floor);
int elev_get_floor_sensor_signal(void);
int elev_get_stop_signal(void);
int elev_get_obstruction_signal(void);





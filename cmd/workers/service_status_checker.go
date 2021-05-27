package main

// We need to have a worker ticks every minute and checks routinely on the value of `service_status` for each service record.
// We have to check for the following senarios and change the `service_status` to proper status.
//
// Expired:
//    If current time is greater than the `start_time + buffer time`, we consider the service has expired. Hence,
//    we should set the service status to `expired`.
//
// Completed:
//    If service status is `fulfilling` and current time is greater or equal to `end_time`, we will set the
//    `service_status` of the service to be `completed`.

func main() {
	//ticker := time.NewTicker(1 * time.Minute)

}

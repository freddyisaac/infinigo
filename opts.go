package main


import "flag"

var local_ep *string = flag.String("l", ":8999", "endpoint")
var remote_ep *string = flag.String("d", ":8999", "destination endpoint")


var bbuffer_size      *uint = flag.Uint("bs", 2048, "buffer memory size")


var initiator_flag *bool = flag.Bool("ia", false, "initiator/acceptor flag default acceptor")

func IsInitiator() bool {
	return *initiator_flag
}

func IsAcceptor() bool {
	return ! (*initiator_flag)
}

var if_name *string = flag.String("i", "mthca0" ,"use alternative interface name")

func GetIfName() string {
	return *if_name
}


var use_json *bool = flag.Bool("j", false, "use json for communicating mthca0 details")

func GetModeName() string {
	if *initiator_flag == true {
		return "initiator"
	}
	return "acceptor"
}

func UsingJson() bool {
	return *use_json
}


package main


/*
#cgo CFLAGS: -I.
#cgo LDFLAGS: -L. -lca -libverbs
#include <stdio.h>
#include <stdlib.h>
#include <stdint.h>
#include "ca.h"
*/
import "C"

//import "errors"
import "fmt"
import "unsafe"
import "flag"
import "os"
import "os/signal"
//import "time"


const (
	TEST_RECV_WRID = 1
	TEST_SEND_WRID = 2
	BUF_SIZE = 32 * 1024 // give ourselves a 32k buffer
)

func CreateAcceptorListener(data chan *CIBVAddress) {
	ready := make(chan bool, 1)
	notify := make(chan bool, 16)
	if UsingJson() {
// deprecate
		go RouteSetup(ready)
	}else{
		go ibv_listen(ready, *local_ep, notify, data)
	}
	<- ready
}

func main() {
	var err error

	flag.Parse() // parse command line
	fmt.Printf("Starting as %s...\n", GetModeName())

// Create an acceptor endpoint for incoming requests
	acceptor_data_notify := make(chan *CIBVAddress, 16)
	CreateAcceptorListener(acceptor_data_notify)

// HCA init
	C.go_ibv_init();


// we are an initator and so we 
// create a handle for our adapter

	initiator := &Initiator{}

	if IsInitiator() {
// this sets up all ibv calls for our
// named infiniband card
		initiator.Setup( "mthca0", BUF_SIZE )

/*
		ibvH, err := CreateIBVHandle( "mthca0", BUF_SIZE )
		if err != nil {
			fmt.Printf("CreateIBVHandle error : %+v\n" ,err)
		}

// get a C ibvaddress object and 
		c_ibv_addr, err := ibvH.GetCIBVAddressGenNewPsn()

		initiator_uuid := CreateMapEntry(ibvH, c_ibv_addr)

		GetInitiatorRegionMap().Dump("INITIATOR")

// now we can notify our endpoint of our ibv endpoint config

		local_ep_sa := c_ibv_addr.ToStringArray()
		local_ep_sa = local_ep_sa
		err = ibv_notify(*remote_ep, initiator.GetUUID())
*/
		err = initiator.NotifyIBV(*remote_ep)
		if err != nil {
			fmt.Printf("ibv_notify() error : %+v\n", err)
		}

		GetInitiatorRegionMap().Dump("INITIATOR")

// OUR SETUP IS NOW COMPLETE
// FINALLY OUR TRANSFER OF DATA

fmt.Println("Initator sending data!")
		var tpl *TestPayload
		tpl = (*TestPayload) (initiator.GetIBVHandle().cbuf)
		size := CreateSampleTestPayload( tpl )
/*
		tpl.I = 1
		tpl.I32 = 2
		tpl.I64 = -1
		copy(tpl.Bytes[:], "abcdefghi")
		tpl.F32 = 0.00746
		tpl.F64 = 10.998576	
		size := unsafe.Sizeof( *tpl )
*/
		fmt.Printf("Sending : %v\n", *tpl)
// new version

		fmt.Printf("Post sending %d bytes\n", int(size))
		err = IBVPostSend(initiator.GetIBVHandle(), TEST_SEND_WRID, int(size))
		if err != nil {
			fmt.Printf("IBVPost failure error : %+v\n", err)
		}
// we need to poll our ibv interface
// to notify when send is complete
		var done chan error
		IBVPollInitiatorEvent(done, initiator.GetUUID(), 1)
		err = <- done
		fmt.Printf("Initiator Event error : %+v\n", err)

	}else{

fmt.Printf("Acceptor processing\n")
		c_ibv_address := <- acceptor_data_notify
		if c_ibv_address != nil {
			sr := &SimpleReceiver{}
			IBVPollAcceptorEventWithReceiver(c_ibv_address.uuid, 1, sr)
		}
	}

/*

	if *initiator_flag == false {
		var tpl *TestPayload
		tpl = (*TestPayload) (ibvH.cbuf)
		fmt.Printf("Received : %v\n", *tpl)
	}

// Must cleanup all of our c mallocs
	c_rval = C.go_ibv_close_device(c_op)
	fmt.Printf("ibv_close_device() : %d\n", int(c_rval))
*/

//
// just wait here for a interrupt
// to kill our process

	ic := make(chan os.Signal)
	signal.Notify(ic, os.Interrupt, os.Kill)
	<- ic
	fmt.Printf("Fin!\n")
}



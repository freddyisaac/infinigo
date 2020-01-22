package main

/*
#cgo CFLAGS: -I.
#include <stdint.h>
#include "ca.h"
*/
import "C"

import "errors"
import "fmt"
import "strconv"
import "unsafe"

const (
    NON_RDMA = 0
    IS_RDMA = 1 
)


type IBVHandle struct {
	c_op C.OPAQUE
	c_devnum C.int
	devnames []string
	devindex int
	devnum int
	devname string
	cbuf unsafe.Pointer
}

func (ibvh *IBVHandle) GetCIBVAddress() (*CIBVAddress, error) {
	var err error
	var C_ibv_addr C.IBVAddress
	c_rval := C.go_ibv_get_ibv_address_non_rdma_nogen(ibvh.c_op, &C_ibv_addr)
	rval := int(c_rval)
	if rval == -1 {
		serr := C.GoString( C.go_ibv_get_error(ibvh.c_op) )
		err = errors.New( serr )
		fmt.Printf("Error : %+v\n", err)
		return nil, err
	}
	c_ibv_addr := &CIBVAddress{}
	c_ibv_addr.FromC(&C_ibv_addr)
	fmt.Printf("Addr : %v\n", c_ibv_addr)
	return c_ibv_addr, nil
}

func (ibvh *IBVHandle) GetCIBVAddressGenNewPsn() (*CIBVAddress, error) {
	var err error
	var C_ibv_addr C.IBVAddress
	c_rval := C.go_ibv_get_ibv_address_non_rdma(ibvh.c_op, &C_ibv_addr)
	rval := int(c_rval)
	if rval == -1 {
		serr := C.GoString( C.go_ibv_get_error(ibvh.c_op) )
		err = errors.New( serr )
		fmt.Printf("Error : %+v\n", err)
		return nil, err
	}
	c_ibv_addr := &CIBVAddress{}
	c_ibv_addr.FromC(&C_ibv_addr)
	fmt.Printf("Addr : %v\n", c_ibv_addr)
	return c_ibv_addr, nil
}


var cibvLabels  = []string {
				"Lid",
				"Qpn",
				"Psn",
				"Raddr",
				"Rkey",
				"Flag",
}

type CIBVAddress struct {
	Lid uint16
	Qpn uint32
	Psn uint32
	Raddr uint64
	Rkey uint32
	Flag uint32
	SubnetPrefix uint64
	InterfaceId uint64

	uuid string // used in communicating between initiator/acceptor
}

func (o *CIBVAddress) Dump() string {
	strVal := o.ToStringArray()
	var val string
	for i, v := range cibvLabels {
		val += v + " : " + (*strVal)[i+2] + " "
	}
	return val
}

// should this use version and rdma type
func (o *CIBVAddress) FromStringArray(sa StringArray) error {
	var err error
	var val int
	if val, err = strconv.Atoi(sa[0]); err != nil {
		return err
	}
	o.Lid = uint16(val)
	if val, err = strconv.Atoi(sa[1]); err != nil {
		return err
	}
	o.Qpn = uint32(val)
	if val, err = strconv.Atoi(sa[2]); err != nil {
		return err
	}
	o.Psn = uint32(val)
	if val, err = strconv.Atoi(sa[3]); err != nil {
		return err
	}
	o.Raddr = uint64(val)
	if val, err = strconv.Atoi(sa[4]); err != nil {
		return err
	}
	o.Rkey = uint32(val)
	if val, err = strconv.Atoi(sa[5]); err != nil {
		return err
	}
	o.Flag = uint32(val)
	o.uuid = sa[6]
	return nil
}

func (o *CIBVAddress) ToStringArray() *StringArray {
	sa := make(StringArray, 9)
	sa[0] = "nonrdma" // communication type
	sa[1] = "1" // version
	sa[2] = strconv.Itoa( int(o.Lid) )
	sa[3] = strconv.Itoa( int(o.Qpn) )
	sa[4] = strconv.Itoa( int(o.Psn) )
	sa[5] = strconv.Itoa( int(o.Raddr) )
 	sa[6] = strconv.Itoa( int(o.Rkey) )
	sa[7] = strconv.Itoa( int(o.Flag) )
	sa[8] = o.uuid
	return &sa
}

// this is always called by the acceptor
func (o *CIBVAddress) Process(data chan *CIBVAddress) (BufferProvider, error) {
	localMapEntry := GetAcceptorRegionMap().GetLocalMapEntry( o.uuid )
fmt.Printf("Process : %s -> %+v\n", o.uuid, localMapEntry)
//	remoteMapEntry := GetAcceptorRegionMap().GetRemoteMapEntry( localMapEntry.uuid )
// lets get our local endpoint
	cibvAddress, _ := localMapEntry.ibvHandle.GetCIBVAddressGenNewPsn()
println("===")
fmt.Printf("Processing : %s\n", o.Dump())
fmt.Printf("Local : %+v\n", cibvAddress.Dump())
println("===")

GetAcceptorRegionMap().Dump("ACCEPTOR")

	err := AssocRemoteEndpoint( localMapEntry.ibvHandle, o)
	if err != nil {
		data <- nil
		return nil, err
	}
	
// accept initiator queue data
	data <- o
// return acceptor queue data for reply to initiator
	return cibvAddress.ToStringArray(), nil
}

func (o *CIBVAddress) String() string {
	return fmt.Sprintf("lid: %#x qpn: %#x psn: %#x", o.Lid, o.Qpn, o.Psn)
}

func (o *CIBVAddress) FromC(c_ibva *C.IBVAddress) {
	o.Lid = uint16(c_ibva.lid)
	o.Qpn = uint32(c_ibva.qpn)
	o.Psn = uint32(c_ibva.psn)
	o.Raddr = uint64(c_ibva.raddr)
	o.Rkey = uint32(c_ibva.rkey)
}

func (o *CIBVAddress) ToC() *C.IBVAddress {
	c_ibva := &C.IBVAddress{}
	c_ibva.lid = C.uint16(o.Lid)
	c_ibva.qpn = C.uint32(o.Qpn)
	c_ibva.psn = C.uint32(o.Psn)
	c_ibva.raddr = C.uint64(o.Raddr)
	c_ibva.rkey = C.uint32(o.Rkey)
	return c_ibva
}

func IBVPostSend(ibvH *IBVHandle, id, size int) error {
	c_rval := C.go_ibv_post_send(ibvH.c_op, C.int(id), C.int(size))
	rval := int(c_rval)
	if rval == -1 {
		serr := C.GoString( C.go_ibv_get_error(ibvH.c_op) )
		err := errors.New( serr )
		fmt.Printf("Error : %+v\n", err)
		return err
	}
	return nil
}

func IBVPollCQEvent(done chan error, ibvH *IBVHandle, val int) {
	c_rval := C.go_ibv_poll_cq_event(ibvH.c_op, C.int(val)); // we are going to use an event mechanism
	rval := int(c_rval)
	if rval == -1 {
		serr := C.GoString( C.go_ibv_get_error(ibvH.c_op) )
		err := errors.New( serr )
		fmt.Printf("Error : %+v\n", err)
		done <- err	
		return
    }
	done <- nil
}

func IBVPollInitiatorEvent(done chan error, uuid string, val int) {
	localMapEntry := GetInitiatorRegionMap().GetLocalMapEntry( uuid )
	IBVPollCQEvent(done, localMapEntry.ibvHandle, val)
}

func IBVPollAcceptorEvent(done chan error, uuid string, val int) {
	localMapEntry := GetAcceptorRegionMap().GetLocalMapEntry( uuid )
	IBVPollCQEvent(done, localMapEntry.ibvHandle, val)
}

type ReceiverFunc func(unsafe.Pointer) error

type ReceiveProvider interface {
	Process(unsafe.Pointer) error
}

func IBVPollAcceptorEventWithReceiver(uuid string, val int, rp ReceiveProvider) error {
	done := make(chan error, 1)
	localMapEntry := GetAcceptorRegionMap().GetLocalMapEntry( uuid )
	IBVPollCQEvent(done, localMapEntry.ibvHandle, val)
	err := <- done
	if err != nil {
		return err	
	}
	err = rp.Process( localMapEntry.ibvHandle.cbuf )
	return err
}

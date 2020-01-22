package main

/*
#cgo CFLAGS: -I.
#include <stdint.h>
#include "ca.h"
*/
import "C"


import "fmt"
import "errors"


// everything we need to create a local endpoint for send/receive messages

func CreateIBVHandle(card_name string, size uint32) (*IBVHandle, error) {
    var err error
    ibvh := &IBVHandle{}
    ibvh.c_op = C.go_ibv_get_device_list()
    ibvh.c_devnum = C.go_ibv_get_device_list_num(ibvh.c_op)
    totdevnum := int(ibvh.c_devnum)
    fmt.Printf("Have %d devices\n", int(ibvh.c_devnum))
    var devname string
    var devnum int
    ibvh.devnames = make([]string, totdevnum)
    for i:=0;i<totdevnum;i++ {
        c_devname := C.go_ibv_get_device_name(ibvh.c_op, C.int(i));
        devname = C.GoString(c_devname);
        devnum = i
        ibvh.devnames[i] = devname
        if devname == card_name {
// set our device
            ibvh.devindex = i
            C.go_ibv_set_device_index(ibvh.c_op,C.int(i))
            break
        }
    }
    ibvh.devnum = devnum
    ibvh.devname = devname
    fmt.Printf("device name : <%d,%s>\n", devnum,devname)

// open a device
    c_rval := C.go_ibv_open_device(ibvh.c_op)
    rval := int(c_rval)
    if rval == -1 {
        serr := C.GoString( C.go_ibv_get_error(ibvh.c_op) )
        err = errors.New( serr )
        fmt.Printf("Error : %+v\n", err)
        return nil, err
    }
    fmt.Printf("ibv_open_device() : %d\n", rval)

// create protection domain
    c_rval = C.go_ibv_alloc_pd(ibvh.c_op);
    rval = int(c_rval)
    if rval == -1 {
        serr := C.GoString( C.go_ibv_get_error(ibvh.c_op) )
        err = errors.New( serr )
        fmt.Printf("Error : %+v\n", err)
        return nil, err
    }
    fmt.Printf("ibv_alloc_pd() : %d\n", rval)

// allocate a buffer
//  ibvh.cbuf = C.go_ibv_alloc_buffer(ibvh.c_op, C.int(BUF_SIZE * 1024))
    ibvh.cbuf = C.go_ibv_alloc_buffer(ibvh.c_op, C.int( size ))
    fmt.Printf("ibv_alloc_buffer() : %p %T\n", ibvh.cbuf, ibvh.cbuf)

// register memory region
    c_rval = C.go_ibv_reg_mr(ibvh.c_op)
    rval = int(c_rval)
    if rval == -1 {
        serr := C.GoString( C.go_ibv_get_error(ibvh.c_op) )
        err = errors.New( serr )
        fmt.Printf("Error : %+v\n", err)
        return nil, err
    }
    fmt.Printf("ibv_reg_mr() : %d\n", rval)

// create completion channel

    c_rval = C.go_ibv_create_comp_channel(ibvh.c_op)
    rval = int(c_rval)
    if rval == -1 {
        serr := C.GoString( C.go_ibv_get_error(ibvh.c_op) )
        err = errors.New( serr )
        fmt.Printf("Error : %+v\n", err)
        return nil, err
    }
    fmt.Printf("go_ibv_create_comp_channel() : %d\n", rval)

// create completion queues

    c_rval = C.go_ibv_create_cq(ibvh.c_op,C.int(20),C.int(2))
    rval = int(c_rval)
    if rval == -1 {
        serr := C.GoString( C.go_ibv_get_error(ibvh.c_op) )
        err = errors.New( serr )
        fmt.Printf("Error : %+v\n", serr)
        return nil, err
    }
    fmt.Printf("go_ibv_create_cq() : %d\n", int(c_rval))

// create queue pair
    c_rval = C.go_ibv_simple_create_qp(ibvh.c_op)
    rval = int(c_rval)
    if rval == -1 {
        serr := C.GoString( C.go_ibv_get_error(ibvh.c_op) )
        err = errors.New( serr )
        fmt.Printf("Error : %+v\n", err)
        return nil, err
    }
    fmt.Printf("go_ibv_simple_create_qp() : %d\n", rval)

// init our queue pair
    c_rval = C.go_ibv_init_qp(ibvh.c_op,C.uchar(1)) // param is port number
    rval = int(c_rval)
    if rval == -1 {
        serr := C.GoString( C.go_ibv_get_error(ibvh.c_op) )
        err = errors.New( serr )
        fmt.Printf("Error : %+v\n", err)
        return nil, err
    }
    fmt.Printf("go_ibv_init_qp() : %d\n", rval)

// post entries to the recv queue

    c_rval = C.go_ibv_post_recv(ibvh.c_op, C.int(TEST_RECV_WRID)); 
    rval = int(c_rval)
    if rval == -1 {
        serr := C.GoString( C.go_ibv_get_error(ibvh.c_op) )
        err = errors.New( serr )
        fmt.Printf("Error : %+v\n", err)
        return nil, err
    }
    fmt.Printf("go_ibv_post_recv() : %d\n", rval)

// if using events we want to request our eents

    c_rval = C.go_ibv_req_notify_cq(ibvh.c_op)
    rval = int(c_rval)
    rval = int(c_rval)
    if rval == -1 {
        serr := C.GoString( C.go_ibv_get_error(ibvh.c_op) )
        err = errors.New( serr )
        fmt.Printf("Error : %+v\n", err)
        return nil, err
    }
    fmt.Printf("go_ibv_req_notify_cq() : %d\n", rval)


// query our port info
    c_rval = C.go_ibv_query_port(ibvh.c_op,C.uchar(1))
    rval = int(c_rval)
    if rval == -1 {
        serr := C.GoString( C.go_ibv_get_error(ibvh.c_op) )
        err = errors.New( serr )
        fmt.Printf("Error : %+v\n", err)
        return nil, err
    }
    fmt.Printf("go_ibv_query_port() : %d\n", rval)
    return ibvh, nil
}


// associate a remote endpoint with our local endpoint

func AssocRemoteEndpoint(ibvH *IBVHandle, cibvAddr *CIBVAddress) error {
	c_rval := C.go_ibv_modify_qp_remote_ep(ibvH.c_op,
					C.uint32(cibvAddr.Qpn),
					C.uint32(cibvAddr.Psn),
					C.uint16(cibvAddr.Lid))
	rval := int(c_rval)
	if rval == -1 {
		serr := C.GoString( C.go_ibv_get_error(ibvH.c_op) )
		err := fmt.Errorf( serr )
		return err
	}
	return nil
}

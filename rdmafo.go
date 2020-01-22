package main

import (
	"fmt"
	"time"
)


type RDMARepoInfo struct {
	LocalId		string
	RemoteId	string	
	Size		uint32
	data interface{} // TBD ??
}

func (o *RDMARepoInfo) String() string {
	return fmt.Sprintf("Local: %s Remote: %s Size: %d\n", o.LocalId, o.RemoteId, o.Size)
}


// for now lets maintain two maps
// one for local lookups
// one for remote lookups
var remote_rdma_repo	map[string]*RDMARepoInfo
var local_rdma_repo	map[string]*RDMARepoInfo

func RemoteRDMARepoLookup(uid string) *RDMARepoInfo {
	if v,ok := remote_rdma_repo[uid]; ok == true {
		return v
	}
	return nil
}

func LocalRDMARepoLookup(uid string) *RDMARepoInfo {
	if v,ok := local_rdma_repo[uid]; ok == true {
		return v
	}
	return nil
}

func RemoteRDMARepoNewEntry(uid string, rri *RDMARepoInfo) {
	remote_rdma_repo[uid] = rri
}

func LocalRDMARepoNewEntry(uid string, rri *RDMARepoInfo) {
	local_rdma_repo[uid] = rri
}



//
// a few simple aux routines
//

type UUID   [16]byte
var xorstate uint32 = uint32(time.Now().UnixNano())

func xorshift32() uint32 {
	x := xorstate
	x = x ^ (x << 13)
	x = x ^ (x >> 17)
	x = x ^ (x << 5)
	xorstate = x
	return x
}

func xorshift32b4(x uint32,b []byte) {
	b[0] = byte(x & 0x000000FF)
	b[1] = byte(x & 0x0000FF00 >> 8)
	b[2] = byte(x & 0x00FF0000 >> 16)
	b[3] = byte(x & 0xFF000000 >> 24)
}

func RandomUUID() (string,error) {
	var uuid UUID
	x := xorshift32()
	xorshift32b4(x, uuid[0:4])
	x = xorshift32()
	xorshift32b4(x, uuid[4:8])
	x = xorshift32()
	xorshift32b4(x, uuid[8:12])
	x = xorshift32()
	xorshift32b4(x, uuid[12:])
	uuid[8] = uuid[8] &^ 0xC0 | 0x80
	uuid[6] = uuid[6] &^ 0xF0 | 0x40
	return fmt.Sprintf("%x-%x-%x-%x-%x", uuid[0:4],uuid[4:6],uuid[6:8],uuid[8:10],uuid[10:]),nil
}

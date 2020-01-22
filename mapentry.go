package main

/*
#cgo CFLAGS: -I.
#include <stdint.h>
#include "ca.h"
*/
import  "C"

import "fmt"


type MapEntry struct {
	uuid string
	size uint32
//	c_op C.OPAQUE
	ibvHandle *IBVHandle
}

func (me *MapEntry) String() string {
	cibvHandle, _ := me.ibvHandle.GetCIBVAddress()
	return fmt.Sprintf("uuid : %s size %d handle : %v\n", me.uuid, me.size, cibvHandle.Dump())
}

type RegionMaps struct {
	remote map[string]*MapEntry
	local map[string]*MapEntry
}

func (rm *RegionMaps) Init() {
	rm.remote = make(map[string]*MapEntry)
	rm.local = make(map[string]*MapEntry)
}

func (rm *RegionMaps) Dump(tag string) {
	fmt.Printf("=== %s ===\n", tag)
	defer fmt.Printf("=== %s ===\n", tag)

	fmt.Printf("Local : \n")
	for k,v := range rm.local {
		fmt.Printf("key : %s entry : {%s}\n", k, v)
	}
	fmt.Printf("Remote : \n")
	for k,v := range rm.remote {
		fmt.Printf("key : %s entry : {%s}\n", k, v)
	}
}

var global_map *RegionMaps

// mostly for debugging 
type NamedRegionMaps struct {
	namedRegions map[string]*RegionMaps
}


var global_named_map NamedRegionMaps

func GetGlobalRegionMaps() *RegionMaps {
	return global_map
}

func GetNamedRegionMaps(name string) *RegionMaps {
	return global_named_map.namedRegions[name]
}

func GetAcceptorRegionMap() *RegionMaps {
	return global_named_map.namedRegions["acceptor"]
}

func GetInitiatorRegionMap() *RegionMaps {
	return global_named_map.namedRegions["initiator"]
}

func init() {
	global_named_map.namedRegions = make(map[string]*RegionMaps)
	acceptor_map := new(RegionMaps)
	acceptor_map.Init()
	global_named_map.namedRegions["acceptor"] = acceptor_map
	initiator_map := new(RegionMaps)
	initiator_map.Init()
	global_named_map.namedRegions["initiator"] = initiator_map
	global_map = new(RegionMaps)
	global_map.Init()
}

func (rm *RegionMaps) AddMapEntry(local, remote string, size uint32) (*MapEntry,*MapEntry) {
// should check for existence
	localme := &MapEntry{}
	localme.uuid = local
	localme.size = size
	rm.remote[remote] = localme
	remoteme := &MapEntry{}
	remoteme.uuid = remote
	remoteme.size = size
	rm.local[local] = remoteme
	return localme, remoteme
}

func (rm *RegionMaps) GetLocalMapEntry(id string) *MapEntry {
	return rm.local[id]
}

func (rm *RegionMaps) GetRemoteMapEntry(id string) *MapEntry {
	return rm.remote[id]
}


// cleanup number of return parameters
func CreateMapEntry(ibvH *IBVHandle, c_ibv_addr *CIBVAddress) string {
    initiator_uuid, _ := RandomUUID()
    ibv_request_region(*remote_ep, initiator_uuid, BUF_SIZE)
// it is probably overkill to keep both remote and local maps but until I work this code
// a bit cleaner continue to trim redundancy
    localMapEntry := GetInitiatorRegionMap().GetLocalMapEntry(initiator_uuid)
    remoteMapEntry := GetInitiatorRegionMap().GetRemoteMapEntry( localMapEntry.uuid )
    localMapEntry.ibvHandle = ibvH
    remoteMapEntry.ibvHandle = ibvH
    c_ibv_addr.uuid = localMapEntry.uuid
    return initiator_uuid
}


package main

import (
	"fmt"
	"net"
	"time"
	"strings"
	"encoding/binary"
	"encoding/hex"
	"strconv"
	"log"
)

var _ = log.Printf

type BufferProvider interface {
	Marshal() []byte
	Unmarshal([]byte)
}

type StringArray []string

func (sa StringArray) Marshal() []byte {
	rval := strings.Join(sa, "|")
	return []byte(rval)	
}

func (sa *StringArray) Unmarshal(b []byte) {
	splitStr := strings.Split(string(b), "|")
	*sa = splitStr
}


type HandlerProvider interface {
	ToStringArray() *StringArray
	FromStringArray(StringArray) error
	Process(chan *CIBVAddress) (BufferProvider, error)
}

type SizeProtoProvider struct {
	uuid string
	size uint32 // should limit this ?
}

func (sp *SizeProtoProvider) ToStringArray() *StringArray {
	sa := make(StringArray, 4)
	sa[0] = "size"
	sa[1] = "1"
	sa[2] = sp.uuid
	sa[3] = strconv.Itoa( int( sp.size ) )
	return &sa
}

func (sp *SizeProtoProvider) FromStringArray(sa StringArray) error {
	sp.uuid = sa[0]
	val, err := strconv.Atoi(sa[1])
	if err != nil {
		return err
	}
	sp.size = uint32( val )
	return nil
}

// only call Process if we are an Acceptor
func (sp *SizeProtoProvider) Process(data chan *CIBVAddress) (BufferProvider, error) {

// we have a uuid from our reomte side and a size request
// create a local uuid and then create a map entry for this remote side request
	uuid, _ := RandomUUID()
	local, remote := GetAcceptorRegionMap().AddMapEntry(uuid, sp.uuid, sp.size)	
	ibvHandle, err := CreateIBVHandle(GetIfName(), sp.size)
	if err != nil {
		fmt.Printf("CreateIBVHandle error : %+v\n", err)
	}
	local.ibvHandle = ibvHandle
	remote.ibvHandle = ibvHandle
// we should create an IBV handle here
	reply := &SizeProtoProvider{}
	reply.uuid = uuid
	reply.size = sp.size
	sa := reply.ToStringArray()
fmt.Printf("replying : %+v\n", sa)
	return reply.ToStringArray(), nil
}

func ResolveHandler(name string) HandlerProvider {
	switch name {
		case "nonrdma" :
			return new(CIBVAddress)
		case "size" :
			return new(SizeProtoProvider)
	}
	return nil
}

// handler routines for communicating
// ibv parameters via a plain socket

func handleIBVEP(conn net.Conn, notify chan bool, data chan *CIBVAddress) error {
	defer conn.Close()
	lenBuf := make([]byte, 4)
	var ulen uint32
	n, err := conn.Read(lenBuf)
	if err != nil {
		return err
	}
	ulen = binary.BigEndian.Uint32(lenBuf)
	fmt.Printf("length : %d\n", ulen)
	packetBuf := make([]byte, ulen)
	n, err = conn.Read(packetBuf)	
	if err != nil {
		return err
	}
	if uint32(n) != ulen {
		return fmt.Errorf("incorrect buffer read : %d of %d\n", n, lenBuf)
	}
	fmt.Printf("read : %s\n", hex.Dump( packetBuf ) )
	strBuf := string(packetBuf)
	vals := strings.Split(strBuf, "|")
	fmt.Printf("vals : %+v\n", vals)
	handler := ResolveHandler(vals[0])
	handler.FromStringArray( StringArray(vals[2:]) )
	reply, err := handler.Process(data)
	if err != nil {
fmt.Printf("Process() error : %+v\n", err)
		return err
	}
	reply_buf := reply.Marshal()
	writePacket(conn, reply_buf)
	return nil
}


func ibv_listen(ready chan bool, ep string, notify chan bool, data chan *CIBVAddress) error {
	ln, err := net.Listen("tcp4", ep)
	if err != nil {
		return err
	}
	count := 0
	ready <- true
	for {
		conn, err := ln.Accept()
		if err != nil {
// just ignore
		}
		count++
		log.Printf("Processing client : %d\n", count)
		go handleIBVEP(conn, notify, data)
	} 
}

// this is a fire and forget protocol
// no reply is returned
type GetCIBVAddressFunc func() *CIBVAddress

func nilFunc() *CIBVAddress {
	return nil
}

// noify always is from an initiator
func ibv_notify(host string, uuid string) error {
	localMapEntry := GetInitiatorRegionMap().GetLocalMapEntry( uuid )
	c_ibv_addr, err := localMapEntry.ibvHandle.GetCIBVAddress()
	c_ibv_addr.uuid = localMapEntry.uuid
	sa := c_ibv_addr.ToStringArray()
//func ibv_notify(host string, sa *StringArray) (GetCIBVAddressFunc, error) {
	conn, err := net.DialTimeout("tcp4", host, 4 * time.Second)
	defer conn.Close()
	if err != nil {
		return err
	}
	buf := make([]byte, 4)
	val := sa.Marshal()
	binary.BigEndian.PutUint32(buf, uint32(len(val)))
	buf = append(buf, val...)
	n,err := conn.Write(buf)
	fmt.Printf("wrote : %d of %d\n%s\n", n, len(buf), hex.Dump(buf))
	packet, err := readPacket(conn)
	fmt.Printf("notify received : %s\n", string(packet))
	cibvAddress := &CIBVAddress{}
	strBuf := string( packet )
	vals := strings.Split(strBuf, "|")
	cibvAddress.FromStringArray( StringArray(vals[2:]) )
	fmt.Printf("notify cibvAddress : %s\n", cibvAddress.Dump())
	AssocRemoteEndpoint( localMapEntry.ibvHandle, cibvAddress )
	return nil
}

func ibv_request_region(host string, uuid string, size uint32) error {
	conn, err := net.DialTimeout("tcp4", host, 4 * time.Second)
	if err != nil {
		return err
	}
	defer conn.Close()
	net_packet := make([]byte, 4)
	sp := &SizeProtoProvider{}
	sp.uuid = uuid
	sp.size = size
	buffer := sp.ToStringArray().Marshal()
	binary.BigEndian.PutUint32(net_packet, uint32(len(buffer)) )
	net_packet = append(net_packet, buffer...)
	_, err = conn.Write( net_packet )
	if err != nil {
		return err
	}
	reply_buffer, err := readPacket(conn)
	if err != nil {
		return err
	}
fmt.Printf("reply : %s\n", hex.Dump(reply_buffer))
	reply_sp := &SizeProtoProvider{}
	sa := &StringArray{}
	sa.Unmarshal( reply_buffer )
	strSlice := (*sa)[2:]
	reply_sp.FromStringArray( StringArray(strSlice) )

// this returns our local and remote maps
// we can ignore this as our invoker can query the map entries for these from the
	GetInitiatorRegionMap().AddMapEntry(uuid, reply_sp.uuid, size)
	return nil		
}

// convert to ByteBuffer and writeFull/readFull
func readPacket(conn net.Conn) ([]byte, error) {
	net_packet := make([]byte, 4)
	_, err := conn.Read(net_packet) // need read fully
	if err != nil {
		return nil, err
	}
	packetLen := binary.BigEndian.Uint32(net_packet)
fmt.Printf("read packet length : %x\n", packetLen)
	net_packet = make([]byte, packetLen)
	_, err = conn.Read(net_packet)
	if err != nil {
		return nil, err
	}
	return net_packet, err
}

func writePacket(conn net.Conn, buf []byte) error {
	net_packet := make([]byte, 4)
	binary.BigEndian.PutUint32(net_packet, uint32(len(buf)) )
	net_packet = append(net_packet, buf...)
fmt.Printf("Replying : %s\n", hex.Dump(net_packet))
	_, err := conn.Write(net_packet)
	return err
}


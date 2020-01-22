package main

// handlers for communicating ibv parameters 
// via json - needs uuid fix

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
//	"os"
//	"os/signal"
//	"time"
	"github.com/gorilla/mux"
)


// initial setup for shared memory

type RDMAInitialRequest struct {
	Id	string	`json:"rdma-uid"`
	Size uint32	`json:"rdma-size"`
}

func ProcessRDMAInitialRequest(w http.ResponseWriter, r *http.Request) {
	info := &RDMAInitialRequest{}
	err := json.NewDecoder(r.Body).Decode(info)
	if err != nil {
		QuickJsonReply(w, 400, "bad request")
		return
	}
	defer r.Body.Close()
	rri := &RDMARepoInfo{}
	rri.Size = info.Size
	rri.RemoteId = info.Id
	local_id,_ := RandomUUID()
	rri.LocalId = local_id
	RemoteRDMARepoNewEntry(rri.RemoteId, rri)
	LocalRDMARepoNewEntry(local_id, rri)
	reply_info := &RDMAInitialRequest{}
	reply_info.Id = local_id
	reply_info.Size = info.Size
	b,err := json.Marshal(reply_info)
	if err != nil {
		QuickJsonReply(w,400,"unable to parse reply!!")
		return
	}
	w.Header().Set("Content-Type", "application/json;charset=utf-8")
	w.WriteHeader(200)
	w.Write(b)
//	QuickJsonReply(w,200,"ok")
}

func NegotiateInitialSize(endpoint string, size uint32) (string,error) {
	rri := &RDMARepoInfo{}
	rir := &RDMAInitialRequest{}
	local_id,_ := RandomUUID()
	rir.Id = local_id // generate a local for this pair
	rir.Size = size

	rri.Size = size
	rri.LocalId = local_id

	req, err := json.Marshal(rir)
	url := "http://" + endpoint + "/api/v1/ibv/endpoint/negotiate"
	req_buf := bytes.NewBuffer(req)
	resp,err := http.Post(url, "application/json", req_buf)
	if err != nil {
		fmt.Printf("Error : %s\n", err.Error())
		return "",nil
	}
	defer resp.Body.Close()
	reply_rir := &RDMAInitialRequest{}
	err = json.NewDecoder(resp.Body).Decode(reply_rir)

	rri.RemoteId = reply_rir.Id
	RemoteRDMARepoNewEntry(rri.RemoteId, rri)
	LocalRDMARepoNewEntry(rri.LocalId, rri)
		
	return local_id,nil
}

type JSONReplyMsg struct {
	Code int	`json:"code"`
	Mesg string	`json:"message"`
}


func (o *JSONReplyMsg) HttpWrite(w http.ResponseWriter) error {
	b,err := json.Marshal(o)
	if err != nil {
		return err
	}
	w.Header().Set("Content-Type", "application/json;charset=utf-8")
	w.WriteHeader(o.Code)
	w.Write(b)
	return nil
}

func QuickJsonReply(w http.ResponseWriter, code int, mesg string) error {
	jr := &JSONReplyMsg{Code: code, Mesg: mesg}
	return jr.HttpWrite(w)
}

type IBVAddressPayload struct {
	Id		string	`json:"uid"`
	Lid		string	`json:"lid"`
	Qpn		string	`json:"qpn"`
	Psn		string	`json:"psn"`
	Raddr	string	`json:"raddr"`
	Rkey	string	`json:"rkey"`
	Flag	uint32		`json:"flag"` // typically for stating rdma or send/recv
	SubnetPrefix	string	`json:"subnet-prefix"`
	InterfaceId		string	`json:"interface-id"`
}

func (o *IBVAddressPayload) FromIBVAddress(ibva *CIBVAddress) {
	o.Lid = fmt.Sprintf("%d", ibva.Lid)
	o.Qpn = fmt.Sprintf("%d", ibva.Qpn)
	o.Psn = fmt.Sprintf("%d", ibva.Psn)
	o.Raddr = fmt.Sprintf("%d", ibva.Raddr)
	o.Rkey = fmt.Sprintf("%d", ibva.Rkey)
	o.Flag = ibva.Flag
	o.SubnetPrefix = fmt.Sprintf("%d\n", ibva.SubnetPrefix)
	o.InterfaceId = fmt.Sprintf("%d\n", ibva.InterfaceId)
}

func (o *IBVAddressPayload) ToIBVAddress() *CIBVAddress {
	rval := &CIBVAddress{}
	var val int
	val,_ = strconv.Atoi(o.Lid)
	rval.Lid = uint16(val)
	val,_ = strconv.Atoi(o.Qpn)
	rval.Qpn = uint32(val)
	val,_ = strconv.Atoi(o.Psn)
	rval.Psn = uint32(val)
	fmt.Sscanf(o.Psn,"%d",rval.Raddr)
//	rval.Raddr = o.Raddr
	val,_ = strconv.Atoi(o.Rkey)
	rval.Rkey = uint32(val)
	rval.Flag = o.Flag
	var val64 uint64
	val64,_ = strconv.ParseUint(o.SubnetPrefix,10,64)
	rval.SubnetPrefix = uint64(val64)
	val64,_ = strconv.ParseUint(o.InterfaceId,10,64)
	rval.InterfaceId = uint64(val64)
	
	return rval	
}

// probably won't use this - easier to just transfer string between C and go as
// this part is not really lataency critical

type IBVAddress struct {
	Lid int16
	Qpn uint32
	Psn uint32
	Raddr uint64
	Rkey uint32
}

func IBVPing(w http.ResponseWriter, r *http.Request) {
	QuickJsonReply(w,200,"ok")
}

// not a great solution - need to fix
// only one region can be associated with this
var config_channel chan *CIBVAddress = make(chan *CIBVAddress,16)

func GetConfigNotifyChannel() chan *CIBVAddress {
	return config_channel
}

func GetSocketChannel() chan *CIBVAddress {
	return config_channel
}

func RecvIbvEndpoint(w http.ResponseWriter, r *http.Request) {
fmt.Printf("RecvIbvEndpoint()\n")
	ibv_addr := &IBVAddressPayload{}	
	err := json.NewDecoder(r.Body).Decode(ibv_addr)
	if err != nil {
		QuickJsonReply(w,400,"bad request : " + err.Error())
		return
	}
	defer r.Body.Close()
	QuickJsonReply(w, 200, "ok")
	reply_val := ibv_addr.ToIBVAddress()
	config_channel <- reply_val
}

// copy of above will depracate
func RecvIbvEndpointNotifyChannel(w http.ResponseWriter, r *http.Request, cfg_ch chan *CIBVAddress) {
fmt.Printf("RecvIbvEndpointNotifyChannel()\n")
	ibv_addr := &IBVAddressPayload{}	
	err := json.NewDecoder(r.Body).Decode(ibv_addr)
	if err != nil {
		QuickJsonReply(w,400,"bad request : " + err.Error())
		return
	}
	defer r.Body.Close()
	QuickJsonReply(w, 200, "ok")
	reply_val := ibv_addr.ToIBVAddress()
	cfg_ch <- reply_val
}

func AttachV1IBVHandlers(v1 *mux.Router) {
	v1.HandleFunc("/ibv/ping", IBVPing).Methods("GET")
	v1.HandleFunc("/ibv/endpoint/push",
		func(w http.ResponseWriter, r *http.Request) {
			RecvIbvEndpointNotifyChannel(w, r, config_channel)
		}).Methods("POST")
	v1.HandleFunc("/ibv/endpoint/negotiate", ProcessRDMAInitialRequest).Methods("POST")
}

func RouteSetup(ready chan bool) error {
	router := mux.NewRouter()
	api_route := router.PathPrefix("/api")
	api_subr := api_route.Subrouter()
	v1ibv := api_subr.PathPrefix("/v1").Subrouter()
	AttachV1IBVHandlers(v1ibv)

	http.Handle("/", router)
	fmt.Printf("Listening on : %s\n", *local_ep)
ready <- true
	err := http.ListenAndServe(*local_ep,nil)
	if err != nil {
		return err
	}
	return nil
}

// this processing receives acks

func IBVShareConfig(endpoint string,ibva *CIBVAddress, flag uint32) error {
	pibva := &IBVAddressPayload{}
	pibva.Flag = flag
	pibva.FromIBVAddress(ibva)

	b,err := json.Marshal(pibva)
	if err != nil {
		fmt.Printf("Error : %s\n", err.Error())
		return nil
	}
	url := "http://" +  endpoint + "/api/v1/ibv/endpoint/push"
	brdr := bytes.NewBuffer(b)
	resp,err := http.Post(url,"application/json", brdr)
	if err != nil {
		fmt.Printf("Error : %s\n", err.Error())
		return err 
	}
	defer resp.Body.Close()
	jrm := &JSONReplyMsg{}
	err = json.NewDecoder(resp.Body).Decode(jrm)
	fmt.Printf("status : %d %v\n", resp.StatusCode, *jrm)
	return err
}




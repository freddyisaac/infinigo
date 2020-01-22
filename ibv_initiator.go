package main


import (
)

type Initiator struct {
	ibvHandle *IBVHandle
	cibvAddress *CIBVAddress
	uuid string
	local_ep_sa *StringArray
}

func (o *Initiator) GetUUID() string {
	return o.uuid
}

func (o *Initiator) GetIBVHandle() *IBVHandle {
	return o.ibvHandle
}

func (o *Initiator) Setup(if_name string, buf_size uint32) error {
	var err error
	o.ibvHandle, err = CreateIBVHandle( if_name, buf_size )
	if err != nil {
		return  err
	}

// get a C ibvaddress object
	o.cibvAddress, err = o.ibvHandle.GetCIBVAddressGenNewPsn()
	if err != nil {
		return err
	}
	o.uuid = CreateMapEntry(o.ibvHandle, o.cibvAddress)

// this is not used - deprecate
	o.local_ep_sa = o.cibvAddress.ToStringArray()

	return nil
}

func (o *Initiator) NotifyIBV(remote_ep string) error {
	err := ibv_notify(remote_ep, o.uuid)
	return err
}


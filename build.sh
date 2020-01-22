#!/bin/sh


time cc -g -c ca.c

ar r libca.a ca.o

time cc -g cat.c ca.c -libverbs

time go build main.go jsonif.go rdmafo.go mdata.go ibv_consts.go sockif.go mapentry.go create_ibv_handle.go ibv_handle.go opts.go ibv_initiator.go




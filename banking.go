package main

import (
	"errors"
	"fmt"
)

/*
#cgo LDFLAGS: -laqbanking
#cgo LDFLAGS: -lgwenhywfar
#cgo darwin CFLAGS: -I/usr/local/include/gwenhywfar4
#cgo darwin CFLAGS: -I/usr/local/include/aqbanking5
#include <aqbanking/banking.h>
*/
import "C"

type AQBankingVersion struct {
	Major      int
	Minor      int
	Patchlevel int
}

type AQBanking struct {
	Name    string
	Version AQBankingVersion

	Ptr *C.AB_BANKING
}

func NewAQBanking(name string) (*AQBanking, error) {
	inst := &AQBanking{}
	inst.Name = name

	inst.Ptr = C.AB_Banking_new(C.CString(inst.Name), nil, 0)
	if err := C.AB_Banking_Init(inst.Ptr); err != 0 {
		return nil, errors.New(fmt.Sprintf("unable to initialized aqbanking: %d", err))
	}
	if err := C.AB_Banking_OnlineInit(inst.Ptr); err != 0 {
		return nil, errors.New(fmt.Sprintf("unable to initialized aqbanking: %d", err))
	}

	inst.loadVersion()

	return inst, nil
}

func (ab *AQBanking) loadVersion() {
	var major, minor, patchlevel, build C.int
	C.AB_Banking_GetVersion(&major, &minor, &patchlevel, &build)
	ab.Version = AQBankingVersion{int(major), int(minor), int(patchlevel)}
}

func (ab *AQBanking) Free() error {
	if err := C.AB_Banking_OnlineFini(ab.Ptr); err != 0 {
		return errors.New(fmt.Sprintf("unable to free aqbanking online: %d\n", err))
	}

	C.AB_Banking_Fini(ab.Ptr)
	C.AB_Banking_free(ab.Ptr)
	return nil
}

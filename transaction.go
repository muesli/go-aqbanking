package main

import (
	"errors"
	"fmt"
	"time"
)

/*
#cgo LDFLAGS: -laqbanking
#cgo LDFLAGS: -lgwenhywfar
#cgo darwin CFLAGS: -I/usr/local/include/gwenhywfar4
#cgo darwin CFLAGS: -I/usr/local/include/aqbanking5
#include <aqbanking/jobgettransactions.h>
#include <aqbanking/banking.h>
#include <aqbanking/banking_ob.h>
*/
import "C"

type Transaction struct {
	// Type               string // AB_Transaction_Type
	// SubType            string // AB_TRANSACTION_SUBTYPE
	Purpose           string
	Text              string
	Status            string
	Date              time.Time
	ValutaDate        time.Time
	MandateReference  string
	CustomerReference string
	Currency          string
	Total             float32
	TransactionPeriod string

	LocalBankCode      string
	LocalAccountNumber string
	LocalIBAN          string
	LocalBIC           string

	RemoteBankCode      string
	RemoteAccountNumber string
	RemoteIBAN          string
	RemoteBIC           string
}

func newTransaction(t *C.AB_TRANSACTION, v *C.AB_VALUE) Transaction {
	transaction := Transaction{}

	transaction.Purpose = (*GwStringList)(C.AB_Transaction_GetPurpose(t)).toString()
	transaction.Text = C.GoString(C.AB_Transaction_GetTransactionText(t))
	transaction.Status = C.GoString(C.AB_Transaction_Status_toString(C.AB_Transaction_GetStatus(t)))
	transaction.MandateReference = C.GoString(C.AB_Transaction_GetMandateReference(t))
	transaction.CustomerReference = C.GoString(C.AB_Transaction_GetMandateReference(t))
	transaction.Date = (*GwTime)(C.AB_Transaction_GetDate(t)).toTime()
	transaction.ValutaDate = (*GwTime)(C.AB_Transaction_GetValutaDate(t)).toTime()

	transaction.Currency = C.GoString(C.AB_Value_GetCurrency(v))
	transaction.Total = float32(C.AB_Value_GetValueAsDouble(v))

	transaction.LocalIBAN = C.GoString(C.AB_Transaction_GetLocalIban(t))
	transaction.LocalBIC = C.GoString(C.AB_Transaction_GetLocalBic(t))
	transaction.LocalBankCode = C.GoString(C.AB_Transaction_GetLocalBankCode(t))
	transaction.LocalAccountNumber = C.GoString(C.AB_Transaction_GetLocalAccountNumber(t))

	transaction.RemoteIBAN = C.GoString(C.AB_Transaction_GetRemoteIban(t))
	transaction.RemoteBIC = C.GoString(C.AB_Transaction_GetRemoteBic(t))
	transaction.RemoteBankCode = C.GoString(C.AB_Transaction_GetRemoteBankCode(t))
	transaction.RemoteAccountNumber = C.GoString(C.AB_Transaction_GetRemoteAccountNumber(t))

	// fmt.Printf("###### '%d'\n", int(C.AB_Transaction_GetStatus(t)))
	// fmt.Printf("###### '%v'\n", C.GoString(C.AB_Transaction_GetLocalCountry(t)))
	// fmt.Printf("###### '%v'\n", C.GoString(C.AB_Transaction_GetLocalSuffix(t)))
	// fmt.Printf("###### '%v'\n", C.GoString(C.AB_Transaction_GetLocalCountry(t)))
	// fmt.Printf("###### '%v'\n", C.GoString(C.AB_Transaction_GetLocalBranchId(t)))

	// transaction.Type = C.GoString(C.AB_Transaction_Type_toString(C.AB_Transaction_GetType(t)))
	// transaction.SubType = C.GoString(C.AB_Transaction_SubType_toString(C.AB_Transaction_GetSubType(t)))

	transaction.TransactionPeriod = C.GoString(C.AB_Transaction_Period_toString(C.AB_Transaction_GetPeriod(t)))

	return transaction
}

// implements AB_JobGetTransactions_new
func (ab *AQBanking) Transactions(acc Account) ([]Transaction, error) {
	fmt.Println("before get transactions")
	var abJob *C.AB_JOB = C.AB_JobGetTransactions_new(acc.Ptr)
	fmt.Println("after get transactions")
	if abJob == nil {
		return nil, errors.New("Unable to load transactions.")
	}

	if err := C.AB_Job_CheckAvailability(abJob); err != 0 {
		return nil, errors.New(fmt.Sprintf("Transactions is not supported by backend: %d", err))
	}

	// TODO set arguments?
	// AB_JobGetTransactions_SetFromTime
	// AB_JobGetTransactions_SetToTime

	var abJobList *C.AB_JOB_LIST2 = C.AB_Job_List2_new()
	C.AB_Job_List2_PushBack(abJobList, abJob)
	var abContext *C.AB_IMEXPORTER_CONTEXT = C.AB_ImExporterContext_new()

	if err := C.AB_Banking_ExecuteJobs(ab.Ptr, abJobList, abContext); err != 0 {
		return nil, errors.New(fmt.Sprintf("Unable to execute Transactions: %d", err))
	}

	var abInfo *C.AB_IMEXPORTER_ACCOUNTINFO = C.AB_ImExporterContext_GetFirstAccountInfo(abContext)
	var transactions []Transaction = make([]Transaction, 0)

	for abInfo != nil {
		var abTransaction *C.AB_TRANSACTION = C.AB_ImExporterAccountInfo_GetFirstTransaction(abInfo)

		for abTransaction != nil {
			var abValue *C.AB_VALUE
			abValue = C.AB_Transaction_GetValue(abTransaction)

			if abValue != nil {
				transactions = append(transactions, newTransaction(abTransaction, abValue))
			}

			abTransaction = C.AB_ImExporterAccountInfo_GetNextTransaction(abInfo)
		}
		abInfo = C.AB_ImExporterContext_GetNextAccountInfo(abContext)
	}

	C.AB_Job_free(abJob)

	return transactions, nil
}

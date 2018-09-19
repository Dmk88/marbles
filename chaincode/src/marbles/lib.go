/*
Licensed to the Apache Software Foundation (ASF) under one
or more contributor license agreements.  See the NOTICE file
distributed with this work for additional information
regarding copyright ownership.  The ASF licenses this file
to you under the Apache License, Version 2.0 (the
"License"); you may not use this file except in compliance
with the License.  You may obtain a copy of the License at

  http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing,
software distributed under the License is distributed on an
"AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY
KIND, either express or implied.  See the License for the
specific language governing permissions and limitations
under the License.
*/

package main

import (
	"encoding/json"
	"errors"
	"net/http"
	"strconv"

	"github.com/stellar/go/clients/horizon"
	hProtocol "github.com/stellar/go/protocols/horizon"

	"github.com/hyperledger/fabric/core/chaincode/shim"
)

const stellar_base_URL = "https://horizon-testnet.stellar.org/transactions/"

// ============================================================================================================================
// Get Marble - get a marble asset from ledger
// ============================================================================================================================
func get_marble(stub shim.ChaincodeStubInterface, id string) (Marble, error) {
	var marble Marble
	marbleAsBytes, err := stub.GetState(id) //getState retreives a key/value from the ledger
	if err != nil {                         //this seems to always succeed, even if key didn't exist
		return marble, errors.New("Failed to find marble - " + id)
	}
	json.Unmarshal(marbleAsBytes, &marble) //un stringify it aka JSON.parse()

	if marble.Id != id { //test if marble is actually here or just nil
		return marble, errors.New("Marble does not exist - " + id)
	}

	return marble, nil
}

// ============================================================================================================================
// Get Owner - get the owner asset from ledger
// ============================================================================================================================
func get_owner(stub shim.ChaincodeStubInterface, id string) (Owner, error) {
	var owner Owner
	ownerAsBytes, err := stub.GetState(id) //getState retreives a key/value from the ledger
	if err != nil {                        //this seems to always succeed, even if key didn't exist
		return owner, errors.New("Failed to get owner - " + id)
	}
	json.Unmarshal(ownerAsBytes, &owner) //un stringify it aka JSON.parse()

	if len(owner.Username) == 0 { //test if owner is actually here or just nil
		return owner, errors.New("Owner does not exist - " + id + ", '" + owner.Username + "' '" + owner.Company + "'")
	}

	return owner, nil
}

// ========================================================
// Input Sanitation - dumb input checking, look for empty strings
// ========================================================
func sanitize_arguments(strs []string) error {
	for i, val := range strs {
		if len(val) <= 0 {
			return errors.New("Argument " + strconv.Itoa(i) + " must be a non-empty string")
		}
		if len(val) > 32 {
			return errors.New("Argument " + strconv.Itoa(i) + " must be <= 32 characters")
		}
	}
	return nil
}

func is_payment_done_for_offer(offer *Offer, accountId, stellar_transaction_id string) (bool, error) {

	var payment horizon.Payment

	resp, err := http.Get(stellar_base_URL + stellar_transaction_id + "/payments?limit=1")
	if err != nil {
		return false, errors.New(" error getting payment details from stellar. Please try again later")
	}

	decodeResponse(resp, &payment)
	paymentAmount, err := strconv.Atoi(payment.Amount) // payment has from, to and amount details
	if err != nil {
		return false, errors.New("Unable to parse amount in payment")
	}

	resp, err = http.Get(stellar_base_URL + stellar_transaction_id)
	if err != nil {
		return false, errors.New(" error getting transaction details from stellar. Please try again later")
	}

	var transaction hProtocol.Transaction
	decodeResponse(resp, &transaction) // transaction has Memo. Memo is set to offerId so that payment can be linked to offer.

	if payment.To == accountId && paymentAmount == offer.OfferPrice && transaction.MemoType == "text" && transaction.Memo == offer.Id {
		return true, nil
	} else {
		return false, nil
	}

}

func decodeResponse(resp *http.Response, object interface{}) (err error) {
	defer resp.Body.Close()
	decoder := json.NewDecoder(resp.Body)

	if !(resp.StatusCode >= 200 && resp.StatusCode < 300) {
		horizonError := &horizon.Error{
			Response: resp,
		}
		decodeError := decoder.Decode(&horizonError.Problem)
		if decodeError != nil {
			return decodeError
		}
		return horizonError
	}

	err = decoder.Decode(&object)
	if err != nil {
		return
	}
	return
}

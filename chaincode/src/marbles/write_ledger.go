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
	"fmt"
	"strconv"
	"strings"

	"github.com/hyperledger/fabric/core/chaincode/shim"
	pb "github.com/hyperledger/fabric/protos/peer"
)

// ============================================================================================================================
// write() - genric write variable into ledger
//
// Shows Off PutState() - writting a key/value into the ledger
//
// Inputs - Array of strings
//    0   ,    1
//   key  ,  value
//  "abc" , "test"
// ============================================================================================================================
func write(stub shim.ChaincodeStubInterface, args []string) pb.Response {
	var key, value string
	var err error
	fmt.Println("starting write")

	if len(args) != 2 {
		return shim.Error("Incorrect number of arguments. Expecting 2. key of the variable and value to set")
	}

	// input sanitation
	err = sanitize_arguments(args)
	if err != nil {
		return shim.Error(err.Error())
	}

	key = args[0] //rename for funsies
	value = args[1]
	err = stub.PutState(key, []byte(value)) //write the variable into the ledger
	if err != nil {
		return shim.Error(err.Error())
	}

	fmt.Println("- end write")
	return shim.Success(nil)
}

// ============================================================================================================================
// delete_marble() - remove a marble from state and from marble index
//
// Shows Off DelState() - "removing"" a key/value from the ledger
//
// Inputs - Array of strings
//      0      ,         1
//     id      ,  authed_by_company
// "m999999999", "united marbles"
// ============================================================================================================================
func delete_marble(stub shim.ChaincodeStubInterface, args []string) pb.Response {
	fmt.Println("starting delete_marble")

	if len(args) != 2 {
		return shim.Error("Incorrect number of arguments. Expecting 2")
	}

	// input sanitation
	err := sanitize_arguments(args)
	if err != nil {
		return shim.Error(err.Error())
	}

	id := args[0]
	authed_by_company := args[1]

	// get the marble
	marble, err := get_marble(stub, id)
	if err != nil {
		fmt.Println("Failed to find marble by id " + id)
		return shim.Error(err.Error())
	}

	// check authorizing company (see note in set_owner() about how this is quirky)
	if marble.Owner.Company != authed_by_company {
		return shim.Error("The company '" + authed_by_company + "' cannot authorize deletion for '" + marble.Owner.Company + "'.")
	}

	// remove the marble
	err = stub.DelState(id) //remove the key from chaincode state
	if err != nil {
		return shim.Error("Failed to delete state")
	}

	fmt.Println("- end delete_marble")
	return shim.Success(nil)
}

// ============================================================================================================================
// Init Marble - create a new marble, store into chaincode state
//
// Shows off building a key's JSON value manually
//
// Inputs - Array of strings
//      0      ,    1  ,  2  ,      3          ,       4
//     id      ,  color, size,     owner id    ,  authing company
// "m999999999", "blue", "35", "o9999999999999", "united marbles"
// ============================================================================================================================
func init_marble(stub shim.ChaincodeStubInterface, args []string) pb.Response {
	var err error
	fmt.Println("starting init_marble")

	if len(args) != 5 {
		return shim.Error("Incorrect number of arguments. Expecting 5")
	}

	//input sanitation
	err = sanitize_arguments(args)
	if err != nil {
		return shim.Error(err.Error())
	}

	id := args[0]
	color := strings.ToLower(args[1])
	owner_id := args[3]
	authed_by_company := args[4]
	size, err := strconv.Atoi(args[2])
	if err != nil {
		return shim.Error("3rd argument must be a numeric string")
	}

	//check if new owner exists
	owner, err := get_owner(stub, owner_id)
	if err != nil {
		fmt.Println("Failed to find owner - " + owner_id)
		return shim.Error(err.Error())
	}

	//check authorizing company (see note in set_owner() about how this is quirky)
	if owner.Company != authed_by_company {
		return shim.Error("The company '" + authed_by_company + "' cannot authorize creation for '" + owner.Company + "'.")
	}

	//check if marble id already exists
	marble, err := get_marble(stub, id)
	if err == nil {
		fmt.Println("This marble already exists - " + id)
		fmt.Println(marble)
		return shim.Error("This marble already exists - " + id) //all stop a marble by this id exists
	}

	//build the marble json string manually
	str := `{
		"docType":"marble", 
		"id": "` + id + `", 
		"color": "` + color + `", 
		"size": ` + strconv.Itoa(size) + `, 
		"owner": {
			"id": "` + owner_id + `", 
			"username": "` + owner.Username + `", 
			"company": "` + owner.Company + `"
		}
	}`
	err = stub.PutState(id, []byte(str)) //store marble with id as key
	if err != nil {
		return shim.Error(err.Error())
	}

	fmt.Println("- end init_marble")
	return shim.Success(nil)
}

// ============================================================================================================================
// Init Owner - create a new owner aka end user, store into chaincode state
//
// Shows off building key's value from GoLang Structure
//
// Inputs - Array of Strings
//           0     ,     1   ,   2
//      owner id   , username, company
// "o9999999999999",     bob", "united marbles"
// ============================================================================================================================
func init_owner(stub shim.ChaincodeStubInterface, args []string) pb.Response {
	var err error
	fmt.Println("starting init_owner")

	if len(args) != 3 {
		return shim.Error("Incorrect number of arguments. Expecting 3")
	}

	//input sanitation
	err = sanitize_arguments(args)
	if err != nil {
		return shim.Error(err.Error())
	}

	var owner Owner
	owner.ObjectType = "marble_owner"
	owner.Id = args[0]
	owner.Username = strings.ToLower(args[1])
	owner.Company = args[2]
	owner.Enabled = true
	fmt.Println(owner)

	//check if user already exists
	_, err = get_owner(stub, owner.Id)
	if err == nil {
		fmt.Println("This owner already exists - " + owner.Id)
		return shim.Error("This owner already exists - " + owner.Id)
	}

	//store user
	ownerAsBytes, _ := json.Marshal(owner)      //convert to array of bytes
	err = stub.PutState(owner.Id, ownerAsBytes) //store owner by its Id
	if err != nil {
		fmt.Println("Could not store user")
		return shim.Error(err.Error())
	}

	fmt.Println("- end init_owner marble")
	return shim.Success(nil)
}

// ============================================================================================================================
// Set Owner on Marble
//
// Shows off GetState() and PutState()
//
// Inputs - Array of Strings
//       0     ,        1      ,        2
//  marble id  ,  to owner id  , company that auth the transfer
// "m999999999", "o99999999999", united_mables"
// ============================================================================================================================
func set_owner(stub shim.ChaincodeStubInterface, args []string) pb.Response {
	var err error
	fmt.Println("starting set_owner")

	// this is quirky
	// todo - get the "company that authed the transfer" from the certificate instead of an argument
	// should be possible since we can now add attributes to the enrollment cert
	// as is.. this is a bit broken (security wise), but it's much much easier to demo! holding off for demos sake

	if len(args) != 3 {
		return shim.Error("Incorrect number of arguments. Expecting 3")
	}

	// input sanitation
	err = sanitize_arguments(args)
	if err != nil {
		return shim.Error(err.Error())
	}

	var marble_id = args[0]
	var new_owner_id = args[1]
	var authed_by_company = args[2]
	fmt.Println(marble_id + "->" + new_owner_id + " - |" + authed_by_company)

	// check if user already exists
	owner, err := get_owner(stub, new_owner_id)
	if err != nil {
		return shim.Error("This owner does not exist - " + new_owner_id)
	}

	// get marble's current state
	marbleAsBytes, err := stub.GetState(marble_id)
	if err != nil {
		return shim.Error("Failed to get marble")
	}
	res := Marble{}
	json.Unmarshal(marbleAsBytes, &res) //un stringify it aka JSON.parse()

	// check authorizing company
	if res.Owner.Company != authed_by_company {
		return shim.Error("The company '" + authed_by_company + "' cannot authorize transfers for '" + res.Owner.Company + "'.")
	}

	// transfer the marble
	res.Owner.Id = new_owner_id //change the owner
	res.Owner.Username = owner.Username
	res.Owner.Company = owner.Company
	jsonAsBytes, _ := json.Marshal(res)       //convert to array of bytes
	err = stub.PutState(args[0], jsonAsBytes) //rewrite the marble with id as key
	if err != nil {
		return shim.Error(err.Error())
	}

	fmt.Println("- end set owner")
	return shim.Success(nil)
}

// ============================================================================================================================
// Owner sets the Marble on sale
//
//
// Inputs - Array of Strings
//       0     ,        1      ,                         2
//  marble id  ,  company that auth the transfer  ,   minPrice
// "m999999999", "united_mables",                       100
// ============================================================================================================================

func mark_for_sale(stub shim.ChaincodeStubInterface, args []string) pb.Response {
	var err error
	fmt.Println("starting mark_for_sale")

	// this is quirky
	// todo - get the "company that authed the transfer" from the certificate instead of an argument
	// should be possible since we can now add attributes to the enrollment cert
	// as is.. this is a bit broken (security wise), but it's much much easier to demo! holding off for demos sake

	if len(args) != 3 {
		return shim.Error("Incorrect number of arguments. Expecting 3")
	}

	// input sanitation
	err = sanitize_arguments(args)
	if err != nil {
		return shim.Error(err.Error())
	}

	var marble_id = args[0]
	var authed_by_company = args[1]
	min_price, err2 := strconv.Atoi(args[2])

	if err2 != nil {
		return shim.Error("3rd argument must be a numeric string")
	}
	fmt.Println(marble_id + "->" + strconv.Itoa(min_price) + " - |" + authed_by_company)

	// get marble's current state
	marbleAsBytes, err := stub.GetState(marble_id)
	if err != nil {
		return shim.Error("Failed to get marble")
	}
	res := Marble{}
	json.Unmarshal(marbleAsBytes, &res) //un stringify it aka JSON.parse()

	// check authorizing company
	if res.Owner.Company != authed_by_company {
		return shim.Error("The company '" + authed_by_company + "' cannot authorize offer_for_sale for '" + res.Owner.Company + "'.")
	}

	// mark the marble for sale
	res.IsForSale = true     //set for Sale
	res.MinPrice = min_price // set minPrice

	jsonAsBytes, _ := json.Marshal(res)       //convert to array of bytes
	err = stub.PutState(args[0], jsonAsBytes) //rewrite the marble with id as key
	if err != nil {
		return shim.Error(err.Error())
	}

	fmt.Println("- end mark_for_sale")
	return shim.Success(nil)

}

// ============================================================================================================================
// Buyer makes offer for a Marble on sale
//
//
// Inputs - Array of Strings
//       0     ,   1,                      2      ,                         3            4
//  marble id  ,  buyer_id         company that auth the transfer  ,   offerPrice      offerId
// "m999999999",   o99999999,        "united_mables",                       200         offer99999999
// ============================================================================================================================

func make_offer(stub shim.ChaincodeStubInterface, args []string) pb.Response {

	var err error
	fmt.Println("starting mark_for_sale")

	// this is quirky
	// todo - get the "company that authed the transfer" from the certificate instead of an argument
	// should be possible since we can now add attributes to the enrollment cert
	// as is.. this is a bit broken (security wise), but it's much much easier to demo! holding off for demos sake

	if len(args) != 5 {
		return shim.Error("Incorrect number of arguments. Expecting 3")
	}

	// input sanitation
	err = sanitize_arguments(args)
	if err != nil {
		return shim.Error(err.Error())
	}

	var marble_id = args[0]
	var buyer_id = args[1]
	var authed_by_company = args[2]
	offer_price, err2 := strconv.Atoi(args[3])
	var offer_id = args[4]

	if err2 != nil {
		return shim.Error("4th argument must be a numeric string")
	}
	fmt.Println(marble_id + "->" + buyer_id + "->" + offer_id + "->" + strconv.Itoa(offer_price) + " - |" + authed_by_company)

	// check if user already exists
	buyer, err := get_owner(stub, buyer_id)
	if err != nil {
		return shim.Error("This buyer does not exist - " + buyer_id)
	}

	marble, err := get_marble(stub, marble_id)
	if err != nil {
		return shim.Error("This marble does not exist -" + marble_id)
	}

	var offer Offer
	offer.Id = offer_id
	offer.Buyer = buyer
	offer.Marble = marble
	offer.OfferPrice = offer_price
	offer.Status = "PROPOSED"

	//store user
	offerAsBytes, _ := json.Marshal(offer)      //convert to array of bytes
	err = stub.PutState(offer.Id, offerAsBytes) //store owner by its Id
	if err != nil {
		fmt.Println("Could not store offer")
		return shim.Error(err.Error())
	}

	fmt.Println("- end make_offer")
	return shim.Success(nil)

}

// ============================================================================================================================
// Seller accepts offer for a Marble on sale
//
//
// Inputs - Array of Strings
//       0     ,                             1
//  offerId_id  ,             company that auth the transfer
// "offer999999999",           "united_mables"
// ============================================================================================================================
func accept_offer(stub shim.ChaincodeStubInterface, args []string) pb.Response {
	var err error
	fmt.Println("starting mark_for_sale")

	// this is quirky
	// todo - get the "company that authed the transfer" from the certificate instead of an argument
	// should be possible since we can now add attributes to the enrollment cert
	// as is.. this is a bit broken (security wise), but it's much much easier to demo! holding off for demos sake

	if len(args) != 2 {
		return shim.Error("Incorrect number of arguments. Expecting 3")
	}

	// input sanitation
	err = sanitize_arguments(args)
	if err != nil {
		return shim.Error(err.Error())
	}

	var offer_id = args[0]
	var authed_by_company = args[1]

	//check if offer exists and is authed by company to which owner of marble belongs.

	fmt.Println(offer_id + " - |" + authed_by_company)

	offerAsBytes, err := stub.GetState(offer_id)
	if err != nil {
		return shim.Error("This offer does not exist")
	}

	offer := Offer{}
	json.Unmarshal(offerAsBytes, &offer)

	if offer.Marble.Owner.Company != authed_by_company {
		return shim.Error("This user is not authorized to perform this operation")
	}

	offer.Status = "ACCEPTED"

	//store user
	updateOfferAsBytes, _ := json.Marshal(offer)      //convert to array of bytes
	err = stub.PutState(offer.Id, updateOfferAsBytes) //store owner by its Id
	if err != nil {
		fmt.Println("Could not update offer")
		return shim.Error(err.Error())
	}

	fmt.Println("- end accept offer")
	return shim.Success(nil)

}

// ============================================================================================================================
// Seller indicates that payment is complete for a given offer
//
//
// Inputs - Array of Strings
//       0     ,                            , 1
//  offerId_id  ,                 , stellar_transaction_id
// "offer999999999"               , bbfaf6c5d4a0ddbd69f4592736986a7596b5b18dec6fde0658f12fb2e6900d81
// ============================================================================================================================

func payment_complete_against_offer(stub shim.ChaincodeStubInterface, args []string) pb.Response {
	var err error
	fmt.Println("starting mark_for_sale")

	if len(args) != 2 {
		return shim.Error("Incorrect number of arguments. Expecting 3")
	}

	// input sanitation
	err = sanitize_arguments(args)
	if err != nil {
		return shim.Error(err.Error())
	}

	var offer_id = args[0]
	var stellar_transaction_id = args[1]

	fmt.Println(offer_id + "-> " + stellar_transaction_id)

	//check if offer exists
	offerAsBytes, err := stub.GetState(offer_id)
	if err != nil {
		return shim.Error("This offer does not exist")
	}

	offer := Offer{}
	json.Unmarshal(offerAsBytes, &offer)

	owner, err := get_owner(stub, offer.Marble.Owner.Id)
	if err != nil {
		return shim.Error(" Transfer not done. Current Onwer not found")
	}

	paymentDone, err := is_payment_done_for_offer(&offer, owner.AccountId, stellar_transaction_id)

	if err != nil {
		return shim.Error("Unable to verify payment information from stellar. Please try again later")
	}

	if paymentDone {
		set_owner(stub, []string{offer.Buyer.Id}) // transfer the marble to Buyer

		//TODO update offer to indicate completion of transfer.

		return shim.Success(nil)

	} else {
		return shim.Error(" Payment not done in stellar or mismatch in payment information.  ")
	}

}

// ============================================================================================================================
// Disable Marble Owner
//
// Shows off PutState()
//
// Inputs - Array of Strings
//       0     ,        1
//  owner id       , company that auth the transfer
// "o9999999999999", "united_mables"
// ============================================================================================================================
func disable_owner(stub shim.ChaincodeStubInterface, args []string) pb.Response {
	var err error
	fmt.Println("starting disable_owner")

	if len(args) != 2 {
		return shim.Error("Incorrect number of arguments. Expecting 2")
	}

	// input sanitation
	err = sanitize_arguments(args)
	if err != nil {
		return shim.Error(err.Error())
	}

	var owner_id = args[0]
	var authed_by_company = args[1]

	// get the marble owner data
	owner, err := get_owner(stub, owner_id)
	if err != nil {
		return shim.Error("This owner does not exist - " + owner_id)
	}

	// check authorizing company
	if owner.Company != authed_by_company {
		return shim.Error("The company '" + authed_by_company + "' cannot change another companies marble owner")
	}

	// disable the owner
	owner.Enabled = false
	jsonAsBytes, _ := json.Marshal(owner)     //convert to array of bytes
	err = stub.PutState(args[0], jsonAsBytes) //rewrite the owner
	if err != nil {
		return shim.Error(err.Error())
	}

	fmt.Println("- end disable_owner")
	return shim.Success(nil)
}

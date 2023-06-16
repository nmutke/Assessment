package main

import (
	"encoding/json"
	"fmt"
	"log"
	"time"

	"github.com/golang/protobuf/ptypes"
	"github.com/hyperledger/fabric-contract-api-go/contractapi"
)

type SmartContract struct {
	contractapi.Contract
}

type Asset struct {
	Id             string `json:"Id"`
	Owner          string `json:"Owner"`
	Price          int    `json:"Price"`
}

type HistoryResult struct {
	Record    *Asset    `json:"record"`
	TxId      string    `json:"txId"`
	Timestamp time.Time `json:"timestamp"`
	IsDelete  bool      `json:"isDelete"`
}

type PaginatedResult struct {
	Records             []*Asset `json:"records"`
	FetchedRecordsCount int32    `json:"fetchedRecordsCount"`
	Bookmark            string   `json:"bookmark"`
}

// Check Existing Asset
func (s *SmartContract) IsExists(ctx contractapi.TransactionContextInterface, id string) (bool, error) {
  
	assetJson, err := ctx.GetStub().GetState(id)
  	if err != nil {
    	return false, fmt.Errorf("Error in Reading an Asset: %v", err)
  	}

  	return assetJson != nil, nil
}

// Create An Asset
func (s *SmartContract) CreateAsset(ctx contractapi.TransactionContextInterface, id string, owner string, price int) error {
	
	isExists, err := s.IsExists(ctx, id)
	if err != nil {
    	return err
  	}
	if isExists {
		return fmt.Errorf("Asset %s already exists", id)
	}

	asset := Asset {
		Id:		id,
		Owner: 	owner,
		Price: 	price,
	}

	assetJson, err := json.Marshal(asset)
	if err != nil {
		return err
	}

	return ctx.GetStub().PutState(id, assetJson)
}

// Delete An Asset
func (s *SmartContract) DeleteAsset(ctx contractapi.TransactionContextInterface, id string) error {
	
	isExists, err := s.IsExists(ctx, id)
	if err != nil {
		return err
	}
	if !isExists {
		return fmt.Errorf("Asset %s does not exists", id)
	}

	return ctx.GetStub().DelState(id)
}

// Update an Asset
func (s *SmartContract) UpdateAsset(ctx contractapi.TransactionContextInterface, id string, owner string, price int) error {
		
	isExists, err := s.IsExists(ctx, id)
	if err != nil {
		return err
	}
	if !isExists {
		return fmt.Errorf("Asset %s does not exist", id)
	}

	asset := Asset{
		Id:             id,
		Owner:          owner,
		Price: 			price,
	}
	assetJSON, err := json.Marshal(asset)
	if err != nil {
		return err
	}

	return ctx.GetStub().PutState(id, assetJSON)
}


// Get History of an Asset
func (s *SmartContract) getHistoryofAsset(ctx contractapi.TransactionContextInterface, id string) ([]HistoryResult, error) {
	
	resultIterator, err := ctx.GetStub().GetHistoryForKey(id)
	if err != nil {
		return nil, fmt.Errorf(err.Error())
	}
	defer resultIterator.Close()

	var records []HistoryResult
	for resultIterator.HasNext() {
		response, err := resultIterator.Next()
		if err != nil {
			return nil, err
		}

		var asset Asset
		if len(response.Value) > 0 {
			err = json.Unmarshal(response.Value, &asset)
			if err != nil {
				return nil, err
			}
		} else {
			asset = Asset{
				Id: id,
			}
		}

		timestamp, err := ptypes.Timestamp(response.Timestamp)
		if err != nil {
			return nil, err
		}

		record := HistoryResult{
			TxId:      response.TxId,
			Timestamp: timestamp,
			Record:    &asset,
			IsDelete:  response.IsDelete,
		}
		records = append(records, record)
	}

	return records, nil
}

func main() {
	chaincode, err := contractapi.NewChaincode(&SmartContract{})
	if err != nil {
		log.Panicf("Error Creating Asset Chaincode: %v", err)
	}

	if err := chaincode.Start(); err != nil {
		log.Panicf("Error Starting Asset Chaincode: %v", err)
	}
}

package main

import (
	"encoding/json"
	"fmt"
	"log"
	"time"

	"github.com/hyperledger/fabric-contract-api-go/contractapi"
)

type SmartContract struct (
	contractapi.Contract
)

type Asset struct {
	Id		string	`json:"Id"`
	Owner	string	`json:"owner"`
	Price	int		`json:"price"`	
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
func (s *SmartContract) IsExists(ctx contractapi.TransactionContextInterface, id string) (bool, err) {
  
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
		Price: 	price
	}

	assetJson, err := json.Marshal(asset)
	if err != nill {
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
	
	currentTimestamp := time.Now().Unix()

	assetJson, err := ctx.GetStub().GetState(id)
	if err != nil {
		return nil
	}

	creationTimestamp := parseCreationTimestamp(assetJson)

	timeLimit := 5 * 60
	if currentTimestamp-creationTimestamp > timeLimit {
		return fmt.Errorf("Asset update not allowed after 5 minutes of creation time")
	}
	
	checkAsset := &Asset{}
	err := json.Unmarshal(assetJson, checkAsset)
	if err != nil {
		return err
	}

	if asset.Owner != owner {
		return fmt.Errorf("Only the owner can update the asset")
	}

	asset := Asset {
		Id:		id,
		Owner: 	owner,
		Price: 	price
	}	

	assetJson, err := josn.Marshal(asset)
	if err != nil {
		return err
	}
	
	return ctx.GetStub().PutState(id, assetJson)
}

func (s *SmartContract) getAssetWithPagination(ctx contractapi.TransactionContextInterface, id string, pageSize int) (*PaginatedResult, error) {

	resultIterator, responseMetadata, err := ctx.GetStub().GetQueryResultWithPagination(id, pageSize, bookmark)
	if err != nil {
		return nil, err
	}
	defer resultIterator.Close()

	var assets []*Asset
	for resultIterator.HasNext() {
		queryResult, err := resultIterator.Next()
		if err != nil {
			return nil, err
		}
		var asset Asset
		err = json.Unmarshal(queryResult.Value, &asset)
		if err != nil {
			return nil, err
		}
		assets = append(assets, &asset)
	}

	return &PaginatedResult{
		Records:				assets,
		FetchedRecordsCount:	responseMetadata.FetchedRecordsCount,
		Bookmark:				responseMetadata.Bookmark
	}, nil
}

// Get History of an Asset
func (s *SmartContract) getHistoryofAsset(ctx contractapi.TransactionContextInterface, id string) (string, error) {
	
	resultIterator, err := ctx.GetStub().GetHistoryForKey(id)
	if err != nil {
		return "", fmt.Errorf(err.Error())
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
				ID: assetID,
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
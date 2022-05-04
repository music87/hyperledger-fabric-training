package smartcontract

import (
	"fmt"
	"encoding/json"
	"github.com/hyperledger/fabric-contract-api-go/contractapi"
)

// SmartContract provides functions for managing archives
type SmartContract struct {
	contractapi.Contract
}

// User Data struct
type User struct {
	ID           string `json:"id"`
	Name         string `json:"name"`
	Email        string `json:"email"`
	Transactions []Transaction `json:"transactions,omitempty" metadata:",optional"`
}
// Transaction Data struct
type Transaction struct {
	Hash         string `json:"hash"`
	Amount       string `json:"amount"`
	Currency string `json:"currency"`
	Date    string `json:"date"`
}

type TransactionHashMapUserId struct {
	UserId           string `json:"user_id"`
}

func (s *SmartContract) InitLedger(ctx contractapi.TransactionContextInterface) error {
	return nil
}

func (s *SmartContract) UserExists(ctx contractapi.TransactionContextInterface, id string) (bool, error) {
	assetJSON, err := ctx.GetStub().GetState(id)
	if err != nil {
		return false, fmt.Errorf("failed to read from world state: %v", err)
	}
	return assetJSON != nil, nil
}

func (s *SmartContract) CreateUser(ctx contractapi.TransactionContextInterface, id string, name string, email string) error {
	exists, err := s.UserExists(ctx, id)
	if err != nil {
		return err
	}
	if exists {
		return fmt.Errorf("the user %s already exists", id)
	}

	user := User{
		ID:    id,
		Name:  name,
		Email: email,
	}
	userJson, err := json.Marshal(user)
	if err != nil {
		return err
	}
	
	return ctx.GetStub().PutState(id, userJson)
}

func (s *SmartContract) GetUser(ctx contractapi.TransactionContextInterface, id string) (*User, error) {
	userJson, err := ctx.GetStub().GetState(id)
	if err != nil {
		return nil, fmt.Errorf("failed to read from world state: %v", err)
	}
	if userJson == nil {
		return nil, fmt.Errorf("the user %s does not exist", id)
	}

	var user User
	err = json.Unmarshal(userJson, &user)
	if err != nil {
		return nil, err
	}

	return &user, nil
}

func (s *SmartContract) UpdateUser(ctx contractapi.TransactionContextInterface, id string, name string, email string) error {
	user, err := s.GetUser(ctx, id)
	if err != nil {
		return err
	}
	user.Email = email
	user.Name = name
	userJson, err := json.Marshal(user)
	if err != nil {
		return err
	}

	return ctx.GetStub().PutState(id, userJson)
}

func (s *SmartContract) DeleteUser(ctx contractapi.TransactionContextInterface, id string) error {
	exists, err := s.UserExists(ctx, id)
	if err != nil {
		return err
	}
	if !exists {
		return fmt.Errorf("the user %s does not exist", id)
	}

	return ctx.GetStub().DelState(id)
}

func (s *SmartContract) GetAllUsers(ctx contractapi.TransactionContextInterface) ([]*User, error) {
	resultsIterator, err := ctx.GetStub().GetStateByRange("", "")
	if err != nil {
		return nil, err
	}
	defer resultsIterator.Close()

	var users []*User
	for resultsIterator.HasNext() {
		queryResponse, err := resultsIterator.Next()
		if err != nil {
			return nil, err
		}

		var user User
		err = json.Unmarshal(queryResponse.Value, &user)
		if err != nil {
			return nil, err
		}
		users = append(users, &user)
	}

	return users, nil
}

func (s *SmartContract) CreateTransaction(ctx contractapi.TransactionContextInterface, userId string, hash string, amount string, currency string, date string) (bool, error) {
	user, err := s.GetUser(ctx, userId)
	if err != nil {
		return false, err
	}

	var transaction Transaction = Transaction{
		Hash:      hash,
		Amount:    amount,
		Currency:  currency,
		Date:      date,
	}
	user.Transactions = append(user.Transactions, transaction)

	userJson, err := json.Marshal(user)
	if err != nil {
		return false, err
	}

	ctx.GetStub().PutState(userId, userJson)

	var transactionHashMapUserId TransactionHashMapUserId = TransactionHashMapUserId{
		UserId:      user.ID,
	}

	transactionHashMapUserIdJson, err := json.Marshal(transactionHashMapUserId)
	if err != nil {
		return false, err
	}

	ctx.GetStub().PutState(hash, transactionHashMapUserIdJson)

	return true, nil
}

func (s *SmartContract) GetUserByTransactionHash(ctx contractapi.TransactionContextInterface, hash string) (*User, error) {
	transactionHashMapUserIdJson, err := ctx.GetStub().GetState(hash)
	if err != nil {
		return nil, fmt.Errorf("failed to read from world state: %v", err)
	}
	if transactionHashMapUserIdJson == nil {
		return nil, fmt.Errorf("the transaction %s does not exist", hash)
	}
	var transactionHashMapUserId TransactionHashMapUserId
	err = json.Unmarshal(transactionHashMapUserIdJson, &transactionHashMapUserId)
	if err != nil {
		return nil, err
	}

	user, err := s.GetUser(ctx, transactionHashMapUserId.UserId)
	if err != nil {
		return nil, err
	}
	return user, nil
}

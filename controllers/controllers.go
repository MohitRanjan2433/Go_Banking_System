package controllers

import (
	"banking/models"
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"

	"github.com/gorilla/mux"
	"github.com/joho/godotenv"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"golang.org/x/crypto/bcrypt"
)

var accountCollection *mongo.Collection

func init() {

	err := godotenv.Load()
	if err != nil{
		log.Fatalf("Error Loading .env file: %v", err)
	}

	connectionString := os.Getenv("MONGODB_URL")
	if connectionString == ""{
		log.Fatalf("MONGODB_URL not set in .env file")
	}

	clientOptions := options.Client().ApplyURI(connectionString)

	client, err := mongo.Connect(context.TODO(), clientOptions)
	if err != nil {
		log.Fatal(err)
	}

	// Check the connection
	err = client.Ping(context.TODO(), nil)
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println("MongoDB Connection Successful")

	accountCollection = client.Database("banking").Collection("account")

	fmt.Println("Account Collections are ready")
}

func createAccount(account models.Account) error {
	existingAccount := models.Account{}
	err := accountCollection.FindOne(context.Background(), bson.M{"security.nickname": account.Security.NickName}).Decode(&existingAccount)

	if err == nil {
		return fmt.Errorf("account with nickname %s already exists", account.Security.NickName)
	} else if err != mongo.ErrNoDocuments {
		log.Printf("Error checking for existing account: %v", err)
		return err
	}

	if account.ID.IsZero() {
		account.ID = primitive.NewObjectID()
	}

	// Hashing the password
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(account.Security.Password), bcrypt.DefaultCost)
	if err != nil {
		log.Printf("Error hashing password: %v", err)
		return err
	}
	account.Security.Password = string(hashedPassword)

	insertResult, err := accountCollection.InsertOne(context.Background(), account)
	if err != nil {
		log.Printf("Could not create account: %v", err)
		return err
	}

	fmt.Printf("Inserted a single document: %v\n", insertResult.InsertedID)
	return nil
}

func getAllAccount() ([]models.Account, error) {
	filter := bson.D{{}}

	cursor, err := accountCollection.Find(context.Background(), filter)
	if err != nil {
		log.Printf("Error fetching Account: %v", err)
		return nil, err
	}

	defer cursor.Close(context.Background())

	var accounts []models.Account
	for cursor.Next(context.Background()) {
		var account models.Account
		if err := cursor.Decode(&account); err != nil {
			log.Printf("Error decoding account document: %v", err)
			return nil, err
		}

		accounts = append(accounts, account)
	}

	if err := cursor.Err(); err != nil {
		log.Printf("Error during cursor iteration: %v", err)
		return nil, err
	}
	return accounts, nil
}

func getOneAccount(accountID primitive.ObjectID) (*models.Account, error) {
	var account models.Account
	err := accountCollection.FindOne(context.Background(), bson.M{"_id": accountID}).Decode(&account)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			return nil, fmt.Errorf("No account found with ID %s", accountID.Hex())
		}
		return nil, err
	}
	return &account, nil
}

func CreateAccountHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Invalid request method", http.StatusMethodNotAllowed)
		return
	}

	var account models.Account
	if err := json.NewDecoder(r.Body).Decode(&account); err != nil {
		http.Error(w, "Cannot parse JSON", http.StatusBadRequest)
		return
	}

	if err := createAccount(account); err != nil {
		if mongo.IsDuplicateKeyError(err) {
			http.Error(w, err.Error(), http.StatusConflict)
		} else {
			http.Error(w, err.Error(), http.StatusInternalServerError)
		}
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	if err := json.NewEncoder(w).Encode(account); err != nil {
		http.Error(w, "Failed to write response", http.StatusInternalServerError)
	}
}

func GetAllAccountHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Invalid request method", http.StatusMethodNotAllowed)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)

	accounts, err := getAllAccount()
	if err != nil {
		http.Error(w, "Failed to fetch account", http.StatusInternalServerError)
		return
	}
	for _, acc := range accounts {
		fmt.Println("Account ID: ", acc.ID.Hex())
		fmt.Println("Account Member Name: ", acc.Security.NickName)
	}

	if err := json.NewEncoder(w).Encode(accounts); err != nil {
		http.Error(w, "Failed to write response", http.StatusInternalServerError)
		return
	}
}

func GetOneAccountHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Invalid request method", http.StatusMethodNotAllowed)
		return
	}

	// THIS IS USING QUERY PARAMETER---->>>
	// accountIDHex := r.URL.Query().Get("id")

	// HERE USING PATH PARAMETER-->
	vars := mux.Vars(r)
	accountIDHex := vars["id"]
	if accountIDHex == "" {
		http.Error(w, "Missing account ID", http.StatusBadRequest)
		return
	}

	accountID, err := primitive.ObjectIDFromHex(accountIDHex)
	if err != nil {
		http.Error(w, "Invalid account ID format", http.StatusBadRequest)
		return
	}

	account, err := getOneAccount(accountID)
	if err != nil {
		http.Error(w, err.Error(), http.StatusNotFound)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(account); err != nil {
		http.Error(w, "Failed to write response", http.StatusInternalServerError)
	}
}

func DeleteAccountHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodDelete {
		http.Error(w, "Invalid request method", http.StatusMethodNotAllowed)
		return
	}

	vars := mux.Vars(r)
	accountIDHex := vars["id"]

	if accountIDHex == "" {
		http.Error(w, "Missing account ID", http.StatusBadRequest)
		return
	}

	accountID, err := primitive.ObjectIDFromHex(accountIDHex)
	if err != nil {
		http.Error(w, "Invalid account ID format", http.StatusBadRequest)
		return
	}

	filter := bson.M{"_id": accountID}
	deleteResult, err := accountCollection.DeleteOne(context.Background(), filter)
	if err != nil {
		http.Error(w, "Failed to delete account", http.StatusInternalServerError)
		return
	}
	fmt.Println("Account Successfully deleted")

	if deleteResult.DeletedCount == 0 {
		http.Error(w, "No account found with the given ID", http.StatusNotFound)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

func DeleteAllAccountsHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodDelete {
		http.Error(w, "Invalid request method", http.StatusMethodNotAllowed)
		return
	}

	filter := bson.D{} // Empty filter matches all documents
	deleteResult, err := accountCollection.DeleteMany(context.Background(), filter)
	if err != nil {
		http.Error(w, "Failed to delete accounts", http.StatusInternalServerError)
		return
	}

	response := map[string]interface{}{
		"deleted_count": deleteResult.DeletedCount,
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	if err := json.NewEncoder(w).Encode(response); err != nil {
		http.Error(w, "Failed to write response", http.StatusInternalServerError)
	}
}

func PaymentHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Invalid request method", http.StatusMethodNotAllowed)
		return
	}

	var paymentRequest struct {
		SenderID   string  `json:"SenderID"`
		ReceiverID string  `json:"ReceiverID"`
		Amount     float64 `json:"Amount"`
	}

	if err := json.NewDecoder(r.Body).Decode(&paymentRequest); err != nil {
		http.Error(w, "Cannot parse JSON", http.StatusBadRequest)
		return
	}

	senderID, err := primitive.ObjectIDFromHex(paymentRequest.SenderID)
	if err != nil {
		http.Error(w, "Invalid SenderID format", http.StatusBadRequest)
		return
	}

	receiverID, err := primitive.ObjectIDFromHex(paymentRequest.ReceiverID)
	if err != nil {
		http.Error(w, "Invalid ReceiverID format", http.StatusBadRequest)
		return
	}

	payment := models.Payment{
		SenderID:   senderID,
		ReceiverID: receiverID,
		Amount:     paymentRequest.Amount,
	}

	if err := processPayment(payment); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
	w.Write([]byte("Payment processed successfully"))
}

func processPayment(payment models.Payment) error {
	session, err := accountCollection.Database().Client().StartSession()
	if err != nil {
		return fmt.Errorf("failed to start session: %v", err)
	}
	defer session.EndSession(context.Background())

	callback := func(sessCtx mongo.SessionContext) (interface{}, error) {
		var sender, receiver models.Account

		fmt.Printf("Looking for sender with ID: %s\n", payment.SenderID.Hex())
		err := accountCollection.FindOne(sessCtx, bson.M{"_id": payment.SenderID}).Decode(&sender)
		if err != nil {
			if err == mongo.ErrNoDocuments {
				return nil, fmt.Errorf("failed to find sender account: no documents in result")
			}
			return nil, fmt.Errorf("failed to find sender account: %v", err)
		}

		fmt.Printf("Looking for receiver with ID: %s\n", payment.ReceiverID.Hex())
		err = accountCollection.FindOne(sessCtx, bson.M{"_id": payment.ReceiverID}).Decode(&receiver)
		if err != nil {
			if err == mongo.ErrNoDocuments {
				return nil, fmt.Errorf("failed to find receiver account: no documents in result")
			}
			return nil, fmt.Errorf("failed to find receiver account: %v", err)
		}

		if sender.Balance < payment.Amount {
			return nil, fmt.Errorf("insufficient funds in sender account")
		}

		sender.Balance -= payment.Amount
		receiver.Balance += payment.Amount

		if _, err := accountCollection.UpdateOne(sessCtx, bson.M{"_id": sender.ID}, bson.M{"$set": bson.M{"balance": sender.Balance}}); err != nil {
			return nil, fmt.Errorf("failed to update sender balance: %v", err)
		}

		if _, err := accountCollection.UpdateOne(sessCtx, bson.M{"_id": receiver.ID}, bson.M{"$set": bson.M{"balance": receiver.Balance}}); err != nil {
			return nil, fmt.Errorf("failed to update receiver balance: %v", err)
		}

		return nil, nil
	}

	if _, err := session.WithTransaction(context.Background(), callback); err != nil {
		return fmt.Errorf("transaction failed: %v", err)
	}

	return nil
}

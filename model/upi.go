package model

import (
	"time"

	"go.mongodb.org/mongo-driver/v2/bson"
)

type VPA struct {
	ID          bson.ObjectID `bson:"_id,omitempty" json:"id"`
	UserID      bson.ObjectID `bson:"user_id" json:"user_id"`
	Address     string        `bson:"address" json:"address"` // user@bankname
	AccountID   string        `bson:"account_id" json:"account_id"`
	IsDefault   bool          `bson:"is_default" json:"is_default"`
	IsActive    bool          `bson:"is_active" json:"is_active"`
	CreatedAt   time.Time     `bson:"created_at" json:"created_at"`
	UpdatedAt   time.Time     `bson:"updated_at" json:"updated_at"`
}

type UPITransaction struct {
	ID              bson.ObjectID `bson:"_id,omitempty" json:"id"`
	UserID          bson.ObjectID `bson:"user_id" json:"user_id"`
	TxnID           string        `bson:"txn_id" json:"txn_id"`
	Type            string        `bson:"type" json:"type"` // pay | collect
	FromVPA         string        `bson:"from_vpa" json:"from_vpa"`
	ToVPA           string        `bson:"to_vpa" json:"to_vpa"`
	Amount          float64       `bson:"amount" json:"amount"`
	Note            string        `bson:"note" json:"note"`
	Status          string        `bson:"status" json:"status"` // pending | success | failed | declined
	FailureReason   string        `bson:"failure_reason,omitempty" json:"failure_reason,omitempty"`
	TransactionDate time.Time     `bson:"transaction_date" json:"transaction_date"`
	CreatedAt       time.Time     `bson:"created_at" json:"created_at"`
}

type Mandate struct {
	ID            bson.ObjectID `bson:"_id,omitempty" json:"id"`
	UserID        bson.ObjectID `bson:"user_id" json:"user_id"`
	MandateID     string        `bson:"mandate_id" json:"mandate_id"`
	PayerVPA      string        `bson:"payer_vpa" json:"payer_vpa"`
	PayeeVPA      string        `bson:"payee_vpa" json:"payee_vpa"`
	Amount        float64       `bson:"amount" json:"amount"`
	Frequency     string        `bson:"frequency" json:"frequency"` // daily | weekly | monthly | yearly | as_presented
	StartDate     time.Time     `bson:"start_date" json:"start_date"`
	EndDate       time.Time     `bson:"end_date" json:"end_date"`
	Purpose       string        `bson:"purpose" json:"purpose"`
	Status        string        `bson:"status" json:"status"` // active | paused | revoked | expired
	CreatedAt     time.Time     `bson:"created_at" json:"created_at"`
	UpdatedAt     time.Time     `bson:"updated_at" json:"updated_at"`
}

type CollectRequest struct {
	ID          bson.ObjectID `bson:"_id,omitempty" json:"id"`
	UserID      bson.ObjectID `bson:"user_id" json:"user_id"`
	FromVPA     string        `bson:"from_vpa" json:"from_vpa"`
	ToVPA       string        `bson:"to_vpa" json:"to_vpa"`
	Amount      float64       `bson:"amount" json:"amount"`
	Note        string        `bson:"note" json:"note"`
	Status      string        `bson:"status" json:"status"` // pending | approved | declined | expired
	ExpiresAt   time.Time     `bson:"expires_at" json:"expires_at"`
	CreatedAt   time.Time     `bson:"created_at" json:"created_at"`
}

// Request/Response types
type CreateVPARequest struct {
	Prefix    string `json:"prefix"` // e.g., "john.doe" -> john.doe@bankname
	AccountID string `json:"account_id"`
}

type ValidateVPARequest struct {
	VPA string `json:"vpa"`
}

type UPIPayRequest struct {
	ToVPA  string  `json:"to_vpa"`
	Amount float64 `json:"amount"`
	Note   string  `json:"note"`
}

type CollectRequestInput struct {
	FromVPA string  `json:"from_vpa"`
	Amount  float64 `json:"amount"`
	Note    string  `json:"note"`
}

type CreateMandateRequest struct {
	PayeeVPA  string    `json:"payee_vpa"`
	Amount    float64   `json:"amount"`
	Frequency string    `json:"frequency"`
	StartDate time.Time `json:"start_date"`
	EndDate   time.Time `json:"end_date"`
	Purpose   string    `json:"purpose"`
}

type VPAValidateResponse struct {
	VPA   string `json:"vpa"`
	Name  string `json:"name"`
	Valid bool   `json:"valid"`
}

package service

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/banking-superapp/upi-service/model"
	"github.com/banking-superapp/upi-service/repository"
	"go.mongodb.org/mongo-driver/v2/bson"
	"go.mongodb.org/mongo-driver/v2/mongo"
)

var (
	ErrVPANotFound   = errors.New("VPA not found or inactive")
	ErrVPAExists     = errors.New("VPA already exists")
	ErrUnauthorized  = errors.New("unauthorized")
	ErrInvalidAmount = errors.New("amount must be greater than zero")
)

const bankSuffix = "@digitalbank"

type UPIService interface {
	CreateVPA(ctx context.Context, userID string, req *model.CreateVPARequest) (*model.VPA, error)
	GetVPAs(ctx context.Context, userID string) ([]model.VPA, error)
	ValidateVPA(ctx context.Context, address string) (*model.VPAValidateResponse, error)
	Pay(ctx context.Context, userID string, req *model.UPIPayRequest) (*model.UPITransaction, error)
	Collect(ctx context.Context, userID string, req *model.CollectRequestInput) (*model.CollectRequest, error)
	GetTransactions(ctx context.Context, userID string, page, limit int64) ([]model.UPITransaction, int64, error)
	CreateMandate(ctx context.Context, userID string, req *model.CreateMandateRequest) (*model.Mandate, error)
	GetMandates(ctx context.Context, userID string) ([]model.Mandate, error)
}

type upiService struct {
	vpaRepo     repository.VPARepo
	txnRepo     repository.UPITransactionRepo
	mandateRepo repository.MandateRepo
	collectRepo repository.CollectRepo
}

func NewUPIService(vr repository.VPARepo, tr repository.UPITransactionRepo, mr repository.MandateRepo, cr repository.CollectRepo) UPIService {
	return &upiService{vpaRepo: vr, txnRepo: tr, mandateRepo: mr, collectRepo: cr}
}

func (s *upiService) CreateVPA(ctx context.Context, userID string, req *model.CreateVPARequest) (*model.VPA, error) {
	oid, err := bson.ObjectIDFromHex(userID)
	if err != nil {
		return nil, ErrUnauthorized
	}

	address := strings.ToLower(req.Prefix) + bankSuffix
	vpa := &model.VPA{
		UserID:    oid,
		Address:   address,
		AccountID: req.AccountID,
		IsDefault: true,
		IsActive:  true,
	}

	if err := s.vpaRepo.Create(ctx, vpa); err != nil {
		if mongo.IsDuplicateKeyError(err) {
			return nil, ErrVPAExists
		}
		return nil, err
	}
	return vpa, nil
}

func (s *upiService) GetVPAs(ctx context.Context, userID string) ([]model.VPA, error) {
	oid, err := bson.ObjectIDFromHex(userID)
	if err != nil {
		return nil, ErrUnauthorized
	}
	return s.vpaRepo.FindByUserID(ctx, oid)
}

func (s *upiService) ValidateVPA(ctx context.Context, address string) (*model.VPAValidateResponse, error) {
	vpa, err := s.vpaRepo.FindByAddress(ctx, address)
	if err != nil {
		if errors.Is(err, mongo.ErrNoDocuments) {
			return &model.VPAValidateResponse{VPA: address, Valid: false}, nil
		}
		return nil, err
	}
	return &model.VPAValidateResponse{
		VPA:   vpa.Address,
		Name:  "Verified User", // In production: fetch from profile service
		Valid: true,
	}, nil
}

func (s *upiService) Pay(ctx context.Context, userID string, req *model.UPIPayRequest) (*model.UPITransaction, error) {
	if req.Amount <= 0 {
		return nil, ErrInvalidAmount
	}

	oid, err := bson.ObjectIDFromHex(userID)
	if err != nil {
		return nil, ErrUnauthorized
	}

	// Get user's default VPA
	vpas, err := s.vpaRepo.FindByUserID(ctx, oid)
	if err != nil || len(vpas) == 0 {
		return nil, errors.New("no VPA found for user")
	}

	fromVPA := vpas[0].Address
	for _, v := range vpas {
		if v.IsDefault {
			fromVPA = v.Address
			break
		}
	}

	txn := &model.UPITransaction{
		UserID:          oid,
		TxnID:           generateTxnID(),
		Type:            "pay",
		FromVPA:         fromVPA,
		ToVPA:           req.ToVPA,
		Amount:          req.Amount,
		Note:            req.Note,
		Status:          "success", // In production: integrate with UPI switch
		TransactionDate: time.Now(),
	}

	if err := s.txnRepo.Create(ctx, txn); err != nil {
		return nil, err
	}
	return txn, nil
}

func (s *upiService) Collect(ctx context.Context, userID string, req *model.CollectRequestInput) (*model.CollectRequest, error) {
	if req.Amount <= 0 {
		return nil, ErrInvalidAmount
	}

	oid, err := bson.ObjectIDFromHex(userID)
	if err != nil {
		return nil, ErrUnauthorized
	}

	vpas, err := s.vpaRepo.FindByUserID(ctx, oid)
	if err != nil || len(vpas) == 0 {
		return nil, errors.New("no VPA found for user")
	}

	toVPA := vpas[0].Address
	for _, v := range vpas {
		if v.IsDefault {
			toVPA = v.Address
			break
		}
	}

	cr := &model.CollectRequest{
		UserID:    oid,
		FromVPA:   req.FromVPA,
		ToVPA:     toVPA,
		Amount:    req.Amount,
		Note:      req.Note,
		Status:    "pending",
		ExpiresAt: time.Now().Add(24 * time.Hour),
	}

	if err := s.collectRepo.Create(ctx, cr); err != nil {
		return nil, err
	}
	return cr, nil
}

func (s *upiService) GetTransactions(ctx context.Context, userID string, page, limit int64) ([]model.UPITransaction, int64, error) {
	oid, err := bson.ObjectIDFromHex(userID)
	if err != nil {
		return nil, 0, ErrUnauthorized
	}
	if page < 1 {
		page = 1
	}
	if limit < 1 || limit > 100 {
		limit = 20
	}
	return s.txnRepo.FindByUserID(ctx, oid, page, limit)
}

func (s *upiService) CreateMandate(ctx context.Context, userID string, req *model.CreateMandateRequest) (*model.Mandate, error) {
	oid, err := bson.ObjectIDFromHex(userID)
	if err != nil {
		return nil, ErrUnauthorized
	}

	vpas, err := s.vpaRepo.FindByUserID(ctx, oid)
	if err != nil || len(vpas) == 0 {
		return nil, errors.New("no VPA found for user")
	}

	payerVPA := vpas[0].Address
	mandate := &model.Mandate{
		UserID:    oid,
		MandateID: generateMandateID(),
		PayerVPA:  payerVPA,
		PayeeVPA:  req.PayeeVPA,
		Amount:    req.Amount,
		Frequency: req.Frequency,
		StartDate: req.StartDate,
		EndDate:   req.EndDate,
		Purpose:   req.Purpose,
		Status:    "active",
	}

	if err := s.mandateRepo.Create(ctx, mandate); err != nil {
		return nil, err
	}
	return mandate, nil
}

func (s *upiService) GetMandates(ctx context.Context, userID string) ([]model.Mandate, error) {
	oid, err := bson.ObjectIDFromHex(userID)
	if err != nil {
		return nil, ErrUnauthorized
	}
	return s.mandateRepo.FindByUserID(ctx, oid)
}

func generateTxnID() string {
	return fmt.Sprintf("UPI%d", time.Now().UnixNano())
}

func generateMandateID() string {
	return fmt.Sprintf("MND%d", time.Now().UnixNano())
}

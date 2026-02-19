package repository

import (
	"context"
	"time"

	"github.com/banking-superapp/upi-service/model"
	"go.mongodb.org/mongo-driver/v2/bson"
	"go.mongodb.org/mongo-driver/v2/mongo"
	"go.mongodb.org/mongo-driver/v2/mongo/options"
)

type VPARepo interface {
	Create(ctx context.Context, v *model.VPA) error
	FindByAddress(ctx context.Context, address string) (*model.VPA, error)
	FindByUserID(ctx context.Context, userID bson.ObjectID) ([]model.VPA, error)
}

type UPITransactionRepo interface {
	Create(ctx context.Context, t *model.UPITransaction) error
	FindByUserID(ctx context.Context, userID bson.ObjectID, page, limit int64) ([]model.UPITransaction, int64, error)
}

type MandateRepo interface {
	Create(ctx context.Context, m *model.Mandate) error
	FindByUserID(ctx context.Context, userID bson.ObjectID) ([]model.Mandate, error)
}

type CollectRepo interface {
	Create(ctx context.Context, c *model.CollectRequest) error
}

type vpaRepo struct{ col *mongo.Collection }
type txnRepo struct{ col *mongo.Collection }
type mandateRepo struct{ col *mongo.Collection }
type collectRepo struct{ col *mongo.Collection }

func NewVPARepo(db *mongo.Database) VPARepo         { return &vpaRepo{col: db.Collection("vpas")} }
func NewTxnRepo(db *mongo.Database) UPITransactionRepo {
	return &txnRepo{col: db.Collection("upi_transactions")}
}
func NewMandateRepo(db *mongo.Database) MandateRepo {
	return &mandateRepo{col: db.Collection("mandates")}
}
func NewCollectRepo(db *mongo.Database) CollectRepo {
	return &collectRepo{col: db.Collection("collect_requests")}
}

func (r *vpaRepo) Create(ctx context.Context, v *model.VPA) error {
	v.CreatedAt = time.Now()
	v.UpdatedAt = time.Now()
	_, err := r.col.InsertOne(ctx, v)
	return err
}

func (r *vpaRepo) FindByAddress(ctx context.Context, address string) (*model.VPA, error) {
	var v model.VPA
	err := r.col.FindOne(ctx, bson.M{"address": address, "is_active": true}).Decode(&v)
	if err != nil {
		return nil, err
	}
	return &v, nil
}

func (r *vpaRepo) FindByUserID(ctx context.Context, userID bson.ObjectID) ([]model.VPA, error) {
	cursor, err := r.col.Find(ctx, bson.M{"user_id": userID, "is_active": true})
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)
	var vpas []model.VPA
	cursor.All(ctx, &vpas)
	return vpas, nil
}

func (r *txnRepo) Create(ctx context.Context, t *model.UPITransaction) error {
	t.CreatedAt = time.Now()
	_, err := r.col.InsertOne(ctx, t)
	return err
}

func (r *txnRepo) FindByUserID(ctx context.Context, userID bson.ObjectID, page, limit int64) ([]model.UPITransaction, int64, error) {
	filter := bson.M{"user_id": userID}
	total, _ := r.col.CountDocuments(ctx, filter)
	opts := options.Find().
		SetSort(bson.D{{Key: "transaction_date", Value: -1}}).
		SetSkip((page - 1) * limit).
		SetLimit(limit)
	cursor, err := r.col.Find(ctx, filter, opts)
	if err != nil {
		return nil, 0, err
	}
	defer cursor.Close(ctx)
	var txns []model.UPITransaction
	cursor.All(ctx, &txns)
	return txns, total, nil
}

func (r *mandateRepo) Create(ctx context.Context, m *model.Mandate) error {
	m.CreatedAt = time.Now()
	m.UpdatedAt = time.Now()
	_, err := r.col.InsertOne(ctx, m)
	return err
}

func (r *mandateRepo) FindByUserID(ctx context.Context, userID bson.ObjectID) ([]model.Mandate, error) {
	cursor, err := r.col.Find(ctx, bson.M{"user_id": userID})
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)
	var mandates []model.Mandate
	cursor.All(ctx, &mandates)
	return mandates, nil
}

func (r *collectRepo) Create(ctx context.Context, c *model.CollectRequest) error {
	c.CreatedAt = time.Now()
	_, err := r.col.InsertOne(ctx, c)
	return err
}

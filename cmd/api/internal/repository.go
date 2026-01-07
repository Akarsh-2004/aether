package order

import (
    "context"
    "github.com/jmoiron/sqlx"
    "github.com/google/uuid"
)

type Repository struct {
    db *sqlx.DB
}

func NewRepository(db *sqlx.DB) *Repository {
    return &Repository{db: db}
}

func (r *Repository) CreateOrder(ctx context.Context, o *Order) error {
    _, err := r.db.ExecContext(ctx,
        `INSERT INTO orders (id, user_id, status, created_at, updated_at) 
         VALUES ($1,$2,$3,NOW(),NOW())`,
        o.ID, o.UserID, o.Status,
    )
    return err
}

func (r *Repository) UpdateStatus(ctx context.Context, orderID uuid.UUID, status Status) error {
    _, err := r.db.ExecContext(ctx,
        `UPDATE orders SET status=$1, updated_at=NOW() WHERE id=$2`,
        status, orderID,
    )
    return err
}

func (r *Repository) GetByID(ctx context.Context, orderID uuid.UUID) (*Order, error) {
    o := &Order{}
    err := r.db.GetContext(ctx, o, "SELECT * FROM orders WHERE id=$1", orderID)
    return o, err
}

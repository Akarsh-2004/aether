package order

import (
    "time"
    "github.com/google/uuid"
)

type Status string

type Order struct {
    ID        uuid.UUID `db:"id"`
    UserID    uuid.UUID `db:"user_id"`
    DriverID  *uuid.UUID `db:"driver_id"` // nullable
    Status    Status    `db:"status"`
    CreatedAt time.Time `db:"created_at"`
    UpdatedAt time.Time `db:"updated_at"`
}

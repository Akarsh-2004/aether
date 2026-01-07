type DriverStatus string

const (
    IDLE     DriverStatus = "IDLE"
    ASSIGNED DriverStatus = "ASSIGNED"
    BUSY     DriverStatus = "BUSY"
)

type Driver struct {
    ID            uuid.UUID   `db:"id"`
    Status        DriverStatus `db:"status"`
    City          string      `db:"city"`
    LastHeartbeat time.Time   `db:"last_heartbeat"`
}

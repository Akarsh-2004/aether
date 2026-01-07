package order

type Status string

const (
	CREATED     Status = "CREATED"
	PAID        Status = "PAID"
	ASSIGNED    Status = "ASSIGNED"
	PICKED_UP  Status = "PICKED_UP"
	IN_TRANSIT Status = "IN_TRANSIT"
	DELIVERED  Status = "DELIVERED"
	CANCELLED  Status = "CANCELLED"
)

var validTransitions = map[Status][]Status{
	CREATED:     {PAID, CANCELLED},
	PAID:        {ASSIGNED, CANCELLED},
	ASSIGNED:    {PICKED_UP},
	PICKED_UP:  {IN_TRANSIT},
	IN_TRANSIT: {DELIVERED},
}

func CanTransition(from, to Status) bool {
	for _, s := range validTransitions[from] {
		if s == to {
			return true
		}
	}
	return false
}

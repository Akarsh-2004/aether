func (s *Service) PayOrder(ctx context.Context, orderID uuid.UUID) error {
	return s.db.WithTx(ctx, func(tx *sql.Tx) error {
		order, err := s.repo.GetForUpdate(tx, orderID)
		if err != nil {
			return err
		}

		if !CanTransition(order.Status, PAID) {
			return errors.New("invalid state transition")
		}

		if err := s.repo.UpdateStatus(tx, orderID, PAID); err != nil {
			return err
		}

		event := event.New("ORDER_PAID", order)
		if err := s.eventRepo.Save(tx, event); err != nil {
			return err
		}

		return nil
	})
}

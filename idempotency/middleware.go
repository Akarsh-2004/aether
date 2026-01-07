func Middleware(store Store) gin.HandlerFunc {
	return func(c *gin.Context) {
		key := c.GetHeader("Idempotency-Key")
		if key == "" {
			c.Next()
			return
		}

		if store.Exists(key) {
			c.AbortWithStatus(http.StatusConflict)
			return
		}

		store.Save(key)
		c.Next()
	}
}

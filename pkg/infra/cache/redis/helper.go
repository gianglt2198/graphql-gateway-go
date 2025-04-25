package credis

// Helper functions
func (c *redisCache) prefixKey(key string) string {
	if c.namespace == "" {
		return key
	}
	return c.namespace + ":" + key
}

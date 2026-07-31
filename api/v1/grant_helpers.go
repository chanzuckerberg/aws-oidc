package v1

// Key uniquely identifies an AWS grant by account and role. The portal round-trips grants
// through form checkbox values of this shape.
func (g AWSGrant) Key() string {
	return g.AccountID + "|" + g.RoleARN
}

// Key returns the grant's identifying key, or empty when no known provider section is set.
func (g Grant) Key() string {
	if g.AWS != nil {
		return g.AWS.Key()
	}
	return ""
}

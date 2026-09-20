package helper

func OwnedByOrg(owner, organizationID *uint64) bool {
	return owner != nil && organizationID != nil && *owner == *organizationID
}

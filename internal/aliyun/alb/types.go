package alb

type AclPolicy struct {
	AclId            string
	AclName          string
	AddressIPVersion string
	ResourceGroupId  string
}

type AclEntry struct {
	Entry            string
	EntryDescription string
}

type ListAclPoliciesRequest struct {
	PageNumber int
	PageSize   int
}

type ListAclPoliciesResponse struct {
	RequestId   string
	TotalCount  int
	PageNumber  int
	PageSize    int
	AclPolicies []AclPolicy
	NextToken   string
}

type ListAclEntriesRequest struct {
	AclId      string
	PageNumber int
	PageSize   int
}

type ListAclEntriesResponse struct {
	RequestId  string
	TotalCount int
	PageNumber int
	PageSize   int
	AclEntries []AclEntry
	NextToken  string
}

type AddAclEntriesRequest struct {
	AclId   string
	Entries []AclEntry
}

type AddAclEntriesResponse struct {
	RequestId string
}

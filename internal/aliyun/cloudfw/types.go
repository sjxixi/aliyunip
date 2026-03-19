package cloudfw

type AddressBook struct {
	AddressBookName string
	AddressBookId   string
	Description     string
	AddressList     []string
	AutoAddTagEcs   bool
	GroupType       string
}

type ListAddressBooksRequest struct {
	PageNumber int
	PageSize   int
}

type ListAddressBooksResponse struct {
	RequestId    string
	TotalCount   int
	PageNumber   int
	PageSize     int
	AddressBooks []AddressBook
}

type AddIpToAddressBookRequest struct {
	AddressBookName string
	AddressBookId   string
	IpList          []string
}

type AddIpToAddressBookResponse struct {
	RequestId string
}

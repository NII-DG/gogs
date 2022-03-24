package ipfs

//自作Mock
type IFIpfsCommandMock struct{}

func (*IFIpfsCommandMock) Run() ([]byte, error) {

	strMsg := `item1
	item2
	item3`
	vec := []byte(strMsg)
	return vec, nil
}

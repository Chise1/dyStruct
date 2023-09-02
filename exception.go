package dyStruct

type Err struct {
	ChainName        error
	SuccessChainName error
}

func (s Err) Error() string {
	return ""
}

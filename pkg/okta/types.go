package okta

type ClientID string

func (c ClientID) String() string {
	return string(c)
}

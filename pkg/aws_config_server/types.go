package aws_config_server

type ClientID string

func (c ClientID) String() string {
	return string(c)
}

package irc

var (
	BucketConfig = []byte("config")
	BucketOpers  = []byte("opers")
)

var Buckets [][]byte = [][]byte{
	BucketConfig,
	BucketOpers,
}

var (
	ConfigPass = []byte("pass")
	ConfigSalt = []byte("salt")
	ConfigCert = []byte("cert")
	ConfigKey  = []byte("key")
)

var (
	OperPass = []byte("pass")
	OperSalt = []byte("salt")
)

var DefaultOper = []byte("irc")

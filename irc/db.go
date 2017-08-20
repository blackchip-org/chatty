package irc

var (
	BucketConfig = []byte("config")
)

var Buckets [][]byte = [][]byte{
	BucketConfig,
}

var (
	ConfigPass = []byte("pass")
	ConfigSalt = []byte("salt")
	ConfigCert = []byte("cert")
	ConfigKey  = []byte("key")
)

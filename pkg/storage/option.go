package storage

type S3FileOption struct {
	Region     string
	Endpoint   string
	Bucket     string
	AK         string
	SK         string
	LocalFile  string
	RemoteFile string
}

package docker

type ClearOption struct{
	Input string
	TagRegx string
	NameRegx string
	Force bool
}
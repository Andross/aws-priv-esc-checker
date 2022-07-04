module github.com/aws-pe-checker-main

go 1.18

replace github.com/aws-pe-checker-lib => ../aws

require github.com/aws-pe-checker-lib v0.0.0-00010101000000-000000000000

require (
	github.com/aws/aws-sdk-go v1.44.40 // indirect
	github.com/jmespath/go-jmespath v0.4.0 // indirect
)

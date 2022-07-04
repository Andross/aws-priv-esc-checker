package main

import (
	"fmt"
	"log"

	"github.com/aws-pe-checker-lib"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/iam"
)

func main() {
	// Set properties of the predefined Logger, including
	// the log entry prefix and a flag to disable printing
	// the time, source file, and line number.
	log.SetPrefix("greetings: ")
	log.SetFlags(0)
	sess, err := session.NewSession(&aws.Config{
		Region: aws.String("us-east-1")},
	)

	// Create a IAM service client.
	svc := iam.New(sess)
	// Request a greeting message.
	// A slice of names.
	// names := []string{"Gladys", "Samantha", "Darrin"}
	ch := make(chan []*iam.Role)
	// Request greeting messages for the names.
	go aws.ListRoles(ch, svc)

	roles := <-ch

	fmt.Println("☑️ list roles")
	for _, idxRole := range roles {

		fmt.Printf("%s\t%s\t%s\t",
			*idxRole.RoleId,
			*idxRole.RoleName,
			*idxRole.Arn)
		if idxRole.Description != nil {
			fmt.Print(*idxRole.Description)
		}
		fmt.Print("\n")
	}
}

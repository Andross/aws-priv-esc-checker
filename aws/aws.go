package aws

import (
	"context"
	"fmt"
	"math/rand"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	_ "github.com/aws/aws-sdk-go/aws/client"
	"github.com/aws/aws-sdk-go/service/iam"
	_ "github.com/aws/aws-sdk-go/service/iam"
)

const (
	// DefaultRetryerMaxNumRetries sets maximum number of retries
	DefaultRetryerMaxNumRetries = 3

	// DefaultRetryerMinRetryDelay sets minimum retry delay
	DefaultRetryerMinRetryDelay = 30 * time.Millisecond

	// DefaultRetryerMinThrottleDelay sets minimum delay when throttled
	DefaultRetryerMinThrottleDelay = 500 * time.Millisecond

	// DefaultRetryerMaxRetryDelay sets maximum retry delay
	DefaultRetryerMaxRetryDelay = 300 * time.Second

	// DefaultRetryerMaxThrottleDelay sets maximum delay when throttled
	DefaultRetryerMaxThrottleDelay = 300 * time.Second
)

// Hello returns a greeting for the named person.
func ListRoles(c chan []*iam.Role, svc &Client) {
	// If no name was given, return an error with a message.

	// cfg, err := config.LoadDefaultConfig(context.TODO(), config.WithRegion("us-east-1"))
	// if err != nil {
	// 	log.Fatalf("failed to load configuration, %v", err)
	// }
	// Initialize a session in us-west-2 that the SDK will use to load
	// credentials from the shared credentials file ~/.aws/credentials.

	roles, err := svc.ListRoles(&iam.ListRolesInput{})

	if err != nil {
		panic("Could not list roles: " + err.Error())
	}

	c <- roles.Roles
}

func createRolePoliciesMap(roles []*iam.Role, policiesMap map[string] svc &Client) {
	// ListRolePolicies

	rolePoliciesList, err := svc.ListRolePolicies(context.Background(), &iam.ListRolePoliciesInput{
		RoleName: aws.String(ExampleRoleName),
	})

	if err != nil {
		panic("Couldn't list policies for role: " + err.Error())
	}

	for _, rolePolicy := range rolePoliciesList.PolicyNames {
		fmt.Printf("Policy ARN: %v", rolePolicy)
	}
}

func checkForPrivEsc() {

}

// init sets initial values for variables used in the function.
func init() {
	rand.Seed(time.Now().UnixNano())
}

// randomFormat returns one of a set of greeting messages. The returned
// message is selected at random.
func randomFormat() string {
	// A slice of message formats.
	formats := []string{
		"Hi, %v. Welcome!",
		"Great to see you, %v!",
		"Hail, %v! Well met!",
	}

	// Return a randomly selected message format by specifying
	// a random index for the slice of formats.
	return formats[rand.Intn(len(formats))]
}

// Return a randomly selected message format by specifying
// a random index for the slice of formats.

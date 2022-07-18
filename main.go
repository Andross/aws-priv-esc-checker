package main

import (
	"flag"
	"fmt"
	"log"
	"sync"
	"time"

	awstools "github.com/Andross/aws-priv-esc-checker/aws-pe-checker-lib"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/iam"
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

func main() {
	// Set properties of the predefined Logger, including
	// the log entry prefix and a flag to disable printing
	// the time, source file, and line number.

	allUsersFlag := flag.Bool("a", false, "Check privilege escalation against all users. Default will only check current user")

	flag.Parse()

	if *allUsersFlag == false {
		fmt.Println("Checking privilege escalation against current user")
	} else {
		fmt.Println("Checking privilege escalation against all available users")
	}

	log.SetPrefix("greetings: ")
	log.SetFlags(0)
	sess, err := session.NewSession(&aws.Config{
		Region: aws.String("us-east-1")},
	)
	if err != nil {
		panic("Could not create session: " + err.Error())
	}
	// Create a IAM service client.
	svc := iam.New(sess)

	// cfg, err := config.LoadDefaultConfig(context.TODO())
	// if err != nil {
	// 	panic("configuration error, " + err.Error())
	// }

	// client := iamv2.NewFromConfig(cfg)

	// workdocsSess := session.Must(session.NewSessionWithOptions(session.Options{
	// 	SharedConfigState: session.SharedConfigEnable,
	// }))

	// workdocsSvc := workdocs.New(workdocsSess)
	// // Request a greeting message.
	// A slice of names.
	// names := []string{"Gladys", "Samantha", "Darrin"}
	// ch := make(chan []*iam.Role)
	// Request greeting messages for the names.

	// usersChannel := make(chan []*iam.User)
	// go awstools.GetUsers(svc, usersChannel)
	// users := <-usersChannel
	// awstools.GetUserGroups(svc, users)

	var wg sync.WaitGroup

	//policyChannel := make(chan *awstools.PolicyDetails, 100)
	userPolicyChannel := make(chan *awstools.UserDetails, 300)
	userInput := &iam.GetUserInput{}
	currentUser, err := svc.GetUser(userInput)

	if err != nil {
		panic("Unable to get user!")
	}
	awstools.CreateUserPolicyMap(svc, userPolicyChannel, *currentUser)
	// fmt.Printf("%v", &wg)
	wg.Add(1)
	go awstools.CheckForPrivEsc(svc, userPolicyChannel, &wg)
	close(userPolicyChannel)
	wg.Wait()
	// policyObjects := make([]*awstools.PolicyObj, 0)
	// policyObjects = awstools.ListAllPolicies(svc, policyObjects)
	// //userDetailsChannel := make(chan []*iam.GetAccountAuthorizationDetailsOutput)
	// for _, polObj := range policyObjects {
	// 	fmt.Printf("Policy arn: %s\n", *polObj.GetArn())
	// 	result, err := awstools.GetPolicyDocument(svc, polObj.GetArn(), polObj.GetVersionid())
	// 	if err != nil {
	// 		fmt.Println("Got an error retrieving the description:")
	// 		fmt.Println(err)
	// 		return
	// 	}
	// 	document, decodeErr := url.QueryUnescape(*result.PolicyVersion.Document)

	// 	if decodeErr != nil {
	// 		fmt.Printf("Decoding error: %s", decodeErr.Error())
	// 	}
	// 	fmt.Printf("%s", document)

	// }
	// roles := <-ch

	// users := <-usersChannel
	// for _, user := range users {
	// 	if user == nil {
	// 		continue
	// 	}
	// 	//fmt.Printf("%d user %s created %v\n", i, *user.UserName, user.CreateDate)

	// }
	// fmt.Println("☑️ list roles")
	// for _, idxRole := range roles {

	// 	fmt.Printf("%s\t%s\t%s\t",
	// 		*idxRole.RoleId,
	// 		*idxRole.RoleName,
	// 		*idxRole.Arn)
	// 	if idxRole.Description != nil {
	// 		fmt.Print(*idxRole.Description)
	// 	}
	// 	fmt.Print("\n")
	// }
	// policiesMap := make(map[string]*iam.ListRolePoliciesOutput)
	// awstools.CreateRolePoliciesMap(roles, policiesMap, svc)
}

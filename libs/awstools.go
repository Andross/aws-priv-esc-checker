package awstools

import (
	"context"
	"fmt"
	"math/rand"
	"net/url"
	"os"
	"sync"
	"time"

	iamv2 "github.com/aws/aws-sdk-go-v2/service/iam"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/iam"
	"github.com/aws/aws-sdk-go/service/iam/iamiface"
	"github.com/aws/aws-sdk-go/service/workdocs"
)

// type AWS_Funcs struct {
// 	svc *Client
// }

type PolicyObj struct {
	arn            *string
	versionId      *string
	policyDocument *string
}

func (policyObj PolicyObj) GetArn() *string {
	return policyObj.arn
}

func (policyObj PolicyObj) GetVersionid() *string {
	return policyObj.versionId
}

func (policyObj PolicyObj) GetPolicyDoc() *string {
	return policyObj.policyDocument
}

type IAMGetPolicyAPI interface {
	GetPolicy(ctx context.Context,
		params *iamv2.GetPolicyInput,
		optFns ...func(*iamv2.Options)) (*iamv2.GetPolicyOutput, error)
}

func GetPolicyDocument(svc *iam.IAM, arn *string, versionId *string) (*iam.GetPolicyVersionOutput, error) {

	policyVersionInput := &iam.GetPolicyVersionInput{
		PolicyArn: arn,
		VersionId: versionId,
	}

	return svc.GetPolicyVersion(policyVersionInput)

}

func GetUsers(svc *iam.IAM, usersChannel chan []*iam.User) {
	result, err := svc.ListUsers(&iam.ListUsersInput{
		MaxItems: aws.Int64(10),
	})

	if err != nil {
		fmt.Println("Error", err)
		return
	}
	usersChannel <- result.Users

}

func GetUserGroups(svc *iam.IAM, users []*iam.User) {
	for _, user := range users {

		userInput := &iam.ListGroupsForUserInput{UserName: user.UserName}

		groups, err := svc.ListGroupsForUser(userInput)

		if err != nil {
			fmt.Printf("Error encountered getting user's %s groups: %s", *user.UserName, err.Error())
		}
		for _, group := range groups.Groups {
			fmt.Printf("User %s is in group %s\n", *user.UserName, *group.GroupName)
		}

	}
}

func CreateUserPolicyMap(svc *iam.IAM, userPolicyChannel chan *UserDetails, currentUserOutput iam.GetUserOutput) {
	user := "User"
	input := &iam.GetAccountAuthorizationDetailsInput{Filter: []*string{&user}}
	resp, err := svc.GetAccountAuthorizationDetails(input)
	if err != nil {
		fmt.Println("Got error getting account details")
		fmt.Println(err.Error())
		os.Exit(1)
	}
	currentUser := currentUserOutput.User
	fmt.Printf("Current User is: %s\n", *currentUser.UserName)
	for _, user := range resp.UserDetailList {

		userName := user.UserName
		policyDetailsList := []PolicyDetails{}

		//fmt.Printf("User ARN %s details: %v\n", *user.Arn, user)

		if len(user.UserPolicyList) > 0 {
			// wg.Add(1)
			policyDetailsList = AddInlinePolicies(svc, user, policyDetailsList)
		}

		if len(user.AttachedManagedPolicies) > 0 {
			// wg.Add(1)
			policyDetailsList = AddAttachedPolicies(svc, user, policyDetailsList)
		}

		if user.GroupList != nil {
			//fmt.Printf("Group List %v", user.GroupList)
			for _, group := range user.GroupList {
				fmt.Printf("User %s is in group: %v\n", *user.UserName, *group)
				req := &iam.ListGroupPoliciesInput{GroupName: aws.String(*group)}
				resp, err := svc.ListGroupPolicies(req)
				if err != nil {
					fmt.Printf("%s", err.Error())
				}
				if len(resp.PolicyNames) > 0 {
					fmt.Printf("Group policy is %s", *resp.PolicyNames[0])
				}

				input := &iam.ListAttachedGroupPoliciesInput{GroupName: aws.String(*group)}
				attachedGroupPolicies, err := svc.ListAttachedGroupPolicies(input)

				if len(attachedGroupPolicies.AttachedPolicies) > 0 {
					// wg.Add(1)
					policyDetailsList = AddGroupPolicies(svc, user, policyDetailsList, attachedGroupPolicies)
				}

			}
		}
		fmt.Printf("User Policy Detail List %v\n", policyDetailsList)
		userDetail := UserDetails{userName, policyDetailsList}
		userPolicyChannel <- &userDetail
	}

	// fmt.Printf("%s\n", resp)
}
func AddInlinePolicies(svc *iam.IAM, user *iam.UserDetail, policyDetailsList []PolicyDetails) []PolicyDetails {
	fmt.Printf("Policy Details List %v:\n", policyDetailsList)
	for _, userPolicy := range user.UserPolicyList {
		// wg.Add(1)
		// userPolicyDoc, err := *userPolicy.PolicyDocument
		decodedUserPolicy, err := url.QueryUnescape(*userPolicy.PolicyDocument)
		if err != nil {
			fmt.Printf("Error decoding inline user policy %s", err.Error())
		}
		//fmt.Printf("Adding inline user to user %s policy\n: %s\n", *user.UserName, decodedUserPolicy)
		policyName := "Inline policy"
		policyType := "inline"
		policyArn := ""
		policyDetail := PolicyDetails{&policyName, &policyType, &policyArn, &decodedUserPolicy}
		fmt.Printf("Inline Policy List %v:\n", *policyDetail.PolicyName)
		policyDetailsList = append(policyDetailsList, policyDetail)
		fmt.Printf("Inline Policy List %v:\n", policyDetailsList)
	}
	return policyDetailsList
}

func AddGroupPolicies(svc *iam.IAM, user *iam.UserDetail, policyDetailsList []PolicyDetails, attachedGroupPolicies *iam.ListAttachedGroupPoliciesOutput) []PolicyDetails {
	for _, attachPol := range attachedGroupPolicies.AttachedPolicies {
		//fmt.Printf("Attached Policy Name: %s\n", *attachPol.PolicyArn)
		policyInput := &iam.GetPolicyInput{PolicyArn: attachPol.PolicyArn}
		policy, err := svc.GetPolicy(policyInput)

		if err != nil {
			fmt.Printf("Error getting policy %s: ", err.Error())
			continue
		}

		policyVersionId := *&policy.Policy.DefaultVersionId
		policyArn := *&policy.Policy.Arn
		policyName := *&policy.Policy.PolicyName

		policyDocument, err := GetPolicyDocument(svc, policyArn, policyVersionId)
		if err != nil {
			fmt.Printf("Issue getting policy document %s with version %s\n", err.Error(), *policyVersionId)
			continue
		}
		document, decodeErr := url.QueryUnescape(*policyDocument.PolicyVersion.Document)
		if decodeErr != nil {
			fmt.Printf("Decoding error: %s", decodeErr.Error())
		}
		//fmt.Printf("Policy document %s\n", document)
		policyType := "group"
		policyDetail := PolicyDetails{policyName, &policyType, policyArn, &document}
		policyDetailsList = append(policyDetailsList, policyDetail)

	}
	return policyDetailsList
}

func AddAttachedPolicies(svc *iam.IAM, user *iam.UserDetail, policyDetailsList []PolicyDetails) []PolicyDetails {
	for _, attachedPolicy := range user.AttachedManagedPolicies {
		//fmt.Printf("User %s has attached managed policy:\n%s\n", *user.UserName, attachedPolicy)
		policyInput := &iam.GetPolicyInput{PolicyArn: attachedPolicy.PolicyArn}
		policyName := attachedPolicy.PolicyName
		policyArn := attachedPolicy.PolicyArn
		policyType := "AttachedManaged"
		policyDoc, err := svc.GetPolicy(policyInput)

		if err != nil {
			fmt.Printf("Error getting attached policy arn for user %s", *user.UserName)
		} else {
			policyVersion := policyDoc.Policy.DefaultVersionId
			policyDocument, err := GetPolicyDocument(svc, policyArn, policyVersion)
			if err != nil {
				fmt.Printf("Error getting attached policy for user %s", *user.UserName)
			} else {

				decodedManagedPolicy, err := url.QueryUnescape(*policyDocument.PolicyVersion.Document)
				if err != nil {
					fmt.Printf("Error decoding inline user policy %s", err.Error())
				}

				policyDetail := PolicyDetails{policyName, &policyType, policyArn, &decodedManagedPolicy}
				policyDetailsList = append(policyDetailsList, policyDetail)
			}
		}

	}
	return policyDetailsList
}

func CheckForPrivEsc(svc *iam.IAM, userPolicyChannel chan *UserDetails, wg *sync.WaitGroup) {

	fmt.Printf("Channel length %v\n", len(userPolicyChannel))
	for userDetail := range userPolicyChannel {

		//fmt.Printf("User info %v\n", userDetail.Policies)
		for _, policy := range userDetail.Policies {
			fmt.Printf("Policy Name: %s\nPolicy ARN: %s\nPolicy Type: %s\nPolicy: %s\n", *policy.PolicyName, *policy.PolicyArn, *policy.PolicyType, *policy.Policy)

		}
		// defer wg.Done()
		// fmt.Printf("Policy: %v", policyDetail1)
	}
	defer wg.Done()
}

func ListUsers(svc *workdocs.WorkDocs, orgId string, requestedUser string) {

	// if orgId == "" {
	// 	fmt.Println("You must supply the organization ID")
	// 	flag.PrintDefaults()
	// 	os.Exit(1)
	// }

	input := new(workdocs.DescribeUsersInput)
	input.OrganizationId = &orgId

	// Show all users if we don't get a user name
	if requestedUser == "" {
		fmt.Println("Getting info about all users")
	} else {
		fmt.Println("Getting info about user " + requestedUser)
		input.Query = &requestedUser
	}

	result, err := svc.DescribeUsers(input)
	if err != nil {
		fmt.Println("Error getting user info", err)
		return
	}

	if requestedUser == "" {
		fmt.Println("Found", *result.TotalNumberOfUsers, "users")
		fmt.Println("")
	}

	for _, user := range result.Users {
		fmt.Println("Username:   " + *user.Username)

		if requestedUser != "" {
			fmt.Println("Firstname:  " + *user.GivenName)
			fmt.Println("Lastname:   " + *user.Surname)
			fmt.Println("Email:      " + *user.EmailAddress)
			fmt.Println("Root folder " + *user.RootFolderId)
		}

		fmt.Println("")
	}
}

// Hello returns a greeting for the named person.
func ListRoles(c chan []*iam.Role, svc *iam.IAM) {
	// If no name was given, return an error with a message.

	// cfg, err := config.LoadDefaultConfig(context.TODO(), config.WithRegion("us-east-1"))
	// if err != nil {
	// 	log.Fatalf("failed to load configuration, %v", err)
	// }
	// Initialize a session in us-west-2 that the SDK will use to load
	// credentials from the shared credentials file ~/.aws/credentials.

	roles, err := svc.ListRoles(&iam.ListRolesInput{})

	if err != nil {
		fmt.Println("Could not list roles: " + err.Error())
	}

	c <- roles.Roles
}

func ListAllPolicies(svc *iam.IAM, policyObjects []*PolicyObj) []*PolicyObj {
	policies, err := svc.ListPolicies(&iam.ListPoliciesInput{OnlyAttached: aws.Bool(true)})

	if err != nil {
		fmt.Println("Could not list roles: " + err.Error())
	}

	for _, policy := range policies.Policies {
		//fmt.Printf("ARN of attached policy %s\n", *policy)

		policyDoc := ""
		policyArn := policy.Arn
		policyVersion := policy.DefaultVersionId
		policyObj := PolicyObj{policyArn, policyVersion, &policyDoc}
		policyObjects = append(policyObjects, &policyObj)
	}

	return policyObjects
}

func UserPolicyHasAdmin(user *iam.UserDetail, admin string) bool {
	for _, policy := range user.UserPolicyList {
		if *policy.PolicyName == admin {
			return true
		}
	}

	return false
}

func CreateRolePoliciesMap(roles []*iam.Role, policiesMap map[string]*iam.ListRolePoliciesOutput, svc *iam.IAM) {
	// ListRolePolicies
	_ = policiesMap
	// fmt.Printf("%v", roles)
	for _, idxRole := range roles {
		fmt.Printf("%v\n", *idxRole.Arn)
		arn := "arn:aws:iam::aws:policy/AdministratorAccess"
		description, err := GetPolicyDescription(svc, &arn)
		if err != nil {
			fmt.Printf("%s", err.Error())
		}

		fmt.Printf(description)
		// rolePoliciesList, err := svc.GetPolicy(&iam.GetPolicyInput{
		// 	PolicyArn: aws.String(*idxRole.Arn),
		// })
		// if err != nil {
		// 	fmt.Println("Couldn't list policies for role: " + err.Error())
		// }

		// for _, rolePolicy := range *rolePoliciesList.Policy.Description {
		// 	fmt.Printf("Policy ARN: %v", rolePolicy)
		// }
	}

}

func GetPolicyDescription(svc iamiface.IAMAPI, arn *string) (string, error) {
	result, err := svc.GetPolicy(&iam.GetPolicyInput{
		PolicyArn: arn,
	})
	if err != nil {
		return "", err
	}

	if result.Policy == nil {
		return "Policy nil", nil
	}

	if result.Policy.Description != nil {
		return *result.Policy.Description, nil
	}

	return "Description nil", nil
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

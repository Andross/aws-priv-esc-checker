package awstools

type UserDetails struct {
	Username *string `json:"username"`
	Policies []PolicyDetails
}

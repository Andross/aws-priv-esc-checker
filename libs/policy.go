package awstools

type PolicyDetails struct {
	PolicyName *string `json:"policyname"`
	PolicyType *string `json:"policyType"`
	PolicyArn  *string `json:"arn"`
	Policy     *string `json:"policy"`
}

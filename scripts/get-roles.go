package main

import (
	"context"
	"fmt"
	"os"

	"github.com/aws/aws-sdk-go/aws/arn"
	"github.com/aws/aws-sdk-go/service/iam"
	"github.com/hashicorp/go-tfe"
)

type accountAndRole struct {
	AccountId string

	RoleARN *arn.ARN
	Role    *iam.Role
}

type OutputMapping struct {
	clientIdMapping map[string][]accountAndRole
}

func main() {
	// Replace with your Terraform Cloud/Enterprise token
	token := os.Getenv("TFE_TOKEN")
	if token == "" {
		fmt.Println("TFE_TOKEN environment variable must be set.")
		os.Exit(1)
	}

	// Replace with the ID of your workspace
	workspaceID := "ws-DYLTn9hNMGMwnGzN"

	config := &tfe.Config{
		Token:   token,
		Address: "https://si.prod.tfe.czi.technology",
	}

	client, err := tfe.NewClient(config)
	if err != nil {
		fmt.Printf("Error creating TFE client: %v\n", err)
		os.Exit(1)
	}

	ctx := context.Background()

	currentState, _ := client.StateVersionOutputs.ReadCurrent(ctx, workspaceID)
	fmt.Printf("Current State: %v\n", currentState.Items)

	OutputMapping := OutputMapping{}

	// get account-id
	accountId := ""
	for _, output := range currentState.Items {
		if output.Name == "account-id" {
			accountId = output.Value
			break
		}
	}

	// okta-czi-prod
	for _, output := range currentState.Items {
		if output.Name == "okta-czi-prod" {
			OutputMapping.clientIdMapping[output.Name] = []accountAndRole{
				{
					AccountId: accountId,
					RoleARN:   nil,
					Role:      "okta-czi-prod",
				},
			}
		}
	}

	// fmt.Println("Workspace Outputs:")
	// for _, output := range currentState.Items {
	// 	OutputMapping.clientIdMapping[output.Name] = []accountAndRole{
	// 		{
	// 			AccountId: "",
	// 			RoleARN:   nil,
	// 			Role: ,
	// 		}
	// 	fmt.Printf("Key: %s, Value: %s\n", output.Name, output.Value)
	// }
}

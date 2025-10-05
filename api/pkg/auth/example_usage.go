package auth

import (
	"context"
	"fmt"
	"time"
)

// ExampleUsage demonstrates how to use the JWKS validator
func ExampleUsage() {
	// Configuration example
	config := AuthConfig{
		JWKSURL:            "https://your-instance.zitadel.cloud/oauth/v2/keys",
		Issuer:             "https://your-instance.zitadel.cloud",
		Audience:           "your_client_id@your_project",
		ProjectID:          "your_project_id",
		OrgID:              "your_org_id",
		RoleClaimName:      "urn:zitadel:iam:org:project:roles",
		CacheTTL:           time.Hour,
		ClockSkewTolerance: 2 * time.Minute,
		HTTPTimeout:        30 * time.Second,
	}

	// Create validator
	validator := NewJWKSValidatorFromConfig(config)

	// Example token (this would be a real JWT token in practice)
	tokenString := "eyJhbGciOiJSUzI1NiIsInR5cCI6IkpXVCIsImtpZCI6InlvdXJfa2V5X2lkIn0..."

	// Validate token
	ctx := context.Background()
	claims, err := validator.ValidateToken(ctx, tokenString)
	if err != nil {
		fmt.Printf("Token validation failed: %v\n", err)
		return
	}

	fmt.Printf("Token validated successfully!\n")
	fmt.Printf("Subject: %s\n", claims.Sub)
	fmt.Printf("Email: %s\n", claims.Email)
	fmt.Printf("Name: %s\n", claims.Name)
	fmt.Printf("Roles: %v\n", claims.Roles)
	fmt.Printf("Organization ID: %s\n", claims.OrgID)
	fmt.Printf("Project ID: %s\n", claims.ProjectID)

	// Extract user info
	userInfo, err := ExtractUserInfoFromToken(validator, tokenString)
	if err != nil {
		fmt.Printf("Failed to extract user info: %v\n", err)
		return
	}

	fmt.Printf("User Info - Name: %s, Email: %s, Roles: %v\n",
		userInfo.Name, userInfo.Email, userInfo.Roles)

	// Validate with additional context
	contextClaims, err := ValidateTokenWithContext(validator, tokenString, config.OrgID, config.ProjectID)
	if err != nil {
		fmt.Printf("Context validation failed: %v\n", err)
		return
	}

	fmt.Printf("Context validation successful for org %s and project %s\n",
		contextClaims.OrgID, contextClaims.ProjectID)
}

// ExampleCustomRoleClaim demonstrates using a custom role claim name
func ExampleCustomRoleClaim() {
	config := AuthConfig{
		JWKSURL:            "https://your-instance.zitadel.cloud/oauth/v2/keys",
		Issuer:             "https://your-instance.zitadel.cloud",
		Audience:           "your_client_id",
		RoleClaimName:      "custom_roles", // Custom role claim name
		CacheTTL:           time.Hour,
		ClockSkewTolerance: 2 * time.Minute,
		HTTPTimeout:        30 * time.Second,
	}

	validator := NewJWKSValidatorFromConfig(config)

	// The validator will now look for roles in the "custom_roles" claim
	// instead of the default Zitadel claim name
	fmt.Println("Validator configured with custom role claim:", config.RoleClaimName)
	_ = validator // Suppress unused variable warning
}

// ExampleAudienceValidation demonstrates audience validation with fallback
func ExampleAudienceValidation() {
	// Example 1: client_id@project format
	config1 := AuthConfig{
		Audience:  "myclient@myproject",
		ProjectID: "myproject",
	}

	// This will accept both "myclient@myproject" and "myclient" as valid audiences
	validator1 := NewJWKSValidatorFromConfig(config1)
	fmt.Printf("Validator 1 accepts audiences: myclient@myproject, myclient\n")

	// Example 2: client_id only format
	config2 := AuthConfig{
		Audience: "myclient",
	}

	// This will only accept "myclient" as a valid audience
	validator2 := NewJWKSValidatorFromConfig(config2)
	fmt.Printf("Validator 2 accepts audience: myclient\n")

	_ = validator1
	_ = validator2
}

package connector

import (
	"errors"

	v2 "github.com/conductorone/baton-sdk/pb/c1/connector/v2"
	"github.com/conductorone/baton-sdk/pkg/crypto"
)

func generateCredentials(credentialOptions *v2.CredentialOptions) (string, error) {
	if credentialOptions.GetRandomPassword() == nil {
		return "", errors.New("unsupported credential option")
	}

	password, err := crypto.GenerateRandomPassword(
		&v2.CredentialOptions_RandomPassword{
			Length: min(12, credentialOptions.GetRandomPassword().GetLength()),
		},
	)
	if err != nil {
		return "", err
	}
	return password, nil
}

package tests

import (
	"app/tests/suite"
	"testing"

	"github.com/golang-jwt/jwt/v5"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	sso "github.com/underground20/sso-grpc-contract/generated"
)

const (
	appId  = 1
	appKey = "test-secret"
)

func TestLoginSuccess(t *testing.T) {
	ctx, suite := suite.New(t)

	suite.Connection.Exec(ctx, `INSERT INTO apps (name, secret) VALUES ($1, $2)`, "test", appKey)

	registerResp, err := suite.AuthClient.Register(ctx, &sso.RegisterRequest{
		Email:    "test@mail.com",
		Password: "password",
	})

	require.NoError(t, err)
	assert.NotEmpty(t, registerResp.GetUserId())

	resp, err := suite.AuthClient.Login(ctx, &sso.LoginRequest{
		Email:    "test@mail.com",
		Password: "password",
		AppId:    appId,
	})

	require.NoError(t, err)

	token := resp.GetToken()
	tokenParsed, err := jwt.Parse(token, func(token *jwt.Token) (interface{}, error) {
		return []byte(appKey), nil
	})

	require.NoError(t, err)

	_, ok := tokenParsed.Claims.(jwt.MapClaims)
	require.True(t, ok)

	suite.Connection.Exec(ctx, `DELETE FROM users WHERE email = $1`, "test@mail.com")
}

func TestRegisterFail(t *testing.T) {
	ctx, st := suite.New(t)

	tests := []struct {
		name        string
		email       string
		password    string
		expectedErr string
	}{
		{
			name:        "Register with Empty Password",
			email:       "test@mail.com",
			password:    "",
			expectedErr: "password is required",
		},
		{
			name:        "Register with Empty Email",
			email:       "",
			password:    "password",
			expectedErr: "email is required",
		},
		{
			name:        "Register with Both Empty",
			email:       "",
			password:    "",
			expectedErr: "email is required",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := st.AuthClient.Register(ctx, &sso.RegisterRequest{
				Email:    tt.email,
				Password: tt.password,
			})
			require.Error(t, err)
			require.Contains(t, err.Error(), tt.expectedErr)

		})
	}
}

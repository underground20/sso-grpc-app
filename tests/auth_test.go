package tests

import (
	"app/tests/suite"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	sso "github.com/underground20/sso-grpc-contract/generated"
	"github.com/underground20/sso-jwt-token/pkg/jwt/user"
)

const (
	secret = "test-secret"
)

func TestRegisterSuccess(t *testing.T) {
	ctx, suite := suite.New(t)
	suite.Cleanup(ctx)

	registerResp, err := suite.AuthClient.Register(ctx, &sso.RegisterRequest{
		Email:    "test@mail.com",
		Password: "password",
	})

	require.NoError(t, err)
	assert.NotEmpty(t, registerResp.GetUserId())

	user, _ := suite.UserStorage.GetUser(ctx, "test@mail.com")
	assert.WithinDuration(t, time.Now(), user.CreatedAt, 10*time.Second)
	assert.Nil(t, user.LastLogin)
}

func TestRegisterWhenUserAlreadyRegistered(t *testing.T) {
	ctx, suite := suite.New(t)

	suite.CreateUser(ctx, "test@mail.com", "password")
	suite.Cleanup(ctx)

	_, err := suite.AuthClient.Register(ctx, &sso.RegisterRequest{
		Email:    "test@mail.com",
		Password: "password",
	})

	require.EqualError(t, err, "rpc error: code = AlreadyExists desc = user already registered")
}

func TestRegisterFail(t *testing.T) {
	ctx, st := suite.New(t)
	st.Parallel()

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

func TestLoginSuccess(t *testing.T) {
	ctx, suite := suite.New(t)

	appId, _ := suite.AppStorage.RegisterApp(ctx, "test", secret)
	uuid := suite.CreateUser(ctx, "test@mail.com", "password")
	suite.Cleanup(ctx)

	resp, err := suite.AuthClient.Login(ctx, &sso.LoginRequest{
		Email:    "test@mail.com",
		Password: "password",
		AppId:    int64(appId),
	})

	require.NoError(t, err)

	tokenParser, _ := user.NewTokenGenerator(time.Hour)
	claims, err := tokenParser.Parse(resp.GetToken(), secret)
	require.NoError(t, err)
	require.Contains(t, claims.Subject, uuid)
}

func TestLoginWithIncorrectPassword(t *testing.T) {
	ctx, suite := suite.New(t)

	appId, _ := suite.AppStorage.RegisterApp(ctx, "test", secret)
	suite.Cleanup(ctx)

	registerResp, err := suite.AuthClient.Register(ctx, &sso.RegisterRequest{
		Email:    "test@mail.com",
		Password: "password",
	})

	require.NoError(t, err)
	assert.NotEmpty(t, registerResp.GetUserId())

	_, err = suite.AuthClient.Login(ctx, &sso.LoginRequest{
		Email:    "test@mail.com",
		Password: "12345",
		AppId:    int64(appId),
	})

	require.EqualError(t, err, "rpc error: code = InvalidArgument desc = invalid email or password")
}

func TestGetEmptyRolesList(t *testing.T) {
	ctx, suite := suite.New(t)

	rolesListResp, err := suite.AuthClient.GetRoles(ctx, &sso.GetRolesRequest{})

	require.NoError(t, err)
	assert.Empty(t, rolesListResp.GetRoles())
}

func TestGetRolesList(t *testing.T) {
	ctx, suite := suite.New(t)

	suite.RoleStorage.CreateRole(ctx, "admin", []string{"read", "write"})
	suite.Cleanup(ctx)

	rolesListResp, err := suite.AuthClient.GetRoles(ctx, &sso.GetRolesRequest{})

	require.NoError(t, err)
	assert.NotEmpty(t, rolesListResp.GetRoles())

	expectedRoles := []*sso.Role{
		{
			Name:        "admin",
			Permissions: []string{"read", "write"},
		},
	}
	actualRoles := rolesListResp.GetRoles()

	require.Len(t, actualRoles, len(expectedRoles))
	for i, expectedRole := range expectedRoles {
		actualRole := actualRoles[i]
		assert.Equal(t, expectedRole.Name, actualRole.Name)
		assert.ElementsMatch(t, expectedRole.Permissions, actualRole.Permissions)
	}
}

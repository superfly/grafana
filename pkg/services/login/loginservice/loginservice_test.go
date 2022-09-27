package loginservice

import (
	"bytes"
	"context"
	"errors"
	"testing"
	"time"

	"github.com/go-kit/log"
	"github.com/go-kit/log/level"
	"github.com/grafana/grafana/pkg/models"
	"github.com/grafana/grafana/pkg/services/login"
	"github.com/grafana/grafana/pkg/services/login/logintest"
	"github.com/grafana/grafana/pkg/services/org"
	"github.com/grafana/grafana/pkg/services/org/orgtest"
	"github.com/grafana/grafana/pkg/services/quota/quotaimpl"
	"github.com/grafana/grafana/pkg/services/sqlstore/mockstore"
	"github.com/grafana/grafana/pkg/services/user"
	"github.com/grafana/grafana/pkg/services/user/usertest"
	"github.com/grafana/grafana/pkg/setting"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

func Test_syncOrgRoles_doesNotBreakWhenTryingToRemoveLastOrgAdmin(t *testing.T) {
	user := createSimpleUser()
	externalUser := createSimpleExternalUser()
	authInfoMock := &logintest.AuthInfoServiceFake{}

	store := &mockstore.SQLStoreMock{
		ExpectedUserOrgList:     createUserOrgDTO(),
		ExpectedOrgListResponse: createResponseWithOneErrLastOrgAdminItem(),
	}

	login := Implementation{
		QuotaService:    &quotaimpl.Service{},
		AuthInfoService: authInfoMock,
		SQLStore:        store,
		userService:     usertest.NewUserServiceFake(),
		orgService:      orgtest.NewOrgServiceFake(),
	}

	err := login.syncOrgRoles(context.Background(), &user, &externalUser)
	require.NoError(t, err)
}

func Test_syncOrgRoles_whenTryingToRemoveLastOrgLogsError(t *testing.T) {
	buf := &bytes.Buffer{}
	logger.Swap(level.NewFilter(log.NewLogfmtLogger(buf), level.AllowInfo()))

	user := createSimpleUser()
	externalUser := createSimpleExternalUser()

	authInfoMock := &logintest.AuthInfoServiceFake{}

	store := &mockstore.SQLStoreMock{
		ExpectedUserOrgList:     createUserOrgDTO(),
		ExpectedOrgListResponse: createResponseWithOneErrLastOrgAdminItem(),
	}

	orgService := orgtest.NewOrgServiceFake()
	orgService.ExpectedError = models.ErrLastOrgAdmin

	login := Implementation{
		QuotaService:    &quotaimpl.Service{},
		AuthInfoService: authInfoMock,
		SQLStore:        store,
		userService:     usertest.NewUserServiceFake(),
		orgService:      orgService,
	}

	err := login.syncOrgRoles(context.Background(), &user, &externalUser)
	require.NoError(t, err)
	assert.Contains(t, buf.String(), models.ErrLastOrgAdmin.Error())
}

func Test_teamSync(t *testing.T) {
	authInfoMock := &logintest.AuthInfoServiceFake{}
	login := Implementation{
		QuotaService:    &quotaimpl.Service{},
		AuthInfoService: authInfoMock,
	}

	email := "test_user@example.org"
	upserCmd := &models.UpsertUserCommand{ExternalUser: &models.ExternalUserInfo{Email: email},
		UserLookupParams: models.UserLookupParams{Email: &email}}
	expectedUser := &user.User{
		ID:    1,
		Email: email,
		Name:  "test_user",
		Login: "test_user",
	}
	authInfoMock.ExpectedUser = expectedUser

	var actualUser *user.User
	var actualExternalUser *models.ExternalUserInfo

	t.Run("login.TeamSync should not be called when  nil", func(t *testing.T) {
		err := login.UpsertUser(context.Background(), upserCmd)
		require.Nil(t, err)
		assert.Nil(t, actualUser)
		assert.Nil(t, actualExternalUser)

		t.Run("login.TeamSync should be called when not nil", func(t *testing.T) {
			teamSyncFunc := func(user *user.User, externalUser *models.ExternalUserInfo) error {
				actualUser = user
				actualExternalUser = externalUser
				return nil
			}
			login.TeamSync = teamSyncFunc
			err := login.UpsertUser(context.Background(), upserCmd)
			require.Nil(t, err)
			assert.Equal(t, actualUser, expectedUser)
			assert.Equal(t, actualExternalUser, upserCmd.ExternalUser)
		})

		t.Run("login.TeamSync should propagate its errors to the caller", func(t *testing.T) {
			teamSyncFunc := func(user *user.User, externalUser *models.ExternalUserInfo) error {
				return errors.New("teamsync test error")
			}
			login.TeamSync = teamSyncFunc
			err := login.UpsertUser(context.Background(), upserCmd)
			require.Error(t, err)
		})
	})
}

func Test_UpsertUser(t *testing.T) {
	extUsr := &models.ExternalUserInfo{
		AuthModule: "auth.saml",
		Name:       "Test User",
		Login:      "user@samltest.id",
		Email:      "user@samltest.id",
		Groups:     []string{"viewer"},
		OrgRoles:   map[int64]org.RoleType{},
		AuthId:     "123",
		UserId:     0,
	}

	usr := &user.User{
		ID:               1,
		Version:          1,
		Email:            "user@samltest.id",
		Name:             "Test User",
		Login:            "user@samltest.id",
		Password:         "",
		Salt:             "sallt",
		Rands:            "rands",
		Company:          "",
		EmailVerified:    false,
		Theme:            "",
		HelpFlags1:       1,
		IsDisabled:       false,
		IsAdmin:          false,
		IsServiceAccount: false,
		OrgID:            1,
		Created:          time.Time{},
		Updated:          time.Time{},
		LastSeenAt:       time.Time{},
	}

	store := mockstore.NewSQLStoreMock()
	authInfoMock := &AuthInfoServiceMock{}
	userServiceMock := NewUserServiceMock()
	login := Implementation{
		SQLStore: store,
		QuotaService: &quotaimpl.Service{
			Cfg: &setting.Cfg{
				Quota: setting.QuotaSettings{
					Enabled: false,
				},
			},
		},
		userService:     userServiceMock,
		AuthInfoService: authInfoMock,
	}

	upsertCmd := &models.UpsertUserCommand{
		ExternalUser:  extUsr,
		SignupAllowed: true,
		UserLookupParams: models.UserLookupParams{
			UserID: nil,
			Email:  &extUsr.Email,
			Login:  &extUsr.Login,
		},
	}

	t.Run("UpsertUser should create a new user when external user is not found", func(t *testing.T) {
		authInfoMock.ExpectedErrorByMethod = map[string]error{
			"LookupAndUpdate": user.ErrUserNotFound,
			"SetAuthInfo":     nil,
		}
		userServiceMock.ExpectedUserByMethod = map[string]*user.User{
			"Create": usr,
		}

		authInfoMock.On("SetAuthInfo", mock.Anything, &models.SetAuthInfoCommand{
			AuthModule: "auth.saml",
			AuthId:     extUsr.AuthId,
			UserId:     1,
			OAuthToken: nil}).Return(nil)
		authInfoMock.On("LookupAndUpdate", mock.Anything, mock.Anything).Return(nil, user.ErrUserNotFound)
		userServiceMock.On("Create", mock.Anything, mock.Anything).Return(usr, nil)

		err := login.UpsertUser(context.Background(), upsertCmd)

		require.Nil(t, err)
		assert.Equal(t, usr, upsertCmd.Result)

		authInfoMock.AssertExpectations(t)
	})

	t.Run("UpsertUser should find the existing user when the external user exists", func(t *testing.T) {
		authInfoMock.ExpectedUserByMethod = map[string]*user.User{
			"LookupAndUpdate": usr,
		}
		authInfoMock.ExpectedErrorByMethod = map[string]error{
			"SetAuthInfo": nil,
		}

		authInfoMock.On("SetAuthInfo", mock.Anything, &models.SetAuthInfoCommand{
			AuthModule: "auth.saml",
			AuthId:     extUsr.AuthId,
			UserId:     1,
			OAuthToken: nil}).Return(nil)
		authInfoMock.On("LookupAndUpdate", mock.Anything, mock.Anything).Return(usr, nil)

		err := login.UpsertUser(context.Background(), upsertCmd)

		require.Nil(t, err)
		assert.Equal(t, usr, upsertCmd.Result)
		authInfoMock.AssertExpectations(t)
		userServiceMock.AssertNotCalled(t, "Create")
	})
}

func createSimpleUser() user.User {
	user := user.User{
		ID: 1,
	}

	return user
}

func createUserOrgDTO() []*models.UserOrgDTO {
	users := []*models.UserOrgDTO{
		{
			OrgId: 1,
			Name:  "Bar",
			Role:  org.RoleViewer,
		},
		{
			OrgId: 10,
			Name:  "Foo",
			Role:  org.RoleAdmin,
		},
		{
			OrgId: 11,
			Name:  "Stuff",
			Role:  org.RoleViewer,
		},
	}
	return users
}

func createSimpleExternalUser() models.ExternalUserInfo {
	externalUser := models.ExternalUserInfo{
		AuthModule: login.LDAPAuthModule,
		OrgRoles: map[int64]org.RoleType{
			1: org.RoleViewer,
		},
	}

	return externalUser
}

func createResponseWithOneErrLastOrgAdminItem() mockstore.OrgListResponse {
	remResp := mockstore.OrgListResponse{
		{
			OrgId:    10,
			Response: models.ErrLastOrgAdmin,
		},
		{
			OrgId:    11,
			Response: nil,
		},
	}
	return remResp
}

type AuthInfoServiceMock struct {
	mock.Mock
	ExpectedUserByMethod  map[string]*user.User
	ExpectedExternalUser  *models.ExternalUserInfo
	ExpectedErrorByMethod map[string]error
}

func (a *AuthInfoServiceMock) LookupAndUpdate(ctx context.Context, query *models.GetUserByAuthInfoQuery) (*user.User, error) {
	a.Called(ctx, query)
	if val, ok := a.ExpectedErrorByMethod["LookupAndUpdate"]; ok {
		return nil, val
	}
	return a.ExpectedUserByMethod["LookupAndUpdate"], nil
}

func (a *AuthInfoServiceMock) GetAuthInfo(ctx context.Context, query *models.GetAuthInfoQuery) error {
	return a.ExpectedErrorByMethod["GetAuthInfo"]
}

func (a *AuthInfoServiceMock) SetAuthInfo(ctx context.Context, cmd *models.SetAuthInfoCommand) error {
	a.Called(ctx, cmd)
	return a.ExpectedErrorByMethod["SetAuthInfo"]
}

func (a *AuthInfoServiceMock) UpdateAuthInfo(ctx context.Context, cmd *models.UpdateAuthInfoCommand) error {
	return a.ExpectedErrorByMethod["UpdateAuthInfo"]
}

func (a *AuthInfoServiceMock) GetExternalUserInfoByLogin(ctx context.Context, query *models.GetExternalUserInfoByLoginQuery) error {
	query.Result = a.ExpectedExternalUser
	return a.ExpectedErrorByMethod["GetExternalUserInfoByLogin"]
}

type UserServiceMock struct {
	mock.Mock
	ExpectedUserByMethod  map[string]*user.User
	ExpectedExternalUser  *models.ExternalUserInfo
	ExpectedErrorByMethod map[string]error
}

func NewUserServiceMock() *UserServiceMock {
	return &UserServiceMock{}
}

func (u *UserServiceMock) Create(ctx context.Context, cmd *user.CreateUserCommand) (*user.User, error) {
	u.Called(ctx, cmd)
	if val, ok := u.ExpectedErrorByMethod["Create"]; ok {
		return nil, val
	}
	return u.ExpectedUserByMethod["Create"], nil
}

func (u *UserServiceMock) Delete(ctx context.Context, cmd *user.DeleteUserCommand) error {
	panic("not implemented")
}

func (u *UserServiceMock) GetByID(ctx context.Context, query *user.GetUserByIDQuery) (*user.User, error) {
	panic("not implemented")
}

func (u *UserServiceMock) GetByLogin(ctx context.Context, query *user.GetUserByLoginQuery) (*user.User, error) {
	panic("not implemented")
}

func (u *UserServiceMock) GetByEmail(ctx context.Context, query *user.GetUserByEmailQuery) (*user.User, error) {
	panic("not implemented")
}

func (u *UserServiceMock) Update(ctx context.Context, cmd *user.UpdateUserCommand) error {
	u.Called(ctx, cmd)
	if val, ok := u.ExpectedErrorByMethod["Update"]; ok {
		return val
	}
	return nil
}

func (u *UserServiceMock) ChangePassword(ctx context.Context, cmd *user.ChangeUserPasswordCommand) error {
	panic("not implemented")
}

func (u *UserServiceMock) UpdateLastSeenAt(ctx context.Context, cmd *user.UpdateUserLastSeenAtCommand) error {
	panic("not implemented")
}

func (u *UserServiceMock) SetUsingOrg(ctx context.Context, cmd *user.SetUsingOrgCommand) error {
	panic("not implemented")
}

func (u *UserServiceMock) GetSignedInUserWithCacheCtx(ctx context.Context, query *user.GetSignedInUserQuery) (*user.SignedInUser, error) {
	return u.GetSignedInUser(ctx, query)
}

func (u *UserServiceMock) GetSignedInUser(ctx context.Context, query *user.GetSignedInUserQuery) (*user.SignedInUser, error) {
	panic("not implemented")
}

func (u *UserServiceMock) Search(ctx context.Context, query *user.SearchUsersQuery) (*user.SearchUserQueryResult, error) {
	panic("not implemented")
}

func (u *UserServiceMock) Disable(ctx context.Context, cmd *user.DisableUserCommand) error {
	panic("not implemented")
}

func (u *UserServiceMock) BatchDisableUsers(ctx context.Context, cmd *user.BatchDisableUsersCommand) error {
	panic("not implemented")
}

func (u *UserServiceMock) UpdatePermissions(userID int64, isAdmin bool) error {
	panic("not implemented")
}

func (u *UserServiceMock) SetUserHelpFlag(ctx context.Context, cmd *user.SetUserHelpFlagCommand) error {
	panic("not implemented")
}

func (u *UserServiceMock) GetUserProfile(ctx context.Context, query *user.GetUserProfileQuery) (user.UserProfileDTO, error) {
	panic("not implemented")
}

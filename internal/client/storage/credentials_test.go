package storage

import (
	"context"
	"testing"
	"time"

	"github.com/SmoothWay/gophkeeper/pkg/models"
	"github.com/stretchr/testify/suite"
)

var (
	cred1 = models.Credentials{Type: models.CredItem, Tag: "tag1", Login: "login1", Password: "password1", Comment: "comment", Created: time.Now().Unix()}
	cred2 = models.Credentials{Type: models.CredItem, Tag: "tag1", Login: "login2", Password: "password2", Comment: "comment", Created: time.Now().Unix()}
)

type CredentialsStorager interface {
	All(ctx context.Context) ([]models.Credentials, error)
	ByLogin(ctx context.Context, login string) (models.Credentials, error)
	Save(ctx context.Context, cred models.Credentials) error
	Update(ctx context.Context, cred models.Credentials) error
}

type testCredentialsStorager interface {
	CredentialsStorager
	clean(ctx context.Context) error
}

func (s *Credentials) clean(ctx context.Context) error {
	newCtx, cancel := context.WithTimeout(ctx, time.Second*10)
	defer cancel()

	_, err := s.db.ExecContext(newCtx, "DELETE FROM credentials")
	return err
}

type CredentialsTestSuite struct {
	suite.Suite
	testCredentialsStorager
}

func (ts *CredentialsTestSuite) SetupSuite() {
	_ = Migrate("client_test.db")
	ts.testCredentialsStorager, _ = NewCredentials("client_test.db", time.Second*5)
}

func TestCredentialsSqlite(t *testing.T) {
	suite.Run(t, new(CredentialsTestSuite))
}

func (ts *CredentialsTestSuite) SetupTest() {
	ts.Require().NoError(ts.clean(context.Background()))
}

func (ts *CredentialsTestSuite) TearDownTest() {
	ts.Require().NoError(ts.clean(context.Background()))
}

func (ts *CredentialsTestSuite) TestSave() {
	err := ts.Save(context.Background(), cred1)
	ts.NoError(err)

	saved, err := ts.ByLogin(context.Background(), cred1.Login)
	ts.NoError(err)
	ts.Equal(cred1, saved)
}

func (ts *CredentialsTestSuite) TestUpdate() {
	credLogin1_1 := models.Credentials{Type: models.CredItem, Tag: "tag1", Login: "login1", Password: "password1", Comment: "comment", Created: time.Now().Unix()}
	credLogin1_2 := models.Credentials{Type: models.CredItem, Tag: "tag1", Login: "login1", Password: "NEW PASSWORD", Comment: "NEW COMMENT", Created: time.Now().Unix()}

	err := ts.Save(context.Background(), credLogin1_1)
	ts.NoError(err)

	err = ts.Update(context.Background(), credLogin1_2)
	ts.NoError(err)

	list, err := ts.All(context.Background())
	ts.NoError(err)
	ts.Equal(1, len(list))
	ts.Equal(credLogin1_2, list[0])
}

func (ts *CredentialsTestSuite) TestByLoginNoRows() {
	cred, err := ts.ByLogin(context.Background(), "login1")
	ts.ErrorIs(err, ErrItemNotFound)
	ts.Equal(models.Credentials{}, cred)
}

func (ts *CredentialsTestSuite) TestAll() {
	err := ts.Save(context.Background(), cred1)
	ts.NoError(err)
	err = ts.Save(context.Background(), cred2)
	ts.NoError(err)

	list, err := ts.All(context.Background())
	ts.NoError(err)
	ts.Equal(2, len(list))

	set := ts.setFromList(list)
	ts.True(set[cred1])
	ts.True(set[cred2])
}

func (ts *CredentialsTestSuite) setFromList(list []models.Credentials) map[models.Credentials]bool {
	res := make(map[models.Credentials]bool, len(list))
	for _, cred := range list {
		res[cred] = true
	}
	return res
}

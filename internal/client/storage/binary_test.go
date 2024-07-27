package storage

import (
	"context"
	"testing"
	"time"

	"github.com/SmoothWay/gophkeeper/pkg/models"
	"github.com/stretchr/testify/suite"
)

var (
	binary1 = models.Binary{Type: models.BinItem, Tag: "tag1", Key: "file1.txt", Value: []byte("file1 content"), Comment: "file with secrets", Created: time.Now().Unix()}
	binary2 = models.Binary{Type: models.BinItem, Tag: "tag1", Key: "file2.txt", Value: []byte("file2 content"), Comment: "file with secrets", Created: time.Now().Unix()}
)

type BinaryStorager interface {
	All(ctx context.Context) ([]models.Binary, error)
	ByKey(ctx context.Context, key string) (models.Binary, error)
	Save(ctx context.Context, bin models.Binary) error
	Update(ctx context.Context, bin models.Binary) error
}

type testBinaryStorager interface {
	BinaryStorager
	clean(ctx context.Context) error
}

func (s *BinaryStorage) clean(ctx context.Context) error {
	newCtx, cancel := context.WithTimeout(ctx, time.Second*10)
	defer cancel()

	_, err := s.db.ExecContext(newCtx, "DELETE FROM binary")
	return err
}

type BinaryTestSuite struct {
	suite.Suite
	testBinaryStorager
}

func (ts *BinaryTestSuite) SetupSuite() {
	_ = Migrate("client_test.db")
	ts.testBinaryStorager, _ = NewBinary("client_test.db", time.Second*5)
}

func TestBinarySqlite(t *testing.T) {
	suite.Run(t, new(BinaryTestSuite))
}

func (ts *BinaryTestSuite) SetupTest() {
	ts.Require().NoError(ts.clean(context.Background()))
}

func (ts *BinaryTestSuite) TearDownTest() {
	ts.Require().NoError(ts.clean(context.Background()))
}

func (ts *BinaryTestSuite) TestSave() {
	err := ts.Save(context.Background(), binary1)
	ts.NoError(err)

	saved, err := ts.ByKey(context.Background(), "file1.txt")
	ts.NoError(err)
	ts.Equal(binary1, saved)
}

func (ts *BinaryTestSuite) TestUpdate() {
	binKey1_1 := models.Binary{Type: models.BinItem, Tag: "tag1", Key: "file1.txt", Value: []byte("file1 content"), Comment: "comment", Created: time.Now().Unix()}
	binKey1_2 := models.Binary{Type: models.BinItem, Tag: "tag1", Key: "file1.txt", Value: []byte("NEW CONTENT"), Comment: "NEW COMMENT", Created: time.Now().Unix()}

	err := ts.Save(context.Background(), binKey1_1)
	ts.NoError(err)

	err = ts.Update(context.Background(), binKey1_2)
	ts.NoError(err)

	list, err := ts.All(context.Background())
	ts.NoError(err)
	ts.Equal(1, len(list))
	ts.Equal(binKey1_2, list[0])
}

func (ts *BinaryTestSuite) TestByKeyNoRows() {
	bin, err := ts.ByKey(context.Background(), "file10.txt")
	ts.ErrorIs(err, ErrItemNotFound)
	ts.Equal(models.Binary{}, bin)
}

func (ts *BinaryTestSuite) TestAll() {
	err := ts.Save(context.Background(), binary1)
	ts.NoError(err)
	err = ts.Save(context.Background(), binary2)
	ts.NoError(err)

	list, err := ts.All(context.Background())
	ts.NoError(err)
	ts.Equal(2, len(list))
	ts.True(contains(binary1, list))
	ts.True(contains(binary2, list))
}

func contains(target models.Binary, list []models.Binary) bool {
	for _, b := range list {
		if b.Type == target.Type && b.Key == target.Key &&
			string(b.Value) == string(target.Value) && b.Comment == target.Comment &&
			b.Tag == target.Tag && b.Created == target.Created {
			return true
		}
	}
	return false
}

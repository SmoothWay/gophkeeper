package storage

import (
	"context"
	"testing"
	"time"

	"github.com/SmoothWay/gophkeeper/pkg/models"
	"github.com/stretchr/testify/suite"
)

var (
	text1 = models.Text{Type: models.TextItem, Tag: "tag1", Key: "key1", Value: "value 1", Comment: "comment 1", Created: 1}
	text2 = models.Text{Type: models.TextItem, Tag: "tag2", Key: "KEY2", Value: "VALUE2", Comment: "commnt", Created: 2}
)

type ITextStorager interface {
	All(ctx context.Context) ([]models.Text, error)
	ByKey(ctx context.Context, key string) (models.Text, error)
	Save(ctx context.Context, text models.Text) error
	Update(ctx context.Context, text models.Text) error
}

type testTextStorager interface {
	ITextStorager
	clean(ctx context.Context) error
}

type TextTestSuite struct {
	suite.Suite
	testTextStorager
}

func (ts *TextTestSuite) SetupSuite() {
	_ = Migrate("client_test.db")
	ts.testTextStorager, _ = NewText("client_test.db", time.Second*3)
}

func (ts *TextStorage) clean(ctx context.Context) error {
	newCtx, cancel := context.WithTimeout(ctx, time.Second*10)
	defer cancel()

	_, err := ts.db.ExecContext(newCtx, "DELETE from text")
	return err
}

func TestText(t *testing.T) {
	suite.Run(t, new(TextTestSuite))
}

func (ts *TextTestSuite) SetupTest() {
	ts.Require().NoError(ts.clean(context.Background()))
}

func (ts *TextTestSuite) TearDownTest() {
	ts.Require().NoError(ts.clean(context.Background()))
}

func (ts *TextTestSuite) TestSave() {
	err := ts.Save(context.Background(), text1)
	ts.NoError(err)

	saved, err := ts.ByKey(context.Background(), text1.Key)
	ts.NoError(err)

	ts.Equal(text1, saved)
}

func (ts *TextTestSuite) TestUpdate() {
	textKey1_1 := models.Text{Type: models.TextItem, Tag: "tag1", Key: "key1", Value: "value 1", Comment: "comment", Created: time.Now().Unix()}
	textKey1_2 := models.Text{Type: models.TextItem, Tag: "tag1", Key: "key1", Value: "NEW VALUE", Comment: "NEW comment", Created: time.Now().Unix()}

	err := ts.Save(context.Background(), textKey1_1)
	ts.NoError(err)

	err = ts.Update(context.Background(), textKey1_2)
	ts.NoError(err)

	list, err := ts.All(context.Background())
	ts.NoError(err)
	ts.Equal(1, len(list))
	ts.Equal(textKey1_2, list[0])
}

func (ts *TextTestSuite) TestByKeyNoRows() {
	text, err := ts.ByKey(context.Background(), "key1")
	ts.ErrorIs(err, ErrItemNotFound)
	ts.Equal(models.Text{}, text)
}

func (ts *TextTestSuite) TestAll() {
	err := ts.Save(context.Background(), text1)
	ts.NoError(err)
	err = ts.Save(context.Background(), text2)
	ts.NoError(err)

	list, err := ts.All(context.Background())
	ts.NoError(err)
	ts.Equal(2, len(list))

	set := ts.setFromList(list)

	ts.True(set[text1])
	ts.True(set[text2])
}

func (ts *TextTestSuite) setFromList(list []models.Text) map[models.Text]bool {
	res := make(map[models.Text]bool, len(list))
	for _, text := range list {
		res[text] = true
	}
	return res
}

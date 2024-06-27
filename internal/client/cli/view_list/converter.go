package viewlist

import (
	"fmt"

	"github.com/SmoothWay/gophkeeper/pkg/models"
)

func Convert(creds []models.Credentials, texts []models.Text, bins []models.Binary, cards []models.Card) []string {
	viewList := make([]string, 0, len(creds)+len(texts)+len(bins)+len(cards))

	if len(creds) > 0 {
		viewList = append(viewList, "Credentials:")
		for _, c := range creds {
			viewList = append(viewList, fmt.Sprintf("tag=%s; login=%s; password=%s; comment=%s.", c.Tag, c.Login, c.Password, c.Comment))
		}
	}

	if len(texts) > 0 {
		viewList = append(viewList, "Text data:")
		for _, t := range texts {
			viewList = append(viewList, fmt.Sprintf(`tag=%s; key=%s; value=%s; comment=%s.`, t.Tag, t.Key, t.Value, t.Comment))
		}
	}
	if len(bins) > 0 {
		viewList = append(viewList, "Binary data:")
		for _, b := range bins {
			viewList = append(viewList, fmt.Sprintf(`tag=%s; key=%s; comment=%s.`, b.Tag, b.Key, b.Comment))
		}
	}
	if len(cards) > 0 {
		viewList = append(viewList, "Card data:")
		for _, c := range cards {
			viewList = append(viewList, fmt.Sprintf(`tag=%s; number=%s; exp=%s; cvv=%d; comment=%s`, c.Tag, c.Number, c.Exp, c.Cvv, c.Comment))
		}
	}
	if len(viewList) == 0 {
		viewList = append(viewList, "secrets list is empty")
	}
	return viewList
}

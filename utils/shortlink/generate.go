package shortlink

import (
	"fmt"
	"log"

	"github.com/google/uuid"
	"github.com/slaveofcode/hansip/repository/pg/models"
	"github.com/spf13/viper"
	"github.com/teris-io/shortid"
	"gorm.io/gorm"
)

const (
	SHORTID_CHARS = "0123456789abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ!$"
)

var shortId *shortid.Shortid

func init() {
	workerNum := viper.GetInt("short_id.worker")
	seed := viper.GetUint64("short_id.seed")
	sId, err := shortid.New(uint8(workerNum), SHORTID_CHARS, seed)
	if err != nil {
		log.Println("Failed to initialize shortid")
	}

	shortId = sId
}

func newRandCode() string {
	code, err := shortId.Generate()
	if err != nil {
		return newRandCode()
	}

	return code
}

func MakeNewCode(fileGroupId *uuid.UUID, pin string, db *gorm.DB) (*models.ShortLink, error) {
	code := newRandCode()
	var shortLink models.ShortLink
	res := db.Where(`"shortCode" = ?`, code).First(&shortLink)
	if res.RowsAffected > 0 {
		return MakeNewCode(fileGroupId, pin, db)
	}

	newShortLink := models.ShortLink{
		FileGroupId: fileGroupId,
		ShortCode:   code,
		PIN:         pin,
	}

	res = db.Create(&newShortLink)
	if res.Error != nil {
		return nil, res.Error
	}

	return &newShortLink, nil
}

func MakeURL(shortLink *models.ShortLink) string {
	siteUrl := viper.GetString("site.url")
	shortlinkPath := viper.GetString("site.shortlink_path")
	return fmt.Sprintf("%s%s/%s", siteUrl, shortlinkPath, shortLink.ShortCode)
}

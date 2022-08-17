package shortlink

import (
	"fmt"
	"log"
	"os"
	"strconv"

	"github.com/google/uuid"
	"github.com/slaveofcode/securi/repository/pg/models"
	"github.com/teris-io/shortid"
	"gorm.io/gorm"
)

const (
	SHORTID_CHARS = "0123456789abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ!$"
)

var shortId *shortid.Shortid

func init() {
	workerNum, _ := strconv.Atoi(os.Getenv("SHORTID_WORKER"))
	seed, _ := strconv.Atoi(os.Getenv("SHORTID_SEED"))
	sId, err := shortid.New(uint8(workerNum), SHORTID_CHARS, uint64(seed))
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
	siteUrl := os.Getenv("SITE_URL_BASE")
	shortlinkPath := os.Getenv("SITE_URL_SHORTLINK_PATH")
	return fmt.Sprintf("%s%s/%s", siteUrl, shortlinkPath, shortLink.ShortCode)
}

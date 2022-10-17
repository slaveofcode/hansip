package shortlink

import (
	"fmt"
	"log"

	"github.com/slaveofcode/hansip/repository/models"
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

func MakeNewCode(fileGroupId uint64, pin string, db *gorm.DB) (*models.ShortLink, error) {
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
	shortlinkPath := viper.GetString("site.shortlink_path")

	siteProtocol := "http://"
	if viper.GetBool("server_web.secure") {
		siteProtocol = "https://"
	}

	siteAddr := siteProtocol + viper.GetString("server_web.host")
	sitePort := viper.GetString("server_web.port")

	if sitePort != "80" {
		siteAddr = fmt.Sprintf("%s:%s", siteAddr, sitePort)
	}

	return fmt.Sprintf("%s%s/%s", siteAddr, shortlinkPath, shortLink.ShortCode)
}

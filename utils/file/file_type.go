package file

import (
	"os"

	"github.com/h2non/filetype"
	"github.com/slaveofcode/securi/repository/pg/models"
)

func GetHeadFilePreviewValue(file *os.File) models.PreviewAsType {
	headBytes := make([]byte, 261)
	file.Read(headBytes)

	if filetype.IsImage(headBytes) {
		return models.PreviewAsImage
	}

	if filetype.IsVideo(headBytes) {
		return models.PreviewAsVideo
	}

	if filetype.IsAudio(headBytes) {
		return models.PreviewAsAudio
	}

	if filetype.IsDocument(headBytes) {
		return models.PreviewAsDocument
	}

	if filetype.IsArchive(headBytes) {
		return models.PreviewAsArchive
	}

	if filetype.IsFont(headBytes) {
		return models.PreviewAsFont
	}

	return models.PreviewAsBinary
}

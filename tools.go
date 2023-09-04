package toolkit

import (
	"crypto/rand"
	"errors"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"strings"
)

const randomStringSource = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789_+"

type Tools struct {
	MaxFileSize      int64
	AllowedFileTypes []string
}

func (t *Tools) RandomString(n int) string {
	s, r := make([]rune, n), []rune(randomStringSource)

	rLen := len(r)
	uIntRlen := uint64(len(r))

	for i := range s {
		p, err := rand.Prime(rand.Reader, rLen)

		if err != nil {
			log.Printf("An error occurred: %v", err)
		}

		x := p.Uint64()
		s[i] = r[x%uIntRlen]
	}

	return string(s)
}

type UploadedFile struct {
	NewFileName      string
	OriginalFileName string
	FileSize         int64
}

func (t *Tools) UploadFiles(r *http.Request, uploadDir string, rename ...bool) ([]*UploadedFile, error) {

	renameFile := true

	if len(rename) > 0 {
		renameFile = rename[0]
	}

	var uploadedFiles []*UploadedFile

	if t.MaxFileSize == 0 {
		t.MaxFileSize = 1024 * 1024 * 1024
	}
	err := r.ParseMultipartForm(t.MaxFileSize)

	if err != nil {
		return nil, errors.New("Uploaded file is too big")
	}

	allowedTypes := []string{"image/jpeg", "image/gif", "image/png"}
	if len(t.AllowedFileTypes) == 0 {
		t.AllowedFileTypes = allowedTypes
	}
	for _, fHeaders := range r.MultipartForm.File {
		for _, hdr := range fHeaders {
			uploadedFiles, err = func(uploadedFiles []*UploadedFile) ([]*UploadedFile, error) {
				var uploadedFile UploadedFile
				inFile, err := hdr.Open()
				if err != nil {
					return nil, err
				}
				defer inFile.Close()
				buff := make([]byte, 512)
				_, err = inFile.Read(buff)
				if err != nil {
					return nil, err
				}

				//TODO =- check to see if filetype is permitted
				allowed := false
				fileType := http.DetectContentType(buff)
				if len(t.AllowedFileTypes) > 0 {
					for _, x := range t.AllowedFileTypes {
						if strings.EqualFold(fileType, x) {
							allowed = true
						}
					}
				}

				if !allowed {
					return nil, errors.New("the uploaded filetype is not permitted")
				}

				_, err = inFile.Seek(0, 0)
				if err != nil {
					return nil, err
				}

				if renameFile {
					uploadedFile.NewFileName = fmt.Sprintf("%s%s", t.RandomString(25), filepath.Ext(hdr.Filename))
				} else {
					uploadedFile.NewFileName = hdr.Filename
				}

				var outFile *os.File
				defer outFile.Close()

				if outFile, err = os.Create(filepath.Join(uploadDir, uploadedFile.NewFileName)); err != nil {
					return nil, err
				} else {
					fileSize, err := io.Copy(outFile, inFile)

					if err != nil {
						return nil, err
					}

					uploadedFile.FileSize = fileSize
				}

				uploadedFiles = append(uploadedFiles, &uploadedFile)

				return uploadedFiles, nil

			}(uploadedFiles)

			if err != nil {
				return uploadedFiles, err
			}
		}
	}

	return uploadedFiles, nil
}

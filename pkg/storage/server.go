package storage

import (
	"bytes"
	"encoding/hex"
	"errors"
	"fmt"
	"github.com/shaowenchen/ops/pkg/option"
	"io"
	"mime/multipart"
	"net/http"
	"os"
	"path/filepath"
	"regexp"
	"strings"
)

func ServerFile(fileOpt option.FileOption) (stdout string, err error) {
	if fileOpt.Api == "" {
		err = errors.New("please provide a valid api")
		return
	}
	if fileOpt.IsUploadDirection() {
		if fileOpt.AesKey != UnSetFlag {
			if fileOpt.AesKey == "" {
				aesKey, err1 := GetDefaultRandomKey()
				if err1 != nil {
					err = err1
					return
				}
				fileOpt.AesKey = string(aesKey)
			}
			tartgetFile := fileOpt.LocalFile + ".aes"
			err = EncryptFile(fileOpt.AesKey, fileOpt.LocalFile, tartgetFile)
			if err != nil {
				return
			}
			fileOpt.LocalFile = tartgetFile
		}

		err1, resp := postFileToRemote(fileOpt.LocalFile, fileOpt.Api)
		if err1 != nil {
			err = err1
			return
		}
		stdout = "Please use the following command to download the file:" +
			buildDowloadOpscliCmd(fileOpt.Api, resp, hex.EncodeToString([]byte(fileOpt.AesKey)))
		return
	} else if fileOpt.IsDownloadDirection() {
		if fileOpt.LocalFile == "" {
			fileOpt.LocalFile = filepath.Base(fileOpt.RemoteFile)
		}
		err = getFileToLocal(fileOpt.RemoteFile, fileOpt.LocalFile)
		if err != nil {
			return
		}
		if fileOpt.AesKey != UnSetFlag && strings.HasSuffix(fileOpt.LocalFile, ".aes") {
			tartgetFile := strings.TrimSuffix(fileOpt.LocalFile, ".aes")
			err = DecryptFile(fileOpt.AesKey, fileOpt.LocalFile, tartgetFile)
			if err != nil {
				return
			}
		}
	} else {
		err = errors.New("Please provide a valid direction")
	}
	return
}

func buildDowloadOpscliCmd(api, resp, aesKey string) string {
	var urlRegex = regexp.MustCompile(`https?:\/\/(www\.)?[-a-zA-Z0-9@:%._\+~#=]{1,256}\.[a-zA-Z0-9()]{1,6}\b([-a-zA-Z0-9()@:%_\+.~#?&//=]*)`)
	subs := urlRegex.FindStringSubmatch(resp)
	downloadUrl := ""
	if len(subs) > 0 {
		downloadUrl = subs[0]
	} else {
		return fmt.Sprintf("No url found in response: %s", resp)
	}
	return fmt.Sprintf("opscli file --api %s --aeskey %s --direction download --remotefile %s", api, aesKey, downloadUrl)
}

func getFileToLocal(downloadUrl, localFilePath string) (err error) {
	file, err := os.Create(localFilePath)
	if err != nil {
		return err
	}
	defer file.Close()

	resp, err := http.Get(downloadUrl)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	_, err = io.Copy(file, resp.Body)
	if err != nil {
		return err
	}
	return nil
}

func postFileToRemote(localFilePath, api string) (error, string) {

	file, err := os.Open(localFilePath)
	if err != nil {
		return err, ""
	}
	defer file.Close()

	bodyBuf := &bytes.Buffer{}

	bodyWriter := multipart.NewWriter(bodyBuf)

	fileWriter, err := bodyWriter.CreateFormFile("file", filepath.Base(localFilePath))
	if err != nil {
		return err, ""
	}

	_, err = io.Copy(fileWriter, file)
	if err != nil {
		return err, ""
	}

	bodyWriter.Close()

	req, err := http.NewRequest("POST", api, bodyBuf)
	if err != nil {
		return err, ""
	}

	req.Header.Add("Content-Type", bodyWriter.FormDataContentType())

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return err, ""
	}
	defer resp.Body.Close()

	body, _ := io.ReadAll(resp.Body)
	return nil, string(body)
}

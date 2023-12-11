package storage

import (
	"bytes"
	"encoding/hex"
	"fmt"
	"io"
	"mime/multipart"
	"net/http"
	"os"
	"path/filepath"
	"strings"

	"github.com/shaowenchen/ops/pkg/log"
	"github.com/shaowenchen/ops/pkg/option"
	"github.com/shaowenchen/ops/pkg/utils"
)

func ServerFile(logger *log.Logger, fileOpt option.FileOption, serverOpt option.FileServerOption) (err error) {
	if serverOpt.Api == "" {
		return fmt.Errorf("please provide a valid api")
	}
	if utils.IsUploadDirection(fileOpt.Direction) {
		if fileOpt.AesKey != UnSetFlag {
			if fileOpt.AesKey == "" {
				aesKey, err := GetDefaultRandomKey()
				logger.Info.Println("aesKey: ", hex.EncodeToString(aesKey))
				if err != nil {
					return err
				}
				fileOpt.AesKey = string(aesKey)
			}
			tartgetFile := fileOpt.LocalFile + ".aes"
			err = EncryptFile(fileOpt.AesKey, fileOpt.LocalFile, tartgetFile)
			if err != nil {
				return err
			}
			fileOpt.LocalFile = tartgetFile
		}

		err, resp := postFileToRemote(fileOpt.LocalFile, serverOpt.Api)
		if err != nil {
			logger.Error.Println(err)
		}
		logger.Info.Println(resp)
	} else if utils.IsDownloadDirection(fileOpt.Direction) {
		if fileOpt.LocalFile == "" {
			fileOpt.LocalFile = filepath.Base(fileOpt.RemoteFile)
		}
		err = getFileToLocal(fileOpt.RemoteFile, fileOpt.LocalFile)
		if err != nil {
			logger.Error.Println(err)
		}
		if fileOpt.AesKey != UnSetFlag && strings.HasSuffix(fileOpt.LocalFile, ".aes") {
			tartgetFile := strings.TrimSuffix(fileOpt.LocalFile, ".aes")
			err = DecryptFile(fileOpt.AesKey, fileOpt.LocalFile, tartgetFile)
			if err != nil {
				return err
			}
		}
	} else {
		logger.Error.Println("Please provide a valid direction")
	}
	if err != nil {
		logger.Error.Println(err)
	}
	return
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

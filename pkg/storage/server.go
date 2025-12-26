package storage

import (
	"bytes"
	"encoding/hex"
	"errors"
	"fmt"
	"io"
	"mime/multipart"
	"net/http"
	"os"
	"path/filepath"
	"regexp"
	"strings"

	"github.com/shaowenchen/ops/pkg/option"
	"github.com/shaowenchen/ops/pkg/utils"
)

func ServerFile(fileOpt option.FileOption) (stdout string, err error) {
	if fileOpt.Api == "" {
		err = errors.New("please provide a valid api")
		return
	}
	if fileOpt.IsUploadDirection() {
		// Default: encrypt with auto-generated key. Only skip if --aeskey "" is set
		if fileOpt.AesKey == UnSetFlag {
			aesKey, err1 := GetDefaultRandomKey()
			if err1 != nil {
				err = err1
				return
			}
			fileOpt.AesKey = string(aesKey)
		} else if fileOpt.AesKey != "" {
			aeskeyBytes, err1 := hex.DecodeString(fileOpt.AesKey)
			if err1 != nil {
				stdout = err1.Error()
				return stdout, err
			}
			fileOpt.AesKey = string(aeskeyBytes)
		}

		// Encrypt if key is set
		if fileOpt.AesKey != "" {
			tartgetFile := fileOpt.LocalFile + ".aes"
			err = EncryptFile(fileOpt.AesKey, fileOpt.LocalFile, tartgetFile)
			defer os.Remove(tartgetFile)
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
		if fileOpt.AesKey != "" {
			stdout = "Please use the following command to download the file: \n" +
				buildDowloadOpscliCmd(fileOpt.Api, resp, hex.EncodeToString([]byte(fileOpt.AesKey)))
		}
		return
	} else if fileOpt.IsDownloadDirection() {
		if fileOpt.LocalFile == "" {
			fileOpt.LocalFile = filepath.Base(fileOpt.RemoteFile)
		}
		err = getFileToLocal(fileOpt.RemoteFile, fileOpt.LocalFile)
		if err != nil {
			return
		}
		targetFile := fileOpt.LocalFile
		if fileOpt.AesKey != UnSetFlag && strings.HasSuffix(fileOpt.LocalFile, ".aes") {
			targetFile = strings.TrimSuffix(fileOpt.LocalFile, ".aes")
			err = DecryptFile(fileOpt.AesKey, fileOpt.LocalFile, targetFile)
			defer os.Remove(fileOpt.LocalFile)
			if err != nil {
				return
			}
		}
		stdout = fmt.Sprintf("success download %s to %s", fileOpt.RemoteFile, targetFile)
	} else {
		stdout = fmt.Sprintf("Unknown direction: %s", fileOpt.Direction)
		err = errors.New(stdout)

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
	return fmt.Sprintf("opscli file --fileapi %s --aeskey %s --direction download --remotefile %s", api, aesKey, downloadUrl)
}

func getFileToLocal(downloadUrl, localFilePath string) (err error) {
	file, err := utils.CreateFile(localFilePath)
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

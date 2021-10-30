package license

import (
	"fmt"
	"io/ioutil"
	"os"
	"time"

	"github.com/admpub/nging/v3/application/library/restclient"
	"github.com/webx-top/com"
	"github.com/webx-top/echo"
)

func DownloadOnce(ctx echo.Context) error {
	downloadOnce.Do(func() {
		downloadTime = time.Now()
		downloadError = Download(ctx)
	})
	return downloadError
}

// Download 从官方服务器重新下载许可证
func Download(ctx echo.Context) error {
	operation := `获取授权证书失败：%v`
	client := restclient.Resty()
	client.SetHeader("Accept", "application/json")
	officialResponse := &OfficialResponse{}
	client.SetResult(officialResponse)
	fullURL := FullLicenseURL(ctx) + `&pipe=download`
	response, err := client.Get(fullURL)
	if err != nil {
		return fmt.Errorf(operation, err)
	}
	if response == nil {
		return ErrConnectionFailed
	}
	if response.IsError() {
		return fmt.Errorf(operation, string(response.Body()))
	}
	if officialResponse.Code != 1 {
		return fmt.Errorf(`%v`, officialResponse.Info)
	}
	if officialResponse.Data == nil {
		return ErrLicenseDownloadFailed
	}
	if com.FileExists(licenseFile) {
		err = os.Rename(licenseFile, licenseFile+`.`+time.Now().Format(`20060102150405`))
		if err != nil {
			return err
		}
	}
	licenseData = &officialResponse.Data.LicenseData
	b, err := com.JSONEncode(licenseData, `  `)
	if err != nil {
		b = []byte(err.Error())
	}
	err = ioutil.WriteFile(licenseFile, b, os.ModePerm)
	if err != nil {
		return fmt.Errorf(`保存授权证书失败：%v`, err)
	}
	return err
}
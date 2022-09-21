package main

import (
	"encoding/json"
	"errors"
	"io/ioutil"
	"log"
	"net/http"

	alidns "github.com/alibabacloud-go/alidns-20150109/v2/client"
	openapi "github.com/alibabacloud-go/darabonba-openapi/client"
	"github.com/alibabacloud-go/tea/tea"
)

type Configs struct {
	UrlForGetIp string `json:"UrlForGetIp"`
	ID          string `json:"ID"`
	SECRET      string `json:"SECRET"`
	DOMAIN      string `json:"DOMAIN"`
	SubDomain   string `json:"SubDomain"`
}

var Config Configs

func init() {
	data, err := ioutil.ReadFile("./config.json")
	if err != nil {
		panic(err)
	}
	json.Unmarshal(data, &Config)
}

func main() {

	client, err := createClient()
	if err != nil {
		panic(err)
	}

	ip, err := getIP()
	if err != nil {
		panic(err)
	}

	id, err := query(client)
	if err != nil {
		panic(err)
	}

	_, err = update(client, ip, id)
	if err != nil {
		log.Println("IP无变动：" + ip)
	} else {
		log.Println("域名解析记录IP更改为：" + ip)
	}
}

func getIP() (string, error) {
	response, err := http.Get(Config.UrlForGetIp)
	if err != nil {
		return "", err
	}
	ip, err := ioutil.ReadAll(response.Body)
	if err != nil {
		return "", err
	}
	return string(ip), nil
}

func createClient() (*alidns.Client, error) {
	config := &openapi.Config{
		AccessKeyId:     tea.String(Config.ID),
		AccessKeySecret: tea.String(Config.SECRET),
	}
	client, err := alidns.NewClient(config)
	if err != nil {
		return nil, err
	}

	return client, nil
}

func update(client *alidns.Client, ip string, id string) (*alidns.UpdateDomainRecordResponse, error) {
	request := &alidns.UpdateDomainRecordRequest{
		RecordId: tea.String(id),
		RR:       tea.String(Config.SubDomain),
		Type:     tea.String("A"),
		TTL:      tea.Int64(600),
		Value:    tea.String(ip),
	}
	response, err := client.UpdateDomainRecord(request)
	if err != nil {
		return response, err
	}
	return response, nil
}

func query(client *alidns.Client) (string, error) {
	request := &alidns.DescribeDomainRecordsRequest{DomainName: tea.String(Config.DOMAIN)}
	response, err := client.DescribeDomainRecords(request)
	if err != nil {
		return "", err
	}
	data := response.Body.DomainRecords.Record
	for _, iterm := range data {
		if *iterm.RR == Config.SubDomain {
			return *iterm.RecordId, nil
		}
	}
	return "response", errors.New("未找到子域名，请先建立对应解析")
}

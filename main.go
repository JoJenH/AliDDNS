package main

import (
	"errors"
	alidns "github.com/alibabacloud-go/alidns-20150109/v2/client"
	openapi "github.com/alibabacloud-go/darabonba-openapi/client"
	"github.com/alibabacloud-go/tea/tea"
	"io/ioutil"
	"log"
	"net/http"
	"os"
)

var (
	UrlForGetIp = "http://www.3322.org/dyndns/getip"
	ID          = ""
	SECRET      = ""
	DOMAIN      = ""
	SubDomain   = ""
)

func main() {
	file, err := os.OpenFile("ddns.log", os.O_CREATE|os.O_APPEND, 0644)
	if err != nil {
		panic(err)
	}
	log.SetOutput(file)

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
	response, err := http.Get(UrlForGetIp)
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
		AccessKeyId:     tea.String(ID),
		AccessKeySecret: tea.String(SECRET),
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
		RR:       tea.String(SubDomain),
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
	request := &alidns.DescribeDomainRecordsRequest{DomainName: tea.String(DOMAIN)}
	response, err := client.DescribeDomainRecords(request)
	if err != nil {
		return "", err
	}
	data := response.Body.DomainRecords.Record
	for _, iterm := range data {
		if *iterm.RR == SubDomain {
			return *iterm.RecordId, nil
		}
	}
	return "response", errors.New("未找到子域名，请先建立对应解析")
}

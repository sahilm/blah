package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"strings"
	"time"

	"strconv"

	"crypto/tls"

	"github.com/cloudfoundry-community/go-cfenv"
)

type Backup struct {
	Status bool `json:"status"`
}

func main() {

	appEnv, err := cfenv.Current()
	if err != nil {
		log.Fatal(err)
	}

	log.Println(appEnv)

	http.HandleFunc("/backups", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			w.WriteHeader(http.StatusMethodNotAllowed)
			return
		}
		body, err := ioutil.ReadAll(r.Body)
		if err != nil {
			log.Fatal(err)
		}
		var backup Backup
		json.Unmarshal(body, &backup)
		log.Println(string(body))

		tr := &http.Transport{
			TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
		}

		client := http.Client{
			Timeout:   5 * time.Second,
			Transport: tr,
		}

		requestBody := strings.NewReader(fmt.Sprintf(METRIC_BODY_TEMPLATE, appEnv.AppID, time.Now().Unix()*1000))

		metricsForwarder, err := appEnv.Services.WithName("metrics1")
		if err != nil {
			log.Fatal(err)
		}

		req, err := http.NewRequest("POST", metricsForwarder.Credentials["endpoint"].(string), requestBody)
		req.Header.Add("Authorization", metricsForwarder.Credentials["access_key"].(string))
		req.Header.Add("Content-Type", "application/json")

		log.Println(req)

		resp, err := client.Do(req)

		if err != nil {
			log.Fatal(err)
		}

		rr, err := ioutil.ReadAll(resp.Body)
		if err != nil {
			log.Fatal(err)
		}
		fmt.Println(resp.Status, string(rr))
	})

	log.Fatal(http.ListenAndServe(":"+strconv.Itoa(appEnv.Port), nil))

}

const METRIC_BODY_TEMPLATE string = `
{
	"applications": [
	{
		"id": "%s",
		"instances": [
		{
			"id": "123456",
			"index": "0",
			"metrics": [
			{
				"name": "backup-abcdedf",
				"type": "gauge",
				"value": 0,
				"timestamp": %d,
				"unit": "bool"
			}]
		}]
	}]
}
`

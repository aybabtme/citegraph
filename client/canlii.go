package client

import (
	"bufio"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"net/url"
	"os"
	"strconv"
)

const (
	apiURL          = "http://api.canlii.org/v1/"
	apiKeysFilename = "apiKey"
)

var (
	apiKeys     []string
	keyRotation int
)

func init() {
	log.Printf("Loading API keys from file `%s`", apiKeysFilename)
	f, err := os.Open(apiKeysFilename)
	if err != nil {
		panic(err)
	}
	defer f.Close()
	scanner := bufio.NewScanner(f)
	for scanner.Scan() {
		apiKeys = append(apiKeys, scanner.Text())
	}
	log.Printf("Done, found %d keys:\n%v", len(apiKeys), apiKeys)
}

func getAPIKey() url.Values {
	if keyRotation >= len(apiKeys) {
		keyRotation = 0
	}
	defer func() { keyRotation++ }()

	val := url.Values{}
	val.Set("api_key", apiKeys[keyRotation])
	return val
}

func doQuery(qURL *url.URL) ([]byte, error) {
	resp, err := http.Get(qURL.String())
	if err != nil {
		return nil, fmt.Errorf("http.Get(url), %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		return nil, fmt.Errorf("query `%v` returned status code %d, %s", qURL.String(), resp.StatusCode, resp.Status)
	}

	dump, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("reading response, %v", err)
	}
	return dump, nil
}

type Database struct {
	ID           string `json: "databaseId"`
	Jurisdiction string `json: "jurisdiction"`
	Name         string `json: "name"`
}

func DatabaseList() ([]Database, error) {
	collection := "caseBrowse/en/?"

	qURL, err := url.Parse(apiURL + collection + getAPIKey().Encode())
	if err != nil {
		return nil, fmt.Errorf("parsing collection url, %v", err)
	}

	dump, err := doQuery(qURL)
	if err != nil {
		return nil, fmt.Errorf("failed querying database list, %v", err)
	}

	dbL := struct {
		DbList []Database `json:"caseDatabases"`
	}{}

	if err := json.Unmarshal(dump, &dbL); err != nil {
		return nil, fmt.Errorf("unmarshalling response, %v, query was `%s`, got `%v`", err, qURL.String(), string(dump))
	}

	return dbL.DbList, nil
}

func (d *Database) CaseList(offset, count int) ([]Case, error) {
	collection := "caseBrowse/en/" + d.ID + "/?"
	val := getAPIKey()
	val.Add("offset", strconv.Itoa(offset))
	val.Add("resultCount", strconv.Itoa(count))

	qURL, err := url.Parse(apiURL + collection + val.Encode())
	if err != nil {
		return nil, fmt.Errorf("parsing query url, %v", err)
	}

	dump, err := doQuery(qURL)
	if err != nil {
		return nil, fmt.Errorf("failed querying case list, %v", err)
	}

	caseL := struct {
		Cases []Case `json:"cases"`
	}{nil}

	if err := json.Unmarshal(dump, &caseL); err != nil {
		return nil, fmt.Errorf("unmarshalling cases, %v, data = %v", err, string(dump))
	}

	return caseL.Cases, nil
}

type Case struct {
	DbID     string `json:"databaseId"`
	ID       string `json:"caseId["en"]"`
	Title    string `json:"title"`
	Citation string `json:"citation"`
}

package main

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"math/rand"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"

	"github.com/romainmenke/web-tests/scripts/web-tests-browserstack/api"
	"github.com/tebeka/selenium"
	"golang.org/x/sync/semaphore"
)

func main() {
	ctx, cancel := context.WithTimeout(context.Background(), time.Minute*30)
	defer cancel()

	sessionName := fmt.Sprintf("Web Tests – %s", time.Now().Format(time.RFC3339))
	userName := os.Getenv("BROWSERSTACK_USERNAME")
	accessKey := os.Getenv("BROWSERSTACK_ACCESS_KEY")

	mapping, err := getMapping()
	if err != nil {
		log.Fatal(err)
	}

	client := api.New(api.Config{
		UserName:  userName,
		AccessKey: accessKey,
	})

	browsers, err := client.ReducedBrowsers(ctx)
	if err != nil {
		log.Fatal(err)
	}

	done, err := client.OpenTunnel(ctx)
	defer done()
	if err != nil {
		log.Fatal(err)
	}

	log.Println("tunnel ready")

	rand.Seed(time.Now().UnixNano())
	rand.Shuffle(len(browsers), func(i, j int) {
		browsers[i], browsers[j] = browsers[j], browsers[i]
	})

	browsers = browsers[:50]

	sema := semaphore.NewWeighted(5)

	for _, b := range browsers {
		if err := sema.Acquire(ctx, 1); err != nil {
			log.Fatal(err)
		}

		go func(b api.Browser) {
			defer sema.Release(1)
			err = runTest(ctx, client, b, sessionName, mapping)
			if err != nil {
				log.Println(err) // non-fatal for us
			}
		}(b)
	}

	// Wait for all
	if err := sema.Acquire(ctx, 5); err != nil {
		log.Fatal(err)
	}

	err = done()
	if err != nil {
		log.Println(err) // non-fatal for us
	}
}

func runTest(parentCtx context.Context, client *api.Client, browser api.Browser, sessionName string, mapping map[string]mappingPart) error {
	ctx, cancel := context.WithTimeout(parentCtx, time.Minute*5)
	defer cancel()

	caps := client.SetCaps(selenium.Capabilities{
		"browserstack.local": "true",
		"browserstack.video": "false",
		// "browserstack.debug":           "true",
		// "browserstack.console":         "errors",
		// "browserstack.networkLogs":     "errors",
		"build": sessionName,
		"name":  fmt.Sprintf("%s – %s", "Web Tests", browser.ResultKey()),
	})

	if browser.Device != "" {
		caps["deviceName"] = browser.Device
		caps["browserstack.appium_version"] = "1.8.0"
	}
	if browser.OS != "" {
		caps["platformName"] = browser.OS
	}
	if browser.OSVersion != "" {
		caps["platformVersion"] = browser.OSVersion
	}
	if browser.Browser != "" {
		caps["browserName"] = browser.Browser
	}
	if browser.BrowserVersion != "" {
		caps["browserVersion"] = browser.BrowserVersion
	}

	tests := []api.Test{}
	testPaths, err := getTestPaths()
	if err != nil {
		return err
	}

	for _, p := range testPaths {
		tests = append(tests, api.Test{
			Path: p,
		})
	}

	in := make(chan api.Test, len(tests))
	out := make(chan api.Test, len(tests))

	testResults := []api.Test{}

	go func() {
		for _, test := range tests {
			in <- test
		}

		close(in)
	}()

	go func() {
		for {
			select {
			case test, ok := <-out:
				if !ok {
					return
				}

				testResults = append(testResults, test)
				log.Println(browser.ResultKey(), test.Path, test.Success(), test.Duration())
			}
		}
	}()

	err = client.RunTest(ctx, caps, in, out)
	if err != nil {
		return err
	}

	for _, testResult := range testResults {
		err = writeResults(browser, testResult, mapping)
		if err != nil {
			return err
		}
	}

	return nil
}

func getTestPaths() ([]string, error) {
	var files []string

	err := filepath.Walk("./tests/", func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		if !strings.HasSuffix(path, ".html") {
			return nil
		}

		files = append(files, path)
		return nil
	})
	if err != nil {
		return nil, err
	}

	return files, nil
}

var fsMu = &sync.Mutex{} // TODO : clean this up
func writeResults(browser api.Browser, test api.Test, mapping map[string]mappingPart) error {
	fsMu.Lock()
	fsMu.Unlock()

	resultsPath := ""
	if itemIndex, ok := mapping[test.MappingID()].BySection[test.MappingSection()]; !ok {
		return errors.New("no mapping for test " + test.Path)
	} else {
		item := mapping[test.MappingID()].Items[itemIndex]
		resultsPath = filepath.Join(item.Path, test.ResultsFileName())
	}

	results := map[string]map[string]interface{}{}

	{
		f1, err := os.Open(resultsPath)
		if os.IsNotExist(err) {
			f1, err = os.Create(resultsPath)
		}
		if err != nil {
			return err
		}

		defer f1.Close()

		{
			b, err := ioutil.ReadAll(f1)
			if err != nil {
				return err
			}

			if len(b) > 0 {
				err = json.Unmarshal(b, &results)
				if err != nil {
					return err
				}
			}
		}

		var newScore float64
		if test.Success() {
			newScore = 1
		}

		if _, ok := results[browser.ResultKey()]; !ok {
			results[browser.ResultKey()] = map[string]interface{}{
				"browser":    browser.Browser,
				"version":    browser.BrowserVersion,
				"os":         browser.OS,
				"os_version": browser.OSVersion,
				"score":      newScore,
			}
		}

		if _, ok := results[browser.ResultKey()]["score"]; !ok {
			results[browser.ResultKey()]["score"] = newScore
		}

		var score float64
		v := results[browser.ResultKey()]["score"]
		if vv, ok := v.(float64); ok {
			score = vv
		}

		score = (score * 0.99) + (newScore * 0.01)

		if score > 1 {
			score = 1
		}

		if score < 0 {
			score = 0
		}

		results[browser.ResultKey()]["score"] = score
		results[browser.ResultKey()]["last_run"] = time.Now()

		err = f1.Close()
		if err != nil {
			return err
		}
	}

	{
		f2, err := os.Create(resultsPath)
		if err != nil {
			return err
		}

		defer f2.Close()

		b, err := json.MarshalIndent(results, "", "  ")
		if err != nil {
			return err
		}

		_, err = io.Copy(f2, bytes.NewBuffer(b))
		if err != nil {
			return err
		}

		err = f2.Close()
		if err != nil {
			return err
		}
	}

	return nil
}

func getMapping() (map[string]mappingPart, error) {
	f, err := os.Open("lib/mapping.json")
	if err != nil {
		return nil, err
	}

	defer f.Close()

	b, err := ioutil.ReadAll(f)
	if err != nil {
		return nil, err
	}

	out := map[string]mappingPart{}

	err = json.Unmarshal(b, &out)
	if err != nil {
		return nil, err
	}

	return out, nil
}

type mappingPart struct {
	BySection map[string]int `json:"bySection"`
	Items     []struct {
		Spec struct {
			Org     string `json:"org"`
			ID      string `json:"id"`
			Section string `json:"section"`
			Name    string `json:"name"`
			URL     string `json:"url"`
		} `json:"spec"`
		Tests map[string]string `json:"tests"`
		Path  string            `json:"path"`
	} `json:"items"`
}

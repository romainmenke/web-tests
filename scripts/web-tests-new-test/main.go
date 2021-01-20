package main

import (
	"bytes"
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"os"
	"path/filepath"
	"strings"
)

func main() {
	var org string
	var id string
	var section string
	var name string

	flag.StringVar(&org, "org", "", "Organisation authoring the spec")
	flag.StringVar(&id, "id", "", "Name of the specification")
	flag.StringVar(&section, "section", "", "Section of the feature (\"6.1.1\" for ecma262 Undefined)")
	flag.StringVar(&name, "name", "", "Name of the feature (single word)")

	flag.Parse()

	if org == "" {
		fmt.Println("-org is required\n\tweb-tests-new-test --help")
		return
	}

	if id == "" {
		fmt.Println("-id is required\n\tweb-tests-new-test --help")
		return
	}

	if section == "" {
		fmt.Println("-section is required\n\tweb-tests-new-test --help")
		return
	}

	if name == "" {
		fmt.Println("-name is required\n\tweb-tests-new-test --help")
		return
	}

	newFeatureDirPath := filepath.Join("./specifications/", org, id, fmt.Sprintf("%s.%s", section, name))
	dir, err := os.Stat(newFeatureDirPath)
	if err == nil {
		if dir.IsDir() {
			log.Fatal("Already exists")
		}
	}

	err = os.MkdirAll(newFeatureDirPath, os.ModePerm)
	if err != nil {
		log.Fatal(err)
	}

	exampleFeatureDirPath := filepath.Join("./specifications/example/test/1.feature")

	log.Println(newFeatureDirPath)
	log.Println(exampleFeatureDirPath)

	err = filepath.Walk(exampleFeatureDirPath, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		if strings.Contains(path, "results") {
			return filepath.SkipDir
		}

		newPath := strings.Replace(path, exampleFeatureDirPath, newFeatureDirPath, 1)
		if info.IsDir() {
			err = os.MkdirAll(newPath, os.ModePerm)
			if err != nil {
				return err
			}

			return nil
		}

		f1, err := os.Open(path)
		if err != nil {
			return err
		}

		defer f1.Close()

		b, err := ioutil.ReadAll(f1)
		if err != nil {
			return err
		}

		err = f1.Close()
		if err != nil {
			return err
		}

		if strings.HasSuffix(path, "meta.json") {
			meta := map[string]interface{}{}
			err = json.Unmarshal(b, &meta)
			if err != nil {
				return err
			}

			if v, ok := meta["spec"]; !ok {
				meta["spec"] = map[string]interface{}{
					"org":     org,
					"id":      id,
					"section": section,
					"name":    name,
				}
			} else if spec, ok := v.(map[string]interface{}); !ok {
				meta["spec"] = map[string]interface{}{
					"org":     org,
					"id":      id,
					"section": section,
					"name":    name,
				}
			} else {
				spec["org"] = org
				spec["id"] = id
				spec["section"] = section
				spec["name"] = name

				meta["spec"] = spec
			}

			meta["polyfill.io"] = []interface{}{}

			b, err = json.MarshalIndent(meta, "", "  ")
			if err != nil {
				return err
			}
		}

		f2, err := os.Create(newPath)
		if err != nil {
			return err
		}

		defer f2.Close()

		_, err = io.Copy(f2, bytes.NewBuffer(b))
		if err != nil {
			return err
		}

		err = f2.Close()
		if err != nil {
			return err
		}

		return nil
	})
	if err != nil {
		log.Fatal(err)
	}
}
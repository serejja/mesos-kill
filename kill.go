/* Licensed to the Apache Software Foundation (ASF) under one or more
contributor license agreements.  See the NOTICE file distributed with
this work for additional information regarding copyright ownership.
The ASF licenses this file to You under the Apache License, Version 2.0
(the "License"); you may not use this file except in compliance with
the License.  You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License. */

package main

import (
	"bufio"
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
	"regexp"
	"strings"
)

const defaultMasterLocation = "127.0.0.1:5050"
const mesosMasterEnv = "MESOS_MASTER"

type Framework struct {
	Name  string
	ID    string
	Tasks int
}

func main() {
	master, framework := parseArgs()
	frameworks, err := findMatchingFrameworks(master, framework)
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
	if len(frameworks) == 0 {
		fmt.Printf("No frameworks matching '%s' found.\n", framework)
		os.Exit(0)
	}

	err = proposeKillFrameworks(master, frameworks)
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}

func parseArgs() (string, string) {
	args := os.Args
	if len(args) == 1 {
		printUsageAndExit(1)
	} else if len(args) == 2 {
		master := os.Getenv(mesosMasterEnv)
		if master == "" {
			master = defaultMasterLocation
		}

		return master, args[1]
	}

	return args[1], args[2]
}

func printUsageAndExit(status int) {
	fmt.Println("Usage: mesos-kill [<master>] <framework-name-regex>")
	fmt.Println()
	fmt.Printf("<master>: host:port pair for Mesos Master node. If not specified will check %s env and fall back to %s if not set\n", mesosMasterEnv, defaultMasterLocation)
	fmt.Println("<framework-name-regex>: name or regular expression of framework to kill. It is ok to match multiple frameworks")
	os.Exit(status)
}

func findMatchingFrameworks(master string, framework string) ([]*Framework, error) {
	frameworkRegex, err := regexp.Compile(framework)
	if err != nil {
		return nil, err
	}

	rawResponse, err := http.Get(fmt.Sprintf("http://%s/state.json", master))
	if err != nil {
		return nil, err
	}

	var responseBody []byte
	responseBody, err = ioutil.ReadAll(rawResponse.Body)
	if err != nil {
		return nil, err
	}

	responseJson := make(map[string]interface{})
	err = json.Unmarshal(responseBody, &responseJson)
	if err != nil {
		return nil, err
	}

	frameworks, ok := responseJson["frameworks"].([]interface{})
	if !ok {
		return nil, fmt.Errorf("frameworks node should be []interface{}, actual %T", responseJson["frameworks"])
	}

	matchingFrameworks := make([]*Framework, 0)
	for _, framework := range frameworks {
		var frameworkMap map[string]interface{}
		frameworkMap, ok = framework.(map[string]interface{})
		if !ok {
			return nil, fmt.Errorf("framework node should be map[string]interface{}, actual %T", framework)
		}

		var name string
		var id string
		var tasks []interface{}

		name, ok = frameworkMap["name"].(string)
		if !ok {
			return nil, fmt.Errorf("framework name node should be string, actual %T", frameworkMap["name"])
		}
		id, ok = frameworkMap["id"].(string)
		if !ok {
			return nil, fmt.Errorf("framework id node should be string, actual %T", frameworkMap["id"])
		}
		tasks, ok = frameworkMap["tasks"].([]interface{})
		if !ok {
			return nil, fmt.Errorf("framework tasks node should be []interface{}, actual %T", frameworkMap["tasks"])
		}

		if frameworkRegex.MatchString(name) {
			matchingFrameworks = append(matchingFrameworks, &Framework{
				Name:  name,
				ID:    id,
				Tasks: len(tasks),
			})
		}
	}

	return matchingFrameworks, nil
}

func proposeKillFrameworks(master string, frameworks []*Framework) error {
	for _, framework := range frameworks {
		err := proposeKillFramework(master, framework)
		if err != nil {
			return err
		}
	}

	return nil
}

func proposeKillFramework(master string, framework *Framework) error {
	reader := bufio.NewReader(os.Stdin)
	fmt.Printf("Kill framework %s (ID %s) with %d tasks running? ([y]es, [n]o): ", framework.Name, framework.ID, framework.Tasks)
	text, _ := reader.ReadString('\n')

	if len(text) != 2 {
		return proposeKillFramework(master, framework)
	}

	switch strings.ToLower(string(text[0])) {
	case "y":
		return killFramework(master, framework.ID)
	case "n":
	default:
		return proposeKillFramework(master, framework)
	}

	return nil
}

func killFramework(master string, id string) error {
	_, err := http.Post(fmt.Sprintf("http://%s/teardown", master), "", bytes.NewReader([]byte(fmt.Sprintf("frameworkId=%s", id))))
	if err != nil {
		return err
	}
	fmt.Printf("\nKilled framework %s\n", id)

	return nil
}

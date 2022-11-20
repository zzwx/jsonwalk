// An example of analyzing a docker inspect of all the existing containers to retrieve just the needed values.
package main

import (
	"bytes"
	_ "embed"
	"encoding/json"
	"fmt"
	"log"
	"os/exec"
	"strings"

	"github.com/zzwx/jsonwalk"
)

func main() {
	cmd := exec.Command("docker", "ps", "-aq")
	var outb, errb bytes.Buffer
	cmd.Stdout = &outb
	cmd.Stderr = &errb
	err := cmd.Run()
	if err != nil {
		log.Fatalln(err, errb.String())
	}
	spl := strings.Split(outb.String(), "\n")
	first := true
	for _, s := range spl {
		s = strings.TrimSpace(s)
		if s != "" {
			if !first {
				fmt.Println()
			}
			readInsp(s)
			first = false
		}
	}
}

func readInsp(cont string) {
	cmd := exec.Command("docker", "inspect", cont)
	var outb, errb bytes.Buffer
	cmd.Stdout = &outb
	cmd.Stderr = &errb
	err := cmd.Run()
	if err != nil {
		log.Fatalln(err, errb.String())
	}
	analyze(outb.Bytes())
}

func analyze(src []byte) {
	var v interface{}
	err := json.Unmarshal(src, &v)
	if err != nil {
		panic(err)
	}
	//jsonwalk.Walk(&v, jsonwalk.Print{})

	name := ""
	hostName := ""
	status := ""
	var env []string
	workingDir := ""
	warning := ""
	image := ""

	jsonwalk.Walk(&v, jsonwalk.Callback(func(path jsonwalk.WalkPath, key interface{}, value interface{}, tp jsonwalk.NodeValueType) {
		if path.Path() == "[0].Config.Env" && tp == jsonwalk.Array {
			for _, v := range value.([]interface{}) {
				env = append(env, v.(string))
			}
		} else if (path.Path() == "[0].Config.Hostname") && tp == jsonwalk.String {
			hostName = value.(string)
		} else if (path.Path() == "[0].Name") && tp == jsonwalk.String {
			name = value.(string)
		} else if (path.Path() == "[0].State.Status") && tp == jsonwalk.String {
			status = value.(string)
		} else if (path.Path() == "[0].Config.WorkingDir") && tp == jsonwalk.String {
			workingDir = value.(string)
		} else if (path.Path() == "[0].Config.Image") && tp == jsonwalk.String {
			image = value.(string)
		} else if strings.HasPrefix(path.Path(), "[1]") {
			warning = "--- warning: [1] found in the array"
		}

	}))

	fmt.Printf("%v\n", image)
	fmt.Printf("%v | %v | %v | %v\n", name, hostName, status, workingDir)
	for _, e := range env {
		fmt.Printf("%v\n", e)
	}
	if warning != "" {
		fmt.Printf("%v\n", warning)
	}

}

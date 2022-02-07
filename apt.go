//
//  This file is part of go-apt-client library
//
//  Copyright (C) 2017  Arduino AG (http://www.arduino.cc/)
//
//  Licensed under the Apache License, Version 2.0 (the "License");
//  you may not use this file except in compliance with the License.
//  You may obtain a copy of the License at
//
//    http://www.apache.org/licenses/LICENSE-2.0
//
//  Unless required by applicable law or agreed to in writing, software
//  distributed under the License is distributed on an "AS IS" BASIS,
//  WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
//  See the License for the specific language governing permissions and
//  limitations under the License.
//

package apt

import (
	"bufio"
	"bytes"
	"fmt"
	"os/exec"
	"regexp"
	"strconv"
	"strings"

	"github.com/pkg/errors"
)

// Package is a package available in the APT system
type Package struct {
	Name             string
	Status           string
	Architecture     string
	Version          string
	ShortDescription string
	InstalledSizeKB  int
}

// List returns a list of packages available in the system with their
// respective status.
func List() ([]*Package, error) {
	return Search("*")
}

// Search list packages available in the system that match the search
// pattern
func Search(pattern string) ([]*Package, error) {
	executer = exec.Command("dpkg-query", "-W", "-f=${Package}\t${Architecture}\t${db:Status-Status}\t${Version}\t${Installed-Size}\t${Binary:summary}\n", pattern)

	out, err := executer.CombinedOutput()
	if err != nil {
		// Avoid returning an error if the list is empty
		if bytes.Contains(out, []byte("no packages found matching")) {
			return []*Package{}, nil
		}
		errMsg := fmt.Sprintf("running dpkg-query: %s", out)
		return nil, errors.Wrap(err, errMsg)
	}

	return parseDpkgQueryOutput(out), nil
}

func parseDpkgQueryOutput(out []byte) []*Package {
	res := []*Package{}
	scanner := bufio.NewScanner(bytes.NewReader(out))
	for scanner.Scan() {
		data := strings.Split(scanner.Text(), "\t")
		size, err := strconv.Atoi(data[4])
		if err != nil {
			// Ignore error
			size = 0
		}
		res = append(res, &Package{
			Name:             data[0],
			Architecture:     data[1],
			Status:           data[2],
			Version:          data[3],
			InstalledSizeKB:  size,
			ShortDescription: data[5],
		})
	}
	return res
}

// CheckForUpdates runs an apt update to retrieve new packages available
// from the repositories
func CheckForUpdates() error {
	executer = exec.Command("apt-get", "update", "-q")
	return executer.Run()
}

// ListUpgradable return all the upgradable packages and the version that
// is going to be installed if an UpgradeAll is performed
func ListUpgradable() ([]*Package, error) {
	pkgs := []*Package{}

	executer = exec.Command("apt", "list", "--upgradable")

	out, err := executer.Output()
	if err != nil {
		return nil, errors.Wrap(err, "error running apt list")
	}
	re := regexp.MustCompile(`^([^ ]+) ([^ ]+) ([^ ]+)( \[upgradable from: [^\[\]]*\])?`)

	scanner := bufio.NewScanner(bytes.NewReader(out))
	for scanner.Scan() {
		matches := re.FindAllStringSubmatch(scanner.Text(), -1)
		if len(matches) == 0 {
			continue
		}

		// Remove repository information in name
		// example: "libgweather-common/zesty-updates,zesty-updates"
		//       -> "libgweather-common"
		name := strings.Split(matches[0][1], "/")[0]

		pkgs = append(pkgs, &Package{
			Name:         name,
			Status:       "upgradable",
			Version:      matches[0][2],
			Architecture: matches[0][3],
		})
	}
	return pkgs, nil
}

// Upgrade runs the upgrade for a set of packages
func Upgrade(packs ...*Package) (err error) {
	args := []string{"upgrade", "-y"}
	for _, pack := range packs {
		if pack == nil || pack.Name == "" {
			return errors.New("invalid package with empty name")
		}
		args = append(args, pack.Name)
	}
	executer = exec.Command("apt-get", args...)
	return executer.Run()
}

// UpgradeAll upgrade all upgradable packages
func UpgradeAll() (err error) {
	executer = exec.Command("apt-get", "upgrade", "-y")
	return executer.Run()
}

// DistUpgrade upgrades all upgradable packages, it may remove older versions to install newer ones.
func DistUpgrade() (err error) {
	executer = exec.Command("apt-get", "dist-upgrade", "-y")
	return executer.Run()
}

// Remove removes a set of packages
func Remove(packs ...*Package) error {
	args := []string{"remove", "-y"}
	for _, pack := range packs {
		if pack == nil || pack.Name == "" {
			return errors.New("invalid package with empty name")
		}
		args = append(args, pack.Name)
	}
	executer = exec.Command("apt-get", args...)
	return executer.Run()
}

// Install installs a set of packages
func Install(packs ...*Package) error {
	args := []string{"install", "-y"}
	for _, pack := range packs {
		if pack == nil || pack.Name == "" {
			return errors.New("invalid package with empty name")
		}
		args = append(args, pack.Name)
	}
	executer = exec.Command("apt-get", args...)
	return executer.Run()
}

// Copyright 2015 ThoughtWorks, Inc.

// This file is part of getgauge/xml-report.

// getgauge/xml-report is free software: you can redistribute it and/or modify
// it under the terms of the GNU General Public License as published by
// the Free Software Foundation, either version 3 of the License, or
// (at your option) any later version.

// getgauge/xml-report is distributed in the hope that it will be useful,
// but WITHOUT ANY WARRANTY; without even the implied warranty of
// MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
// GNU General Public License for more details.

// You should have received a copy of the GNU General Public License
// along with getgauge/xml-report.  If not, see <http://www.gnu.org/licenses/>.

package main

import (
	"errors"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/getgauge/common"
	"github.com/getgauge/xml-report/builder"
	"github.com/getgauge/xml-report/gauge_messages"
	"github.com/getgauge/xml-report/listener"
)

const (
	defaultReportsDir           = "reports"
	gaugeReportsDirEnvName      = "gauge_reports_dir" // directory where reports are generated by plugins
	EXECUTION_ACTION            = "execution"
	GAUGE_HOST                  = "127.0.0.1"
	GAUGE_PORT_ENV              = "plugin_connection_port"
	PLUGIN_ACTION_ENV           = "xml-report_action"
	xmlReport                   = "xml-report"
	overwriteReportsEnvProperty = "overwrite_reports"
	resultFile                  = "result.xml"
	timeFormat                  = "2006-01-02 15.04.05"
)

var projectRoot string
var pluginDir string

func createReport(suiteResult *gauge_messages.SuiteExecutionResult) {
	dir := createReportsDirectory()
	bytes, err := builder.NewXmlBuilder(0).GetXmlContent(suiteResult)
	if err != nil {
		fmt.Printf("Report generation failed: %s \n", err)
		os.Exit(1)
	}
	err = writeResultFile(dir, bytes)
	if err != nil {
		fmt.Printf("Report generation failed: %s \n", err)
		os.Exit(1)
	}
	fmt.Printf("Successfully generated xml-report to => %s\n", dir)
}

func writeResultFile(reportDir string, bytes []byte) error {
	resultPath := filepath.Join(reportDir, resultFile)
	err := ioutil.WriteFile(resultPath, bytes, common.NewFilePermissions)
	if err != nil {
		return errors.New(fmt.Sprintf("Failed to copy file: %s %s\n", resultFile, err))
	}
	return nil
}

func createExecutionReport() {
	os.Chdir(projectRoot)
	listener, err := listener.NewGaugeListener(GAUGE_HOST, os.Getenv(GAUGE_PORT_ENV))
	if err != nil {
		fmt.Println("Could not create the gauge listener")
		os.Exit(1)
	}
	listener.OnSuiteResult(createReport)
	listener.Start()
}

func findPluginAndProjectRoot() {
	projectRoot = os.Getenv(common.GaugeProjectRootEnv)
	if projectRoot == "" {
		fmt.Printf("Environment variable '%s' is not set. \n", common.GaugeProjectRootEnv)
		os.Exit(1)
	}
	var err error
	pluginDir, err = os.Getwd()
	if err != nil {
		fmt.Printf("Error finding current working directory: %s \n", err)
		os.Exit(1)
	}
}

func createReportsDirectory() string {
	reportsDir, err := filepath.Abs(os.Getenv(gaugeReportsDirEnvName))
	if reportsDir == "" || err != nil {
		reportsDir = defaultReportsDir
	}
	currentReportDir := filepath.Join(reportsDir, xmlReport, getNameGen().randomName())
	createDirectory(currentReportDir)
	return currentReportDir
}

func createDirectory(dir string) {
	if common.DirExists(dir) {
		return
	}
	if err := os.MkdirAll(dir, common.NewDirectoryPermissions); err != nil {
		fmt.Printf("Failed to create directory %s: %s\n", defaultReportsDir, err)
		os.Exit(1)
	}
}

func getNameGen() nameGenerator {
	if shouldOverwriteReports() {
		return emptyNameGenerator{}
	}
	return timeStampedNameGenerator{}
}

type nameGenerator interface {
	randomName() string
}
type timeStampedNameGenerator struct{}

func (T timeStampedNameGenerator) randomName() string {
	return time.Now().Format(timeFormat)
}

type emptyNameGenerator struct{}

func (T emptyNameGenerator) randomName() string {
	return ""
}

func shouldOverwriteReports() bool {
	envValue := os.Getenv(overwriteReportsEnvProperty)
	if strings.ToLower(envValue) == "true" {
		return true
	}
	return false
}

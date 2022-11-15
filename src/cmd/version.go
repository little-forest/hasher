/*
Copyright © 2022 Yusuke KOMORI

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

	http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/
package cmd

import (
	"fmt"
	"time"

	"github.com/spf13/cobra"
)

// Will be fixed by LDFLAGS
var (
	version  = "dev"
	revision = "dev"
	date     = "unknown"
	osArch   = "unknown"
)

// versionCmd represents the version command
var versionCmd = &cobra.Command{
	Use:   "version",
	Short: "show version",
	Run:   runVersion,
}

func init() {
	rootCmd.AddCommand(versionCmd)
}

func runVersion(cmd *cobra.Command, args []string) {
	fmt.Printf("hasher version %s %s built from %s on %s\n", version, osArch, revision, date)

	// This is HasherProgressViewer's test code
	// testSingleProcess(true)
	// testMultiProgress(true)
}

func testSingleProcess(showError bool) {
	numOfWorkers := 1
	viewer := NewHasherProgressViewer(numOfWorkers, true)

	total := 3
	viewer.Setup()
	viewer.SetTotal(total)
	sleep()

	viewer.TaskStart(0, "task1")
	sleep()
	viewer.TaskDone(0, 1, total, "[OK]")
	sleep()

	viewer.TaskStart(0, "task2")
	sleep()
	if showError {
		viewer.ShowError("error xxxx")
		sleep()
		viewer.ShowError("error yyyy")
		sleep()
	}
	viewer.TaskDone(0, 2, total, "[NG]")
	sleep()

	viewer.TaskStart(0, "task3")
	sleep()
	viewer.TaskDone(0, 3, total, "[OK]")
	sleep()

	viewer.TearDown()
}

func testMultiProgress(showError bool) {
	numOfWorkers := 3
	viewer := NewHasherProgressViewer(numOfWorkers, true)

	total := 6
	done := 0
	viewer.Setup()
	viewer.SetTotal(total)
	sleep()

	viewer.TaskStart(0, "task1")
	sleep()

	viewer.TaskStart(2, "task3")
	sleep()

	viewer.TaskStart(1, "task2")
	sleep()

	done++
	viewer.TaskDone(2, done, total, "[OK]")
	sleep()

	done++
	viewer.TaskDone(0, done, total, "[OK]")
	sleep()

	viewer.TaskStart(0, "task4")
	sleep()

	if showError {
		viewer.ShowError("task2 : error")
		sleep()
	}
	done++
	viewer.TaskDone(1, done, total, "[OK]")
	sleep()

	viewer.TaskStart(1, "task5")
	sleep()

	done++
	viewer.TaskDone(1, done, total, "[OK]")
	sleep()

	viewer.TaskStart(2, "task6")
	sleep()

	if showError {
		viewer.ShowError("task6 : error")
		sleep()
	}

	done++
	viewer.TaskDone(0, done, total, "[OK]")
	sleep()

	done++
	viewer.TaskDone(2, done, total, "[OK]")
	sleep()

	viewer.TearDown()
}

func sleep() {
	time.Sleep(time.Millisecond * 500)
}

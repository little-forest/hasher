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

	// This is HasherProgressNotifier's test code
	// testSingleProcess(true)
	// testMultiProgress(true)
}

func testSingleProcess(showError bool) {
	numOfWorkers := 1
	viewer := NewHasherProgressNotifier(numOfWorkers, true)

	total := 3
	viewer.Start()
	viewer.SetTotal(total)
	sleep()

	viewer.NotifyTaskStart(0, "task1")
	sleep()
	viewer.NotifyTaskDone(0, "[OK]")
	viewer.NotifyProgress(1, total)
	sleep()

	viewer.NotifyTaskStart(0, "task2")
	sleep()
	if showError {
		viewer.NotifyError(0, "error xxxx")
		sleep()
		viewer.NotifyError(0, "error yyyy")
		sleep()
	}
	viewer.NotifyTaskDone(0, "[NG]")
	viewer.NotifyProgress(2, total)
	sleep()

	viewer.NotifyTaskStart(0, "task3")
	sleep()
	viewer.NotifyTaskDone(0, "[OK]")
	viewer.NotifyProgress(3, total)
	sleep()

	viewer.Shutdown()
}

func testMultiProgress(showError bool) {
	numOfWorkers := 3
	viewer := NewHasherProgressNotifier(numOfWorkers, true)

	total := 6
	done := 0
	viewer.Start()
	viewer.SetTotal(total)
	sleep()

	viewer.NotifyTaskStart(0, "task1")
	sleep()

	viewer.NotifyTaskStart(2, "task3")
	sleep()

	viewer.NotifyTaskStart(1, "task2")
	sleep()

	done++
	viewer.NotifyTaskDone(2, "[OK]")
	viewer.NotifyProgress(done, total)
	sleep()

	done++
	viewer.NotifyTaskDone(0, "[OK]")
	viewer.NotifyProgress(done, total)
	sleep()

	viewer.NotifyTaskStart(0, "task4")
	sleep()

	if showError {
		viewer.NotifyError(1, "task2 : error")
		sleep()
	}
	done++
	viewer.NotifyTaskDone(1, "[OK]")
	viewer.NotifyProgress(done, total)
	sleep()

	viewer.NotifyTaskStart(1, "task5")
	sleep()

	done++
	viewer.NotifyTaskDone(1, "[OK]")
	viewer.NotifyProgress(done, total)
	sleep()

	viewer.NotifyTaskStart(2, "task6")
	sleep()

	if showError {
		viewer.NotifyError(2, "task6 : error")
		sleep()
	}

	done++
	viewer.NotifyTaskDone(0, "[OK]")
	viewer.NotifyProgress(done, total)
	sleep()

	done++
	viewer.NotifyTaskDone(2, "[OK]")
	viewer.NotifyProgress(done, total)
	sleep()

	viewer.Shutdown()
}

func sleep() {
	time.Sleep(time.Millisecond * 500)
}

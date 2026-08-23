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
package term

import (
	"fmt"
	"os"

	"github.com/morikuni/aec"
)

// ANSI escape sequences
//
//	see: https://github.com/morikuni/aec
var C_default = aec.EmptyBuilder.DefaultF().ANSI
var C_green = aec.EmptyBuilder.GreenF().ANSI
var C_red = aec.EmptyBuilder.RedF().ANSI
var C_lred = aec.EmptyBuilder.LightRedF().ANSI
var C_blue = aec.EmptyBuilder.BlueF().ANSI
var C_yellow = aec.EmptyBuilder.YellowF().ANSI
var C_cyan = aec.EmptyBuilder.CyanF().ANSI
var C_white = aec.EmptyBuilder.WhiteF().ANSI
var C_gray = aec.EmptyBuilder.Color8BitF(8).ANSI

var C_pink = aec.EmptyBuilder.Color8BitF(218).ANSI
var C_malibu = aec.EmptyBuilder.Color8BitF(74).ANSI
var C_orange = aec.EmptyBuilder.Color8BitF(214).ANSI
var C_darkorange3 = aec.EmptyBuilder.Color8BitF(166).ANSI
var C_lime = aec.EmptyBuilder.Color8BitF(10).ANSI

var Mark_OK = fmt.Sprintf("[%s]", C_green.Apply("OK"))
var Mark_Error = fmt.Sprintf("[%s]", C_red.Apply("ERROR"))
var Mark_Failed = fmt.Sprintf("[%s]", C_red.Apply("FAILED"))
var Mark_Warning = fmt.Sprintf("[%s]", C_yellow.Apply("WARNING"))
var Mark_Updated = fmt.Sprintf("[%s]", C_green.Apply("UPDATE"))

func ShowWarn(format string, args ...interface{}) {
	fmt.Fprintf(os.Stderr, "["+C_yellow.Apply("WARNING")+"] "+format+"\n", args...)
}

func ShowError(err error) {
	fmt.Fprintln(os.Stderr, "["+C_red.Apply("ERROR")+"] "+err.Error())
}

func ShowErrorMsg(format string, args ...interface{}) {
	fmt.Fprintf(os.Stderr, "["+C_red.Apply("ERROR")+"] "+format+"\n", args...)
}

func ShowCursor() {
	fmt.Print("\x1b[?25h")
}

func HideCursor() {
	fmt.Print("\x1b[?25l")
}

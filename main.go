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
package main

import (
	"context"
	_ "crypto/sha1"
	"os"
	"os/signal"
	"syscall"

	"github.com/little-forest/hasher/cmd"
)

func main() {
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	// signal.NotifyContext hands cmd a cancellable context, but it also
	// disables the default disposition for these signals — and nothing below
	// cmd takes a context yet, so cancelling it cannot unwind the work in
	// progress. Without this the process would ignore Ctrl-C entirely.
	// Watch the signals directly rather than ctx.Done(), which the deferred
	// stop() also closes on a normal exit. Once the hashing layer honours
	// cmd.Context(), this can go away and the interrupted run can be unwound
	// properly instead.
	go func() {
		sig := make(chan os.Signal, 1)
		signal.Notify(sig, os.Interrupt, syscall.SIGTERM)
		<-sig
		os.Exit(130)
	}()

	if err := cmd.Execute(ctx); err != nil {
		os.Exit(1)
	}
}

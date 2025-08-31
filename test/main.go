package main

import (
	"encoding/json"
	"os"
	"time"

	"github.com/elisiei/zlog"
)

func main() {
	// create a new logger
	l := zlog.New()
	l.SetOutput(os.Stdout)        // log to stdout
	l.SetLevel(zlog.LevelDebug)   // show debug and up
	l.SetTimeFormat(time.Kitchen) // nicer time format
	l.EnableColors(true)          // force colors on
	l.ShowCaller(true)            // show file:line
	l.Infof("starting application...")

	// structured fields
	l.WithFields(zlog.F{"user": "yehorovye", "id": 42}).Infow("user login", zlog.F{
		"ip": "127.0.0.1",
	})

	// debug log
	l.Debug("this is a debug message")

	// warn with structured data
	l.Warnw("quota running low", zlog.F{
		"quota": "100MB",
		"used":  "95MB",
	})

	type TestStruct struct {
		Ok      bool   `json:"ok"`
		Message string `json:"message"`
	}

	testStruct, _ := json.Marshal(TestStruct{Ok: true, Message: "welcome"})

	l.Debugf("%v", string(testStruct))

	// switch to json output
	l.SetJSON(true)
	l.Infof("switched to json mode")

	// add global fields (context)
	ctxLogger := l.WithFields(zlog.F{"session": "abcd1234"})
	ctxLogger.Infow("session started", zlog.F{"feature": "search"})

	// demonstrate global std logger
	zlog.Infof("hello from std logger")
	zlog.Warnw("deprecated API used", zlog.F{"api": "/v1/search"})

	// error!!!
	zlog.Error("something bad happened, but i am ok :D")

	// fatal will exit the app
	zlog.Fatal("something unrecoverable happened")

	// zlog.Fatalf("%v", 100percentanerror)
}

package main

import (
	"encoding/json"
	"os"
	"time"

	"github.com/elisiei/zlog"
)

func main() {
	l := zlog.New()
	l.SetOutput(os.Stdout)        // send logs to stdout
	l.SetLevel(zlog.LevelDebug)   // show debug and above
	l.SetTimeFormat(time.Kitchen) // human readable time
	l.EnableColors(true)          // force colors
	l.ShowCaller(true)            // show file:line
	l.Infof("starting application...")

	// call raw logs
	l.Log(zlog.LevelInfo, "raw log", nil)

	// add structured fields
	l.WithFields(zlog.F{"user": "yehorovye", "id": 42}).Infow("user login", zlog.F{
		"ip": "127.0.0.1",
	})

	// plain debug log
	l.Debug("this is a debug message")

	// warn with fields
	l.Warnw("quota running low", zlog.F{
		"quota": "100MB",
		"used":  "95MB",
	})

	// log json data
	type TestStruct struct {
		Ok      bool   `json:"ok"`
		Message string `json:"message"`
	}
	testStruct, _ := json.Marshal(TestStruct{Ok: true, Message: "welcome"})
	l.Debugf("%v", string(testStruct))

	// switch to json output
	l.SetJSON(true)
	l.Infof("switched to json mode")

	// add global fields (ctx style)
	ctxLogger := l.WithFields(zlog.F{"session": "abcd1234"})
	ctxLogger.Infow("session started", zlog.F{"feature": "search"})

	// use global std logger
	zlog.Infof("hello from std logger")
	zlog.Warnw("deprecated api used", zlog.F{"api": "/v1/search"})

	// change level whenever u want
	zlog.SetLevel(zlog.LevelDebug)

	// new! :D format logs with extra fields
	zlog.Debugw("completed task: %s", zlog.F{"time": "200ms"}, "run")

	// also can mix format + fields
	zlog.Debugf("no task named '%s'", "run")

	// error log
	zlog.Error("something bad happened, but i am ok :D")

	// fatal log (exits app)
	zlog.Fatal("something unrecoverable happened")
}

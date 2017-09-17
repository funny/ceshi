package utest

import (
	"fmt"
	"io/ioutil"
	"os"
	"runtime/pprof"
	"strings"
	"sync"
	"time"

	"github.com/pkg/errors"
)

type commandHandler func(string) (string /*debug output*/, error)

var (
	filemap       map[string]*os.File
	cmdHandlerMap map[string]commandHandler
	lock          sync.Mutex
)

func RegisterCmdHandler(fileName string, Handler func(string) (string, error)) error {
	lock.Lock()
	defer lock.Unlock()

	var (
		err error
	)
	fileName = strings.TrimSpace(fileName)
	if len(fileName) <= 0 {
		err = errors.Errorf("file name can't be empy name")
		return err
	}
	if cmdHandlerMap == nil {
		cmdHandlerMap = make(map[string]commandHandler)
	}
	cmdHandlerMap[fileName] = Handler
	return err
}

func init() {
	go func() {
		defer func() {
			for name, file := range filemap {
				if err := file.Close(); err != nil {
					fmt.Println("%s closes failed. error: %s", name, err.Error())
				}
			}
		}()
		for {
			if input, err := ioutil.ReadFile("utest.cmd"); err == nil && len(input) > 0 {
				ioutil.WriteFile("utest.cmd", []byte(""), 0744)

				cmd := strings.Trim(string(input), " \n\r\t")

				var (
					profile  *pprof.Profile
					filename string
				)

				switch cmd {
				case "lookup goroutine":
					profile = pprof.Lookup("goroutine")
					filename = "utest.goroutine"
				case "lookup heap":
					profile = pprof.Lookup("heap")
					filename = "utest.heap"
				case "lookup threadcreate":
					profile = pprof.Lookup("threadcreate")
					filename = "utest.thread"
				default:
					if !strings.HasPrefix(cmd, "lookup") {
						fmt.Println("only support `lookup` command right now.")
					}
					params := strings.Split(cmd, " ")
					if len(params) <= 1 {
						fmt.Println("at least need to two or more params")
					} else {
						if err = cmdHandler(params[1]); err != nil {
							println(err.Error())
						}
					}
				}

				if profile != nil {
					file, err := openFile(filename)
					if err != nil {
						println("couldn't create " + filename)
					} else {
						profile.WriteTo(file, 2)
					}
				}
			}
			time.Sleep(2 * time.Second)
		}
	}()
}

func openFile(fileName string) (*os.File, error) {
	var (
		err error
	)
	if filemap == nil {
		filemap = make(map[string]*os.File)
	}
	fileName = strings.TrimSpace(fileName)
	if len(fileName) <= 0 {
		err = errors.Errorf("file name can't be empty name")
		return nil, err
	}
	if _, ok := filemap[fileName]; !ok {
		filemap[fileName], err = os.Create(fileName)
		if err != nil {
			err = errors.Errorf("os creates `%s` failed. error: %s",
				fileName,
				err.Error())
			return nil, err
		}
	}
	return filemap[fileName], err
}

// i like this. `func cmdHandler(fileName string) (err error)`
func cmdHandler(fileName string) error {
	var (
		err    error
		output string
		file   *os.File
	)
	fileName = strings.TrimSpace(fileName)
	if len(fileName) <= 0 {
		err = errors.Errorf("file name can't be empy name")
		return err
	}
	if handler, ok := cmdHandlerMap[fileName]; !ok {
		err = errors.Errorf("handler command `%s` need to register", fileName)
		return err
	} else {
		if output, err = handler(fileName); err != nil {
			return err
		}
		if file, err = openFile(fileName); err != nil {
			return err
		}
		if _, err = file.Write([]byte(output)); err != nil {
			return err
		}
	}
	return err
}

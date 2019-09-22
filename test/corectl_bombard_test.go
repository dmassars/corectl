// +build integration

package test

import (
	"bytes"
	"fmt"
	"os"
	"io/ioutil"
	"sync"
	"testing"

	"github.com/qlik-oss/corectl/test/toolkit"
)

var temp = "test/temp"
var scripts []string

func setUp(parll int) error {
  err := os.Mkdir(temp, 0755)
	if err != nil {
		return err
	}
	scripts = make([]string, parll)
  for i := 0; i < parll; i++ {
    f, err := ioutil.TempFile(temp, "*.qvs")
    if err != nil {
			return err
		}
    f.Close()
    scripts[i] = f.Name()
  }
	return nil
}

func tearDown() {
  os.RemoveAll(temp)
}

func TestScript(t *testing.T) {
	var parll, runs = 12, 10
	err := setUp(parll)
	defer tearDown()
	if err != nil {
		fmt.Println(err)
		t.Fail()
	}
	p := toolkit.Params{T: t, Engine: *toolkit.EngineStdIP}
	defer p.Reset()
	mux := sync.Mutex{}
	logs := [][]byte{}
	ch := make(chan bool)
	for i := 0; i < parll; i++ {
		go func (id int) {
			for j := 0; j < runs; j++ {
				name := fmt.Sprintf("%s_%d_%d", t.Name(), id, j)
				mux.Lock()
				out1 := p.ExpectOK().Run("-vt", "-a", name, "build", "--ttl=0")
				mux.Unlock()
				out2 := p.ExpectOK().Run("-vt", "-a", name, "script", "set", scripts[id], "--ttl=0")
				out3 := p.ExpectOK().Run("-vt", "-a", name, "script", "get", "--ttl=0")
				mux.Lock()
				logs = append(logs, out1, out2, out3)
				mux.Unlock()
			}
			ch<-true
		}(i)
	}
	for i := 0; i < parll; i++ {
		<-ch
	}
	sub := []byte("SESSION_ATTACHED")
	c := 0
	for _, l := range logs {
		if bytes.Contains(l, sub) {
			c++
		}
	}
	if c > 0 {
		fmt.Printf("Got attached %d times.", c)
		t.Fail()
	}
}

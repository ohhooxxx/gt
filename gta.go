package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"io/ioutil"
	"os"
	"os/exec"
	"sync"

	"github.com/mitchellh/go-homedir"
)

var home, err = homedir.Dir()
var store = initStorage(home + "/.gt-storage")

/*****************************************
*                                        *
*                  init                  *
*                                        *
*****************************************/
func init() {
	flag.String("a", "", "Add an alias")
	flag.String("d", "", "Delete alias")
	flag.String("r", "", "Rename the alias")
	flag.String("c", "", "Clear all alias")
	flag.String("l", "", "List of Aliases")
}

/*****************************************
*                                        *
*                  main                  *
*                                        *
*****************************************/
func main() {
	flag.Parse()
	flag.Visit(action)

	f := flag.NewFlagSet("f", flag.ContinueOnError)
	arg := flag.Arg(0)

	if flag.NFlag() == 0 && len(arg) == 0 {
		fmt.Fprint(f.Output(), store.aliasList(), "\n")
	}

	if flag.NFlag() == 0 && len(arg) != 0 {
		path := store.getAlias(arg)
		if path != nil {
			cdDir(path.(string))
		}
	}

}

func cdDir(path string) {
	cmd := exec.Command("cd", path)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	cmd.Run()
}

func action(f *flag.Flag) {
	arg := flag.Arg(0)

	switch f.Name {
	case "a":
		cur, err := os.Getwd()
		if err != nil {
			panic(err)
		}
		if len(arg) != 0 {
			cur = arg
		}
		store.addAlias(f.Value.String(), cur)
	case "d":
		store.delAlias(f.Value.String())
	case "r":
		store.renameAlias(f.Value.String(), arg)
	case "c":
		store.clearAlias()
	case "l":
		store.aliasList()
	default:
		fmt.Printf("No %s command.", f.Name)
	}
}

/*****************************************
*                                        *
*                 storage                *
*                                        *
*****************************************/
type storage struct {
	path    string
	content map[string]interface{}
	lock    sync.RWMutex
}

func initStorage(path string) *storage {
	content, err := ioutil.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			content = createStorage(path)
		} else {
			panic(err)
		}
	}

	var data map[string]interface{}
	if err := json.Unmarshal(content, &data); err != nil {
		panic(err)
	}

	return &storage{
		path:    path,
		content: data,
	}
}

func createStorage(path string) []byte {
	f, err := os.Create(path)
	if err != nil {
		panic(f)
	}
	defer f.Close()

	content := []byte("{}")
	f.Write(content)
	return content
}

func readStorage(path string) []byte {
	f, err := os.Open(path)
	if err != nil {
		panic(err)
	}
	defer f.Close()

	content, err := ioutil.ReadAll(f)
	if err != nil {
		panic(err)
	}

	return content
}

func (s *storage) writeStorage() {
	s.lock.Lock()
	defer s.lock.Unlock()

	res, err := json.Marshal(s.content)
	if err != nil {
		panic(err)
	}

	if err = ioutil.WriteFile(s.path, res, 0644); err != nil {
		panic(err)
	}
}

func (s *storage) getAlias(key string) interface{} {
	return s.content[key]
}

func (s *storage) addAlias(key, value string) {
	s.content[key] = value
	s.writeStorage()
}

func (s *storage) delAlias(key string) {
	delete(s.content, key)
	s.writeStorage()
}

func (s *storage) renameAlias(key, newName string) {
	if _, ok := s.content[key]; ok {
		s.content[newName] = s.content[key]
		delete(s.content, key)
		s.writeStorage()
	}
}

func (s *storage) clearAlias() {
	s.content = map[string]interface{}{}
	s.writeStorage()
}

func (s *storage) aliasList() interface{} {
	var data sync.Map
	data.Range(func(key, value interface{}) bool {
		s.content[fmt.Sprint(key)] = value
		return true
	})

	b, err := json.MarshalIndent(s.content, "", "  ")
	if err != nil {
		panic(err)
	}

	return string(b)
}
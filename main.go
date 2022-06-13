package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"io/ioutil"
	"os"
)

type Arguments map[string]string

type Item struct {
	Id    string `json:"id"`
	Email string `json:"email"`
	Age   int    `json:"age"`
}

func performAdd(filename string, itemBlob string, writer io.Writer) error {
	if len(itemBlob) == 0 {
		return fmt.Errorf("-item flag has to be specified")
	}

	items, err := readItems(filename)
	if err != nil {
		return err
	}

	var item Item
	err = json.Unmarshal([]byte(itemBlob), &item)
	if err != nil {
		return err
	}

	//check existing
	for _, v := range items {
		if v.Id == item.Id {
			writer.Write([]byte(fmt.Sprintf("Item with id %s already exists", item.Id)))
			return nil
		}
	}

	//add
	items = append(items, item)
	return writeItems(filename, items)
}

func performList(filename string, writer io.Writer) error {
	body, err := readFile(filename)
	if err != nil {
		return err
	}
	writer.Write(body)
	return nil
}

func performFindById(filename string, id string, writer io.Writer) error {
	if len(id) == 0 {
		return fmt.Errorf("-id flag has to be specified")
	}

	items, err := readItems(filename)
	if err != nil {
		return err
	}

	//find
	for _, v := range items {
		if v.Id == id {
			blob, err := json.Marshal(v)
			if err != nil {
				return err
			}
			_, err = writer.Write(blob)
			return err
		}
	}

	//ok, not found
	return nil
}

func performRemove(filename string, id string, writer io.Writer) error {
	if len(id) == 0 {
		return fmt.Errorf("-id flag has to be specified")
	}

	items, err := readItems(filename)
	if err != nil {
		return err
	}

	//find
	for i, v := range items {
		if v.Id == id {
			items = append(items[:i], items[i+1:]...)
			return writeItems(filename, items)
		}
	}

	//ok, not found
	writer.Write([]byte(fmt.Sprintf("Item with id %s not found", id)))
	return nil
}

func Perform(args Arguments, writer io.Writer) error {
	oper, ok := args["operation"]
	if !ok || len(oper) == 0 {
		return fmt.Errorf("-operation flag has to be specified")
	}

	filename, ok := args["fileName"]
	if !ok || len(filename) == 0 {
		return fmt.Errorf("-fileName flag has to be specified")
	}

	switch oper {
	case "add":
		return performAdd(filename, args["item"], writer)
	case "list":
		return performList(filename, writer)
	case "findById":
		return performFindById(filename, args["id"], writer)
	case "remove":
		return performRemove(filename, args["id"], writer)
	}

	//lint:ignore ST1005 coz of unit tests
	return fmt.Errorf("Operation %s not allowed!", oper)
}

func main() {
	err := Perform(parseArgs(), os.Stdout)
	if err != nil {
		panic(err)
	}
}

//helpers

func parseArgs() Arguments {
	names := []string{"id", "item", "operation", "fileName"}
	values := make([]string, len(names))

	for i := range names {
		flag.StringVar(&values[i], names[i], "", "flag "+names[i])
	}

	flag.Parse()

	var args Arguments = make(map[string]string)
	for i := range names {
		if len(values[i]) > 0 {
			args[names[i]] = values[i]
		}
	}

	return args
}

//just like ioutil.ReadFile
func readFile(filename string) ([]byte, error) {
	f, err := os.OpenFile(filename, os.O_RDONLY|os.O_CREATE, 0644)
	if err != nil {
		return nil, err
	}
	defer f.Close()

	body, err := ioutil.ReadAll(f)
	if err != nil {
		return nil, err
	}
	return body, nil
}

func writeFile(filename string, body []byte) error {
	f, err := os.OpenFile(filename, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0644)
	if err != nil {
		return err
	}
	defer f.Close()

	_, err = f.Write(body)
	return err
}

func readItems(filename string) ([]Item, error) {
	blob, err := readFile(filename)
	if err != nil {
		return nil, err
	}

	if len(blob) == 0 {
		return make([]Item, 0), nil
	}

	var items []Item
	err = json.Unmarshal(blob, &items)
	if err != nil {
		return nil, err
	}

	return items, nil
}

func writeItems(filename string, items []Item) error {
	blob, err := json.Marshal(items)
	if err != nil {
		return err
	}

	return writeFile(filename, blob)
}

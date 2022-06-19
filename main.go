package main

import (
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"io"
	"io/ioutil"
	"os"
)

type Arguments map[string]string
type User struct {
	Id    string `json:"id"`
	Email string `json:"email"`
	Age   int    `json:"age"`
}

func parseArgs() Arguments {
	operation := flag.String("operation", "", "")
	fileName := flag.String("fileName", "", "")
	item := flag.String("item", "", "")
	id := flag.String("id", "", "")
	flag.Parse()

	args := Arguments{}
	args["operation"] = *operation
	args["fileName"] = *fileName
	args["item"] = *item
	args["id"] = *id
	return args
}

func Perform(args Arguments, writer io.Writer) error {
	if args["operation"] == "" {
		return errors.New("-operation flag has to be specified")
	}
	if args["operation"] != "list" && args["operation"] != "add" && args["operation"] != "remove" && args["operation"] != "findById" {
		return errors.New(fmt.Sprintf("Operation %v not allowed!", args["operation"]))
	}
	if args["fileName"] == "" {
		return errors.New("-fileName flag has to be specified")
	}
	if args["operation"] == "add" && args["item"] == "" {
		return errors.New("-item flag has to be specified")
	}
	if (args["operation"] == "remove" || args["operation"] == "findById") && args["id"] == "" {
		return errors.New("-id flag has to be specified")
	}

	f, err := os.OpenFile(args["fileName"], os.O_RDWR|os.O_CREATE, 0644)
	defer f.Close()
	if err != nil {
		return err
	}
	bytes, err := ioutil.ReadAll(f)
	if err != nil {
		return err
	}

	var users []User
	if len(bytes) != 0 {
		err = json.Unmarshal(bytes, &users)
		if err != nil {
			return err
		}
	}

	switch args["operation"] {
	case "list":
		if len(users) != 0 {
			bytes, _ = json.Marshal(users)
			fmt.Fprintf(writer, "%v", string(bytes))
		}
	case "add":
		var item User
		err = json.Unmarshal([]byte(args["item"]), &item)
		if err != nil {
			return err
		}
		for _, v := range users {
			if item.Id == v.Id {
				fmt.Fprintf(writer, "Item with id %v already exists", item.Id)
				return nil
			}
		}
		users = append(users, item)
		rewriteUsersToFile(users, f)
		if err != nil {
			return err
		}
	case "remove":
		for i, v := range users {
			if args["id"] == v.Id {
				users[i] = users[len(users)-1] // Copy last element to index i.
				users[len(users)-1] = User{}   // Erase last element (write zero value).
				users = users[:len(users)-1]
				rewriteUsersToFile(users, f)
				return nil
			}
		}
		fmt.Fprintf(writer, "Item with id %v not found", args["id"])
	case "findById":
		for _, v := range users {
			if args["id"] == v.Id {
				bytes, _ = json.Marshal(v)
				fmt.Fprintf(writer, "%v", string(bytes))
				return nil
			}
		}
	}

	return nil
}

func rewriteUsersToFile(users []User, f *os.File) error {
	bytes, err := json.Marshal(users)
	if err != nil {
		return err
	}
	f.Truncate(0)
	f.Seek(0, 0)
	_, err = f.Write(bytes)
	if err != nil {
		return err
	}

	return nil
}

func main() {
	err := Perform(parseArgs(), os.Stdout)
	if err != nil {
		panic(err)
	}
}

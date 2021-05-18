package main

import (
	"fmt"
	"os"
	"strconv"
	"strings"

	"github.com/jessevdk/go-flags"
)

type Options struct {
	DbPath    string   `long:"db-path" description:"Database path" default:".db.tt"`
	New       struct{} `command:"new" alias:"todo" alias:"n" alias:"make" description:"Add new task" required:"false"`
	List      struct{} `command:"list" alias:"l" alias:"ls" description:"List tasks" required:"false"`
	Board     string   `short:"b" long:"board" description:"Specify board name" required:"false"`
	Task      string   `short:"t" long:"task" description:"Specify task name" required:"false"`
	CheckTask string   `short:"c" long:"check" description:"Check task" required:"false"`
}

var options Options

var parser = flags.NewParser(&options, flags.Default)

func main() {
	parser.Command.SubcommandsOptional = true
	args, err := parser.Parse()
	if err != nil {
		if err.(*flags.Error).Type == flags.ErrHelp {
			os.Exit(0)
		}
		parser.WriteHelp(os.Stderr)
		os.Exit(1)
	}

	db, err := ReadDatabaseFromFile(options.DbPath)
	if err != nil {
		if os.IsNotExist(err) {
			db = NewDatabase()
		} else {
			panic(err)
		}
	}
	defer db.WriteToFile(options.DbPath)

	var command string
	switch {
	case parser.Active != nil:
		command = parser.Active.Name
	case len(args) == 0:
		command = "list"
	default:
		command = "new"
	}

	switch command {
	case "new":
		db.NewTask(strings.Join(args, " "))
	case "list":
		db.printDB(strings.Join(args, " "))
	}

	if options.CheckTask != "" {
		if options.Board == "" {
			fmt.Println("You must provide a name of board.")
			os.Exit(1)
		} else {
			id, err := strconv.Atoi(options.CheckTask)
			if err != nil {
				fmt.Println(err)
			}
			err = db.checkTask(int64(id), options.Board)
			if err != nil {
				fmt.Println(err)
			}
			fmt.Printf("task with id %d from board %s checked as done\n", int64(id), options.Board)
		}
	}
}

// vi:noet:ts=4:sw=4:

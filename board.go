package main

import (
	"encoding/json"
	"fmt"
	"os"
	"strings"
	"unicode"

	"github.com/fatih/color"
	"github.com/lithammer/fuzzysearch/fuzzy"
)

type database struct {
	Boards []*board `json:"boards"`
}

type board struct {
	ID     int64   `json:"id"`
	Name   string  `json:"name"`
	Status bool    `json:"status"`
	Tasks  []*task `json:"tasks"`
}

type task struct {
	ID     int64  `json:"id"`
	Text   string `json:"name"`
	Status bool   `json:"status"`
}

var green = color.New(color.FgGreen)
var purple = color.New(color.FgMagenta)
var gray = color.New(color.FgHiBlack)
var underline = color.New(color.Underline)

var indent = 10

func NewDatabase() *database {
	return &database{[]*board{}}
}

func ReadDatabaseFromFile(name string) (*database, error) {
	file, err := os.Open(name)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	var db database
	err = json.NewDecoder(file).Decode(&db)
	return &db, err
}

func (db *database) WriteToFile(name string) error {
	file, err := os.OpenFile(name, os.O_WRONLY|os.O_CREATE, 0600)
	if err != nil {
		return err
	}
	defer file.Close()

	return json.NewEncoder(file).Encode(db)
}

func (db *database) addBoard(b *board) error {
	var err error
	var maxID int64 = 0

	if len(b.Name) >= 33 {
		return fmt.Errorf("board name must not be longer than 32 characters")
	}
	for _, v := range b.Name {
		if unicode.IsSpace(v) {
			return fmt.Errorf("board name must not contain spaces")
		} else if unicode.IsPunct(v) {
			return fmt.Errorf("board name must not contain special characters")
		}
	}

	for _, board := range db.Boards {
		if board.Name == b.Name {
			return fmt.Errorf("board with this name already exist")
		}
		if board.ID > maxID {
			maxID = board.ID
		}
	}

	b.ID = maxID + 1
	db.Boards = append(db.Boards, b)

	return err
}

func (db *database) NewTask(text string) {
	db.addTask(&task{Text: text}, "actual")
}

func (db *database) addTask(t *task, bName string) error {
	var err error
	var maxID int64 = 0

	// TODO: maybe it should be default(system) Board with default(system) name
	if len(db.Boards) == 0 {
		err = db.addBoard(&board{Name: bName, Status: false})
	}

	for _, board := range db.Boards {
		if strings.ToLower(board.Name) == strings.ToLower(bName) {
			for _, task := range board.Tasks {
				if task.ID > maxID {
					maxID = task.ID
				}
			}
			t.ID = maxID + 1
			board.Tasks = append(board.Tasks, t)
			return err
		}
	}

	// TODO: what if there's no Boards with given name: print error or add first one
	/*for _, task := range db.Boards[0].Tasks {
		if task.ID > maxID {
			maxID = task.ID
		}
	}
	t.ID = maxID + 1
	db.Boards[0].Tasks = append(db.Boards[0].Tasks, t)
	err = db.writeToDB()
	return err*/

	return err
}

func (db *database) checkTask(taskId int64, bName string) error {
	var err error
	bName = strings.ToLower(bName)

	for _, board := range db.Boards {
		if strings.ToLower(board.Name) == strings.ToLower(bName) {
			for _, task := range board.Tasks {
				if task.ID == taskId {
					task.Status = true
					return err
				}
			}
		}
	}

	return err
}

type stat struct {
	done       int64
	inProgress int64
	percent    int64
}

func (db *database) stat() stat {
	result := stat{done: 0, inProgress: 0, percent: 0}

	for _, board := range db.Boards {
		if len(board.Tasks) != 0 {
			for _, task := range board.Tasks {
				if task.Status {
					result.done++
				} else {
					result.inProgress++
				}
			}
		}
	}
	if result.done > 0 || result.inProgress > 0 {
		result.percent = result.done * 100 / (result.done + result.inProgress)
	}
	return result
}

func (db *database) printDB(pattern string) {
	for _, board := range db.Boards {
		fmt.Printf("%[2]*[1]s@", "", indent/2)
		underline.Println(board.Name)
		for _, task := range board.Tasks {
			if fuzzy.Match(pattern, task.Text) {
				gray.Printf("%[1]*[2]d. ", indent, task.ID)
				if task.Status {
					green.Print("[✓] ")
					gray.Print(task.Text)
				} else {
					purple.Print("[ ] ")
					fmt.Print(task.Text)
				}
				fmt.Println()
			}
		}
		fmt.Println()
	}
	stat := db.stat()
	if stat.done == 0 && stat.inProgress == 0 {
		fmt.Println("No tasks were found.")
	} else {
		fmt.Printf("%[1]*[2]s%[3]d%% of all tasks complete\n", indent/2, "", stat.percent)
		fmt.Printf("%[1]*[2]s", indent/2, "")
		green.Print(stat.done)
		fmt.Print(" done | ")
		purple.Print(stat.inProgress)
		fmt.Println(" in progress")
	}
}

// vi:noet:ts=4:sw=4:

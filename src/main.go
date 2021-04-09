package main

import (
	"bufio"
	"encoding/csv"
	"fmt"
	"io"
	"log"
	"os"
	"os/exec"
	"sort"
	"strconv"
	"strings"
)

type user struct {
	username    string
	sessionname string
	sessionid   string
	memusage    float64
}

type userSorted []user

func (a userSorted) Len() int {
	return len(a)
}

func (a userSorted) Less(i, j int) bool {
	return a[i].memusage < a[j].memusage
}

func (a userSorted) Swap(i, j int) {
	a[i], a[j] = a[j], a[i]
}

func newUser(name string, session string, id string) *user {
	u := user{username: name, sessionname: session, sessionid: id}
	return &u
}

func (u *user) setMem(value float64) {
	u.memusage = value
}

func main() {
	fmt.Println("--------------------\n     Memory Usage\n--------------------")
	fmt.Println("--------------------\nLeave empty to get your self memory usage")
	fmt.Println("or type \"all\" to retrieve all user's memory usage\n--------------------")

	fmt.Print(">")
	var args string
	fmt.Scanln(&args)

	args = strings.ToLower(args)

	switch args {
	case "all":
		allTasks()
	default:
		actualTask()
	}

	removeUser()

	removeCSV()

	reader := bufio.NewReader(os.Stdin)
	fmt.Println("\n Press ENTER to exit")
	a, _ := reader.ReadString('\n')
	_ = a
}

//return all tasks listed by user and session
func allTasks() {
	users := genUsers(0)
	printUsers(users)
}

//return only the memery ammoutn of the actual session
func actualTask() {
	user := genUsers(1)
	printUsers(user)
}

func genUsers(q int) []user {
	//all users
	com := ""
	if q == 0 {
		com = "quser > users.txt"
	} else {
		com = "quser %username% > users.txt"
	}
	cmd := exec.Command("cmd", "/C", com)

	err := cmd.Run()

	if err != nil {
		log.Fatalln("Error importing user info:\n", err)
	}

	file, err := os.Open("users.txt")

	if err != nil {
		log.Fatalln("Error opening user files: \n", err)
	}

	defer file.Close()

	scanner := bufio.NewScanner(file)

	mUser := make([]user, 0)
	for scanner.Scan() {
		line := scanner.Text()
		line = strings.ToLower(line)
		if strings.Contains(line, "rdp-tcp") || strings.Contains(line, "disco") {
			u := user{
				username:    strings.ReplaceAll(strings.TrimSpace(line[0:23]), ">", ""), //0-23
				sessionname: strings.TrimSpace(line[24:36]),                             //24-36
				sessionid:   strings.TrimSpace(line[39:45]),                             //39-45
			}
			mUser = append(mUser, u)
		}
	}

	return mUser
}

func printUsers(users []user) {
	getAllTasks()
	var total float64
	//get the mem usage
	for i := range users {
		users[i].memusage = calcMem(users[i].sessionid)
		total += users[i].memusage
	}

	sort.Sort(userSorted(users))

	for _, user := range users {
		fmt.Printf("\n----------\n>User: %s | Session: %s | ID: %s\n", user.username, user.sessionname, user.sessionid)
		fmt.Printf(">Memory: %.3f\n----------", user.memusage)
	}

	if len(users) > 1 {
		fmt.Printf("\n----------\n>Mem. average: %.3f GB\n", (total / float64(len(users))))
		fmt.Printf(">Mem. total amount usage: %.3f GB\n----------", (total / 1024))
	}
}

//get tasks from all users
func getAllTasks() {
	cmd := exec.Command("cmd", "/C", "tasklist /FO CSV > tasks.csv")
	err := cmd.Run()

	if err != nil {
		log.Fatalln("Error importing tasks:\n", err)
	}
}

func calcMem(iduser string) float64 {
	file, err := os.Open("tasks.csv")

	if err != nil {
		log.Fatalln("Error on read tasks file: \n", err)
	}

	defer file.Close()

	var total float64

	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		r := csv.NewReader(strings.NewReader(scanner.Text()))
		r.Comma = ','

		for {
			record, err := r.Read()
			if err == io.EOF {
				break
			}
			if err != nil {
				log.Fatalln("Error reading line:\n", err)
			}
			if iduser == record[3] {
				s := strings.ReplaceAll(record[4], " ", "")
				s = strings.ReplaceAll(s, "K", "")

				if a, err := strconv.ParseFloat(s, 64); err == nil {
					total += a
				}

				if err != nil {
					log.Fatalln("Error calculating memory amount:\n", err)
				}
			}

		}
	}
	if err := scanner.Err(); err != nil {
		fmt.Println(err)
	}

	return total
}

func removeCSV() {
	cmd := exec.Command("cmd", "/C", "del tasks.csv")
	err := cmd.Run()

	if err != nil {
		log.Fatal("Error removing tasks file: \n", err)
	}
}

func removeUser() {
	cmd := exec.Command("cmd", "/C", "del users.txt")
	err := cmd.Run()

	if err != nil {
		log.Fatal("Error removingo user file: \n", err)
	}
}

package main

import (
	"bufio"
	"encoding/csv"
	"fmt"
	"io"
	"log"
	"os"
	"os/exec"
	"strconv"
	"strings"
)

type user struct {
	username    string
	sessionname string
	sessionid   string
	memusage    float64
}

func newUser(name string, session string, id string) *user {
	u := user{username: name, sessionname: session, sessionid: id}
	return &u
}

func main() {
	fmt.Println("--------------------\n     Memory Usage\n--------------------")
	fmt.Println("--------------------\nPresione ENTER para obter o seu uso atual de RAM, ou,")
	fmt.Println("digite \"all\" para obter o todos os usos da sessão.\n--------------------")

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
	fmt.Println("---------\nPressione ENTER para sair.")
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

func genUsers(q int) map[int]user {
	//all users
	com := ""
	if q == 0 {
		com = "quser > users.txt"
	} else {
		com = "quser %username% > users.txt"
	}
	cmd := exec.Command("cmd", "/C", com)

	err := cmd.Run()

	check(err)

	file, err := os.Open("users.txt")

	check(err)

	defer file.Close()

	scanner := bufio.NewScanner(file)
	i := 0
	mUser := make(map[int]user)
	for scanner.Scan() {
		line := scanner.Text()
		line = strings.ToLower(line)
		if strings.Contains(line, "rdp-tcp") || strings.Contains(line, "disco") {
			u := user{
				username:    strings.ReplaceAll(strings.TrimSpace(line[0:23]), ">", ""), //0-23
				sessionname: strings.TrimSpace(line[24:36]),                             //24-36
				sessionid:   strings.TrimSpace(line[39:45]),                             //39-45
			}
			mUser[i] = u
			i++
		}
	}

	return mUser
}

func printUsers(users map[int]user) {
	getAllTasks()
	var total float64

	//get the mem usage
	for _, user := range users {
		user.memusage = calcMem(&user.sessionid)
		fmt.Printf("----------\nUsuário: %s     Sessão autal: %s     ID: %s\n----------", user.username, user.sessionname, user.sessionid)
		fmt.Printf("----------\nUso atual de memória: %.3f\n----------", user.memusage)
		total += user.memusage
	}

	if len(users) > 1 {
		fmt.Printf("----------\n>Média de uso: %.3f GB\n----------", (total / float64(len(users))))
		fmt.Printf("----------\n>TOTAL GERAL: %.3f GB\n----------", (total / 1024))
	}
}

//get tasks from all users
func getAllTasks() {
	cmd := exec.Command("cmd", "/C", "tasklist /FO CSV > tasks.csv")
	err := cmd.Run()

	if err != nil {
		log.Fatal(err)
	}
}

func calcMem(iduser *string) float64 {
	file, err := os.Open("tasks.csv")

	if err != nil {
		fmt.Println(err)
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
				log.Fatal(err)
			}
			if *iduser == record[3] {
				s := strings.ReplaceAll(record[4], " ", "")
				s = strings.ReplaceAll(s, "K", "")

				if a, err := strconv.ParseFloat(s, 64); err == nil {
					total += a
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
		log.Fatal("Erro ao remover arquivo de Tasks.\n", err)
	}
}

func removeUser() {
	cmd := exec.Command("cmd", "/C", "del users.txt")
	err := cmd.Run()

	if err != nil {
		log.Fatal("Erro ao remover arquivo de Usuários.\n", err)
	}
}

func check(e error) {
	if e != nil {
		panic(e)
	}
}
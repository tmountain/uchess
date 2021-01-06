package main

import (
	"bufio"
	"fmt"
	"io"
	"log"
	"os"
	"os/exec"
)

func logStdout(stdout io.ReadCloser) {
	// read input from cmd stdout pipe
	reader := bufio.NewReader(stdout)
	scanner := bufio.NewScanner(reader)

	for scanner.Scan() {
		rsp := scanner.Text()
		fmt.Println(rsp)
		log.Printf("> %v\n", rsp)
	}
}

func main() {
	if len(os.Args) != 2 {
		os.Stderr.WriteString("Usage: proxy <uci path>\n")
		os.Exit(0)
	}

	uciCmd := os.Args[1]
	f, _ := os.OpenFile("uchess.log", os.O_RDWR|os.O_CREATE|os.O_TRUNC, 0666)
	log.SetOutput(f)
	defer f.Close()
	cmd := exec.Command(uciCmd)
	cstdin, _ := cmd.StdinPipe()
	cstdout, _ := cmd.StdoutPipe()
	scanner := bufio.NewScanner(os.Stdin)

	if err := cmd.Start(); err != nil {
		panic(err)
	}

	go logStdout(cstdout)

	for scanner.Scan() {
		input := scanner.Text()
		cstdin.Write([]byte(fmt.Sprintln(input)))
		log.Printf("< %v\n", input)
		f.Sync()
	}

	cmd.Wait()
}

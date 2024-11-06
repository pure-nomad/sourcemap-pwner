package main

import (
	"bufio"
	"fmt"
	"log"
	"os"
	"sourcemap-pwner/internals"
	"sourcemap-pwner/internals/utils"
	"sync"
)

var wg sync.WaitGroup

func main() {
	fmt.Println("Welcome to sourcemap-pwner !")

	if len(os.Args) <= 1 {
		fmt.Println("Usage: ./sourcemap-pwner urls.txt")
	}

	file, err := os.OpenFile(os.Args[1], os.O_RDONLY, 0644)
	if err != nil {
		panic(err)
	}

	defer func(file *os.File) {
		err := file.Close()
		if err != nil {

		}
	}(file)

	channel := make(chan utils.SourceMap)

	s := bufio.NewScanner(file)

	for s.Scan() {
		wg.Add(1)
		go internals.CheckUrl(s.Text(), channel, &wg)
	}

	go func() {
		defer close(channel)
		wg.Wait()
	}()

	for rec := range channel {
		log.Printf("%q", rec)
	}

}

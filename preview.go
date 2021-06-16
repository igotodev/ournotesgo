package main

import (
	"bufio"
	"fmt"
	"log"
	"os"
	"time"

	"github.com/muesli/termenv"
)

// previewRainbow printed UTF8 text from file to os.Stdout (not necessarily, it's for fun)
func previewRainbow(file string) {
	logo, err := os.Open(file)
	if err != nil {
		log.Fatal(err)
	}
	p := termenv.ColorProfile()

	time.Sleep(100 * time.Millisecond)
	scanner := bufio.NewScanner(logo)
	for scanner.Scan() {
		myBytes := scanner.Text() + "\n"

		color := 1

		for _, v := range myBytes {
			time.Sleep(25 * time.Millisecond)

			color++

			switch color {
			case 1:
				{
					fmt.Fprint(os.Stdout, termenv.String(string(v)).Foreground(p.Color("#71BEF2")))
				}
			case 2:
				{
					fmt.Fprint(os.Stdout, termenv.String(string(v)).Foreground(p.Color("#E88388")))
				}
			case 3:
				{
					fmt.Fprint(os.Stdout, termenv.String(string(v)).Foreground(p.Color("#A8CC8C")))
				}
			case 4:
				{
					fmt.Fprint(os.Stdout, termenv.String(string(v)).Foreground(p.Color("#DBAB79")))
				}
			case 5:
				{
					fmt.Fprint(os.Stdout, termenv.String(string(v)).Foreground(p.Color("#D290E4")))
				}
			case 6:
				{
					fmt.Fprint(os.Stdout, termenv.String(string(v)).Foreground(p.Color("#66C2CD")))
				}
			case 7:
				{
					fmt.Fprint(os.Stdout, termenv.String(string(v)).Foreground(p.Color("#B9BFCA")))
				}
			}
			if color == 7 {
				color = 1
			}
		}
	}
	time.Sleep(100 * time.Millisecond)
	fmt.Fprintf(os.Stdout, "\n")
}

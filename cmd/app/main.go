package main

import (
	"adminadmin/internal/application"
	"fyne.io/fyne/v2/app"
	"log"
	"os"
)

func main() {
	// Configure logging
	log.SetOutput(os.Stdout)
	log.SetFlags(log.Ldate | log.Ltime | log.Lmicroseconds | log.Lshortfile)

	log.Println("=====================================")
	log.Println("        admin:admin Starting")
	log.Println("=====================================")

	fyneApp := app.New()
	log.Println("MAIN: Fyne application created")

	application.NewApp(fyneApp).Run()

	log.Println("=====================================")
	log.Println("  admin:admin Exited")
	log.Println("=====================================")
}

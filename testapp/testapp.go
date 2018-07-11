package main

import (
	"fmt"

	"github.com/Guardtime/evat/pungi"
)

func main() {
	pungi.New("testapp [your-name]", "Starts music store web application.").
		Key("port", 8080, "Listen port").
		Key("cpuprofile", false, "Starts CPU profiler if set to true.").
		Key("dbUri", "boltdb:db/my.db", "DB Uri").
		Run(startWebApp).Execute()
}

func startWebApp(conf *pungi.Conf, args []string) error {
	port := fmt.Sprintf(`"%s": %d`, "port", conf.GetInt("port"))
	cpuprofile := fmt.Sprintf(`"%s": %v`, "cpuprofile", conf.GetBool("cpuprofile"))
	dbUri := fmt.Sprintf(`"%s": "%s"`, "dbUri", conf.GetString("dbUri"))
	fmt.Printf("{%s, \n%s, \n%s\n}", port, cpuprofile, dbUri)
	return nil
}

package main

import (
	"fmt"

	"github.com/Guardtime/evat/pungi"
)

func main() {
	pungi.New("testapp", "Starts music store web application.").
		Key("cpuprofile", false, "Starts CPU profiler if set to true.").
		Cmd(pungi.Cmd("grpc", "Starts gRPC service.", startGrpcService).
			Key("port", 8080, "Service listen port.").
			Key("dbUri", "boltdb:db/my.db", "Db Uri"),
		).
		Cmd(pungi.Cmd("httpgw", "Starts Http GW.", startHttpGW).
			Key("port", 8080, "Http GW listen port.").
			Key("grpcUri", "http://localhost:5432", "Grpc service Uri."),
		).
		Execute()
}

func startHttpGW(conf *pungi.Conf, args []string) error {
	port := fmt.Sprintf(`"port": %d`, conf.GetInt("port"))
	cpuprofile := fmt.Sprintf(`"cpuprofile": %v`, conf.GetBool("cpuprofile"))
	grpcUri := fmt.Sprintf(`"grpcUri": "%s"`, conf.GetString("grpcUri"))
	fmt.Printf("{%s, \n%s, \n%s\n}", port, cpuprofile, grpcUri)
	return nil
}

func startGrpcService(conf *pungi.Conf, args []string) error {
	port := fmt.Sprintf(`"port": %d`, conf.GetInt("port"))
	cpuprofile := fmt.Sprintf(`"cpuprofile": %v`, conf.GetBool("cpuprofile"))
	dbUri := fmt.Sprintf(`"dbUri": "%s"`, conf.GetString("dbUri"))
	fmt.Printf("{%s, \n%s, \n%s\n}", port, cpuprofile, dbUri)
	return nil
}

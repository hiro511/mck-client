package main

import (
	"fmt"
	"log"

	"golang.org/x/net/context"
	"google.golang.org/grpc"

	pb "github.com/hiro511/mck"
)

const (
	address     = "localhost:8080"
	defaultName = "world"
)

func main() {
	conn, err := grpc.Dial(address, grpc.WithInsecure())
	if err != nil {
		log.Fatalf("did not connect: %v", err)
	}
	defer conn.Close()
	c := pb.NewMckClient(conn)

	r, err := c.FetchJobs(context.Background(), &pb.JobRequest{NumRequest: 10})
	if err != nil {
		log.Fatalf("error")
	}
	fmt.Println(r.Name)
}

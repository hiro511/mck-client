package main

import (
	"bufio"
	"bytes"
	"flag"
	"fmt"
	"log"
	"os"
	"os/exec"
	"sync"
	"time"

	"golang.org/x/net/context"
	"google.golang.org/grpc"

	pb "github.com/hiro511/mck"
)

const (
	address     = "localhost:8080"
	defaultName = "world"
)

var (
	numRepeat  int
	numProcess int
	wg         *sync.WaitGroup
)

func init() {
	flag.IntVar(&numRepeat, "n", 1, "number of jobs")
	flag.IntVar(&numProcess, "p", 1, "number of process")
}

func main() {
	flag.Parse()
	wg = &sync.WaitGroup{}
	for count := 0; count < numProcess; count++ {
		wg.Add(1)
		go execute()
		time.Sleep(10 * time.Second)
	}
	wg.Wait()

}

func execute() {
	defer wg.Done()
	for true {
		routine()
	}
}

func routine() {
	conn, err := grpc.Dial(address, grpc.WithInsecure())
	if err != nil {
		log.Printf("did not connect: %v", err)
		time.Sleep(3 * time.Minute)
		return
	}
	defer conn.Close()

	c := pb.NewMckClient(conn)
	jobs, err := c.FetchJobs(context.Background(), &pb.JobRequest{NumRequest: int32(numRepeat)})
	if err != nil {
		log.Printf("did not fetch jobs: %v", err)
		time.Sleep(3 * time.Minute)
		return
	}
	fmt.Println(jobs.Name)

	_, err = os.Stat(jobs.Name)
	if err != nil {
		var fp *os.File
		fp, err = os.Create(jobs.Name)
		if err != nil {
			log.Printf("did not create: %v", err)
			time.Sleep(3 * time.Minute)
			return
		}

		writer := bufio.NewWriter(fp)
		_, err = writer.Write(jobs.Settings)
		if err != nil {
			log.Printf("did not write: %v", err)
			time.Sleep(3 * time.Minute)
			return
		}
		writer.Flush()
		fp.Close()
	}

	_, err = os.Stat(jobs.MckName)
	if err != nil {
		var mck *pb.MolComKit
		mck, err = c.DownloadMCK(context.Background(), &pb.MCKRequest{Name: jobs.MckName})
		if err != nil {
			log.Printf("did not download: %v", err)
			time.Sleep(3 * time.Minute)
			return
		}
		var fp *os.File
		fp, err = os.OpenFile(jobs.MckName, os.O_WRONLY|os.O_CREATE, 0777)
		if err != nil {
			log.Printf("did not open file: %v", err)
			time.Sleep(3 * time.Minute)
			return
		}

		writer := bufio.NewWriter(fp)
		_, err = writer.Write(mck.Binary)
		if err != nil {
			log.Printf("did not write: %v", err)
			time.Sleep(3 * time.Minute)
			return
		}
		writer.Flush()
		fp.Close()
	}

	// var out []byte
	log.Printf("-jar %s -pfile: %s -n %d", jobs.MckName, jobs.Name, jobs.NumJobs)
	// cmd := exec.Command("java", fmt.Sprintf("-jar %s -pfile: %s -n %d",
	// 	jobs.MckName, jobs.Name, jobs.NumJobs))
	cmd := exec.Command("java", "-version")
	var out bytes.Buffer
	var stderr bytes.Buffer
	cmd.Stdout = &out
	cmd.Stderr = &stderr
	err = cmd.Run()
	if err != nil {
		fmt.Println(fmt.Sprint(err) + ": " + stderr.String())
		return
	}
	fmt.Println("Result: " + out.String())
	if err != nil {
		log.Printf("did not execute command: %v", err)
		return
	}
	// _, err = c.SendResult(context.Background(), &pb.JobResult{Name: jobs.Name, Result: string(out)})
	// if err != nil {
	// 	log.Printf("did not send result: %v", err)
	// 	return
	// }

}

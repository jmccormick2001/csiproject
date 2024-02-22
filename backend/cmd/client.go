package main

import (
	"context"
	"fmt"
	"log"
	"time"

	"example.com/csiproject/backend/client"
	"example.com/csiproject/backend/model"
)

func main() {

	client := client.Client{
		Hostname: "192.168.0.108",
		Port:     "10000",
	}

	ctx := context.Background()
	timeout := 30 * time.Second
	reqContext, _ := context.WithTimeout(ctx, timeout)
	someVolume, err := client.GetVolume(reqContext, "1")
	if err != nil {
		log.Fatal(err)
	}
	fmt.Printf("volume response %+v\n", someVolume)

	volume := model.Volume{ID: "13", Name: "leader", Hostport: "192.168.0.112", Size: "1G"}
	newVolume, err := client.CreateVolume(reqContext, volume)
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println(newVolume)

	mm, err := client.GetAllVolumes(reqContext)
	if err != nil {
		log.Fatal(err)
	}
	for i := 0; i < len(mm.Volumes); i++ {
		fmt.Printf("get all volumes response %d %+v\n", i, mm.Volumes[i])
	}
	err = client.DeleteVolume(reqContext, "13")
	if err != nil {
		log.Fatal(err)
	}
	fmt.Printf("delete volume worked\n")
}

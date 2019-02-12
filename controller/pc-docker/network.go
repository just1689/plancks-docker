package pc_docker

import (
	"context"
	"fmt"
	"github.com/docker/docker/api/types"
	"github.com/docker/docker/client"
	"log"
)

const DefaultNetwork = "plancks-net"


func createNetwork(serviceName string) (err error) {

	exists, err := checkNetworkExists(serviceName)
	if err != nil {
		log.Fatalln("Could not check if the network exists")
	}

	if !exists {
		success, err := createOverlayNetwork(serviceName)

		if err != nil {
			log.Fatalln("Could not check if the network exists")
		}

		if !success {
			log.Fatalln("Create network was not successful")
		}
	}

	return err
}


//CreateOverlayNetwork creates an overlay network in docker swarm
func createOverlayNetwork(name string) (success bool, err error) {
	cli, err := client.NewEnvClient()
	ctx := context.Background()
	if err != nil {
		log.Panicln(fmt.Sprintf("Error getting docker client environment: %s", err))
		return false, err
	}

	res, err := cli.NetworkCreate(ctx, name, types.NetworkCreate{Driver: "overlay", Attachable: true})

	log.Printf(res.ID)
	log.Printf(res.Warning)

	if err != nil {
		log.Printf(err.Error())
		return false, err
	}
	success = true
	return

}

//CheckNetworkExists tells us if a network name exists
func checkNetworkExists(name string) (exists bool, err error) {
	exists, _, err = describeNetwork(name)
	return

}

//DeleteNetwork removes a network by name
func deleteNetwork(name string) (success bool, err error) {
	cli, err := client.NewEnvClient()
	ctx := context.Background()
	if err != nil {
		log.Println(fmt.Sprintf("Error getting docker client environment: %s", err))
		return false, err
	}

	exists, ID, err := describeNetwork(name)
	if err != nil {
		return false, err
	}
	if !exists {
		success = false
		return
	}
	err = cli.NetworkRemove(ctx, ID)

	if err != nil {
		return false, err
	}
	success = true
	return

}

func describeNetwork(name string) (exists bool, ID string, err error) {
	cli, err := client.NewEnvClient()
	ctx := context.Background()
	if err != nil {
		log.Panicln(fmt.Sprintf("Error getting docker client environment: %s", err))
		return false, "", err
	}

	list, err := cli.NetworkList(ctx, types.NetworkListOptions{})
	if len(list) == 0 {
		return false, "", err
	}

	for _, network := range list {
		if network.Name == name {
			return true, network.ID, err
		}
	}
	exists = false
	return
}

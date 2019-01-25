package controller

import (
	"context"
	"fmt"
	"github.com/docker/docker/api/types"
	"github.com/docker/docker/api/types/swarm"
	"github.com/docker/docker/client"
	"github.com/plancks-cloud/plancks-docker/model"
	"log"
)

func CreateService(service *model.Service) {
	cli, err := client.NewEnvClient()
	ctx := context.Background()
	if err != nil {
		log.Panicln(fmt.Sprintf("Error getting docker client environment: %s", err))
	}

	replicas := uint64(service.Replicas)

	spec := swarm.ServiceSpec{
		Annotations: swarm.Annotations{
			Name: service.Name,
		},
		Mode: swarm.ServiceMode{
			Replicated: &swarm.ReplicatedService{
				Replicas: &replicas,
			},
		},
		TaskTemplate: swarm.TaskSpec{
			ContainerSpec: swarm.ContainerSpec{
				Image: service.Image,
			},
			Resources: &swarm.ResourceRequirements{
				Limits: &swarm.Resources{
					MemoryBytes: int64(service.MemoryLimit * 1024 * 1024),
				},
			},
		},
	}

	_, err = cli.ServiceCreate(
		ctx,
		spec,
		types.ServiceCreateOptions{},
	)

	if err != nil {
		log.Fatalf("Error creating docker service: %s", err)
	}
}

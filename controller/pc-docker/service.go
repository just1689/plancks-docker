package pc_docker

import (
	"context"
	"fmt"
	"github.com/docker/docker/api/types"
	"github.com/docker/docker/api/types/filters"
	"github.com/docker/docker/api/types/swarm"
	"github.com/docker/docker/client"
	"github.com/plancks-cloud/plancks-cloud/model"
	pcmodel "github.com/plancks-cloud/plancks-docker/model"
	"log"
	"sort"
)

// CreateService creates a service
func CreateService(service *model.Service) (err error) {
	cli, err := client.NewEnvClient()
	ctx := context.Background()
	if err != nil {
		log.Printf("Error getting docker client environment: %s", err)
		return err
	}

	replicas := uint64(service.Replicas)

	networkName := service.Network
	if networkName != "" {
		err = createNetwork(networkName)
	} else {
		networkName = DefaultNetwork
		err = createNetwork(networkName)
	}
	if err != nil {
		log.Printf("Error occurred while creating the network %s: %s", networkName, err)
	}

	var nets []swarm.NetworkAttachmentConfig
	nets = append(nets, swarm.NetworkAttachmentConfig{Target: DefaultNetwork})

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
			Networks: nets,
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
		log.Printf("Error creating docker service: %s", err)
		return err
	}
	return err
}

func GetAllServices() (results []model.Service, err error) {

	serviceStates, err := GetAllServiceStates()
	if err != nil {
		return nil, err
	}

	for _, serviceState := range serviceStates {
		results = append(results, model.Service{
			ID:          serviceState.ID,
			Image:       serviceState.Image,
			Name:        serviceState.Name,
			Replicas:    int(serviceState.ReplicasRequired),
			MemoryLimit: serviceState.MemoryLimit})
	}

	return
}

// DockerServices gets all running docker services
func GetAllServiceStates() (results []pcmodel.ServiceState, err error) {

	cli, err := client.NewEnvClient()

	ctx := context.Background()

	if err != nil {
		log.Println(fmt.Sprintf("Error getting docker client environment: %s", err))
		return nil, err
	}

	services, err := cli.ServiceList(context.Background(), types.ServiceListOptions{})
	if err != nil {
		log.Println(fmt.Sprintf("Error getting docker client environment: %s", err))
		return nil, err
	}

	sort.Sort(pcmodel.ByName(services))
	if len(services) > 0 {
		// only non-empty services and not quiet, should we call TaskList and NodeList api
		taskFilter := filters.NewArgs()
		for _, service := range services {
			taskFilter.Add("service", service.ID)
		}

		tasks, err := cli.TaskList(ctx, types.TaskListOptions{Filters: taskFilter})
		if err != nil {
			log.Println("Error getting tasks")
			return nil, err
		}

		nodes, err := cli.NodeList(ctx, types.NodeListOptions{})
		if err != nil {
			log.Println("Error getting nodes")
			return nil, err
		}

		info := TotalReplicas(services, nodes, tasks)

		for _, item := range info {
			results = append(results, item)
		}
	}
	return
}

// TotalReplicas returns the total number of replicas running for a service
func TotalReplicas(services []swarm.Service, nodes []swarm.Node, tasks []swarm.Task) map[string]pcmodel.ServiceState {
	running := map[string]int{}
	tasksNoShutdown := map[string]int{}
	activeNodes := make(map[string]struct{})
	replicaState := make(map[string]pcmodel.ServiceState)

	for _, n := range nodes {
		if n.Status.State != swarm.NodeStateDown {
			activeNodes[n.ID] = struct{}{}
		}
	}

	for _, task := range tasks {
		if task.DesiredState != swarm.TaskStateShutdown {
			tasksNoShutdown[task.ServiceID]++
		}
		if _, nodeActive := activeNodes[task.NodeID]; nodeActive && task.Status.State == swarm.TaskStateRunning {
			running[task.ServiceID]++
		}
	}

	for _, service := range services {
		if service.Spec.Mode.Replicated != nil && service.Spec.Mode.Replicated.Replicas != nil {
			//TODO Check if this is a valid way of getting the memory limits. Should probably be a pointer.
			memLimit := 0
			if service.Spec.TaskTemplate.Resources.Limits != nil {
				memLimit = int(service.Spec.TaskTemplate.Resources.Limits.MemoryBytes)
			}
			replicaState[service.ID] = pcmodel.ServiceState{
				ID:               service.ID,
				Name:             service.Spec.Name,
				Image:            service.Spec.TaskTemplate.ContainerSpec.Image,
				ReplicasRunning:  running[service.ID],
				MemoryLimit:      memLimit,
				ReplicasRequired: *service.Spec.Mode.Replicated.Replicas}
		}
	}
	return replicaState
}

func DeleteServices(services []pcmodel.ServiceState) (err error) {
	cli, err := client.NewEnvClient()
	ctx := context.Background()
	if err != nil {
		log.Printf("Error getting docker client environment: %s", err)
		return err
	}

	runningServices, err := serviceIdFromName(services)

	if err != nil {
		log.Printf("Error getting service IDs: %s", err)
	}

	for serviceId, serviceName := range runningServices {
		log.Printf("🔥  Removing service: %s", serviceName)
		err := cli.ServiceRemove(ctx, serviceId)
		if err != nil {
			log.Printf("Error deleting service %s: %s", serviceName, err)
			return err
		}
	}
	return
}

func serviceIdFromName(services []pcmodel.ServiceState) (deletables map[string]string, err error) {
	runningServices, err := GetAllServiceStates()
	if err != nil {
		log.Printf("Error getting services: %s", err)
		return nil, err
	}

	for _, service := range services {
		for _, runningService := range runningServices {
			if service.Name == runningService.Name {
				deletables[runningService.ID] = service.Name
			}
		}
	}

	return deletables, err
}

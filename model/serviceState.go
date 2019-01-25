package model

import (
	"github.com/docker/docker/api/types/swarm"
	"vbom.ml/util/sortorder"
)

type ServiceState struct {
	ID               string
	Name             string
	Image            string `json:"image"`
	ReplicasRunning  int
	ReplicasRequired uint64
}

type ByName []swarm.Service

func (n ByName) Len() int           { return len(n) }
func (n ByName) Swap(i, j int)      { n[i], n[j] = n[j], n[i] }
func (n ByName) Less(i, j int) bool { return sortorder.NaturalLess(n[i].Spec.Name, n[j].Spec.Name) }

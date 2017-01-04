package container

import (
	"github.com/aws/aws-sdk-go/service/ecs"

)

type Service struct {
	Name string
	Arn string
}

func GetServiceInfo(svc *ecs.ECS,clusterName string) []Service {
	return nil
}

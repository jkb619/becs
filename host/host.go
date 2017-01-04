package host

import (
	"becs/container"
	"github.com/aws/aws-sdk-go/service/ecs"
	"github.com/aws/aws-sdk-go/aws"
	"fmt"
)

type Host struct {
	Arn string
	ContainerList []container.Container
}


func GetHostInfo(svc *ecs.ECS,clusterName string) []Host {
	list_params := &ecs.ListContainerInstancesInput{
		Cluster: aws.String(clusterName),
		//aws.String("String"), // Required
		// More values...
		//},
	}
	pageNum := 0
	hostList := []Host{}
	err := svc.ListContainerInstancesPages(list_params,
		func(page *ecs.ListContainerInstancesOutput, lastPage bool) bool {
			pageNum++
			for _, arn := range page.ContainerInstanceArns {
				hostList=append(hostList,Host{*arn,[]container.Container{}})
			}
			return pageNum > 0
		})

	if err != nil {
		fmt.Println(err.Error())
		return nil
	}

	return hostList
}

package host

import (
	"becs/container"
	"github.com/aws/aws-sdk-go/service/ecs"
	"github.com/aws/aws-sdk-go/aws"
	"fmt"
)

type Host struct {
	Arn string
	Ec2Id string
	ServiceList []container.Service
}

func GetHostInfo(svc *ecs.ECS,clusterName string) []Host {
	list_params := &ecs.ListContainerInstancesInput{
		Cluster: aws.String(clusterName),
	}
	pageNum := 0
	hostList := []Host{}
	err := svc.ListContainerInstancesPages(list_params,
		func(page *ecs.ListContainerInstancesOutput, lastPage bool) bool {
			pageNum++
			for _, arn := range page.ContainerInstanceArns {
				ec2list_params := &ecs.DescribeContainerInstancesInput{
					ContainerInstances: []*string{
						aws.String(*arn),
					},
					Cluster: aws.String(clusterName),
				}
				ec2list_resp, err := svc.DescribeContainerInstances(ec2list_params)
				ec2id := ec2list_resp.ContainerInstances[0].Ec2InstanceId

				if err != nil {
					fmt.Println(err.Error())
					return false
				}

				hostList=append(hostList,Host{*arn,*ec2id,[]container.Service{}})
			}
			return pageNum > 0
		})

	if err != nil {
		fmt.Println(err.Error())
		return nil
	}

	return hostList
}

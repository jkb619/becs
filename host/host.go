package host

import (
	"becs/container"
	"github.com/aws/aws-sdk-go/service/ecs"
	"github.com/aws/aws-sdk-go/aws"
	"fmt"
	"github.com/aws/aws-sdk-go/service/ec2"
)

type Host struct {
	Arn string
	Ec2Id string
	Ec2Ip string
	ServiceList []container.Service
}

func getContainerInstanceIpAddress(ec2_svc *ec2.EC2, instanceId string) string {
	var ec2ip string = ""
	ec2list_params := &ec2.DescribeInstancesInput{
		InstanceIds: []*string{
			aws.String(instanceId),
		},
	}
	ec2list_resp, err := ec2_svc.DescribeInstances(ec2list_params)
	if err != nil {
		fmt.Println(err.Error())
		return ""
	}
	if len(ec2list_resp.Reservations) > 0 {
		ec2ip = *ec2list_resp.Reservations[0].Instances[0].PrivateIpAddress
	}
	return ec2ip
}

func GetHostInfo(svc *ecs.ECS,ec2_svc *ec2.EC2, clusterName string) []Host {
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
				ec2ip := getContainerInstanceIpAddress(ec2_svc,*ec2id)

				if err != nil {
					fmt.Println(err.Error())
					return false
				}

				hostList=append(hostList,Host{*arn,*ec2id,ec2ip,[]container.Service{}})
			}
			return pageNum > 0
		})

	if err != nil {
		fmt.Println(err.Error())
		return nil
	}

	return hostList
}

package host

import (
	"github.com/aws/aws-sdk-go/service/ecs"
	"github.com/aws/aws-sdk-go/aws"
	"fmt"
	"github.com/aws/aws-sdk-go/service/ec2"
	"sync"
	"becs/task"
)

type Host struct {
	Arn string
	Ec2Id string
	Ec2Ip string
	Tasks task.Tasks
}

type Hosts struct {
	HostList []Host
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

func (h *Hosts) getHostsGoroutine(svc *ecs.ECS, ec2_svc *ec2.EC2, instanceArn string, clusterName string, wg *sync.WaitGroup, ch *(chan []Host)) {
	defer wg.Done()
	ec2list_params := &ecs.DescribeContainerInstancesInput{
		ContainerInstances: []*string{
			aws.String(instanceArn),
		},
		Cluster: aws.String(clusterName),
	}
	ec2list_resp, err := svc.DescribeContainerInstances(ec2list_params)
	ec2id := ec2list_resp.ContainerInstances[0].Ec2InstanceId
	ec2ip := getContainerInstanceIpAddress(ec2_svc,*ec2id)

	if err != nil {
		fmt.Println(err.Error())
		return
	}
	h.HostList=append(h.HostList,Host{instanceArn,*ec2id,ec2ip,task.Tasks{}})
	*ch<- h.HostList
}

func (h *Hosts) GetHostInfo(svc *ecs.ECS,ec2_svc *ec2.EC2, clusterName string)  {
	var host_ch = make(chan []Host)
	var host_wg sync.WaitGroup
	pageNum := 0

	list_params := &ecs.ListContainerInstancesInput{
		Cluster: aws.String(clusterName),
	}
	err := svc.ListContainerInstancesPages(list_params,
		func(page *ecs.ListContainerInstancesOutput, lastPage bool) bool {
			pageNum++
			for _, instanceArn := range page.ContainerInstanceArns {
				host_wg.Add(1)
				go h.getHostsGoroutine(svc, ec2_svc,*instanceArn,clusterName,&host_wg,&host_ch)
			}
			return pageNum > 0
		})

	if err != nil {
		fmt.Println(err.Error())
		return
	}
	go func() {
		host_wg.Wait()
		close(host_ch)
	}()
	for m := range host_ch {
		h.HostList=append(h.HostList,m[0])
	}
}

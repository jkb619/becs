package task

import (
)
import (
	"github.com/aws/aws-sdk-go/service/ecs"
	"github.com/aws/aws-sdk-go/service/ec2"
	"sync"
	"github.com/aws/aws-sdk-go/aws"
	"fmt"
	"strings"
)

type Task struct {
	Name string
	Arn string
	//ClusterArn string
	//ClusterName string
	//ContainerInstanceArn string
	//ContainerEc2Id string
	//ContainerEc2Ip string
}

type Tasks struct {
	TaskList []Task
}
/*
func getContainerInstanceId(svc *ecs.ECS,clusterName string, containerInstanceArn string) string {
	var ec2id *string
	ec2list_params := &ecs.DescribeContainerInstancesInput{
		ContainerInstances: []*string{
			aws.String(containerInstanceArn),
		},
		Cluster: aws.String(clusterName),
	}
	ec2list_resp, err := svc.DescribeContainerInstances(ec2list_params)
	if len(ec2list_resp.ContainerInstances) > 0 {
		ec2id = ec2list_resp.ContainerInstances[0].Ec2InstanceId
	} else { *ec2id = ""}

	if err != nil {
		fmt.Println(err.Error())
		return ""
	}

	return *ec2id
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
*/
func (t *Tasks) describeTasks(svc *ecs.ECS, ec2_svc *ec2.EC2, taskArn *string, clusterName string, taskFilter string, wg *sync.WaitGroup, ch *(chan []Task)) {
	defer (*wg).Done()
	tasklist_params := &ecs.DescribeTasksInput{
		Tasks: []*string{
			aws.String(*taskArn),
		},
		Cluster: aws.String(clusterName),
	}
	taskDesc_resp, err := svc.DescribeTasks(tasklist_params)
	if err != nil {
		fmt.Println(err.Error())
		return
	}
	taskName := *taskDesc_resp.Tasks[0].Containers[0].Name
	if strings.Contains(taskName,taskFilter) {
//		clusterArn := *taskDesc_resp.Tasks[0].ClusterArn
//		containerInstanceArn := *taskDesc_resp.Tasks[0].ContainerInstanceArn

//		containerInstanceId := getContainerInstanceId(svc, clusterName, containerInstanceArn)
//		containerInstanceIp := getContainerInstanceIpAddress(ec2_svc,containerInstanceId)
		taskList:=[]Task{}
//		taskList = append(taskList, Task{taskName, *taskArn, clusterArn, clusterName, containerInstanceArn, containerInstanceId, containerInstanceIp})
		taskList = append(taskList,Task{taskName, *taskArn})
		*ch <- taskList
	}
}

func (t *Tasks) GetTaskInfo (svc *ecs.ECS, ec2_svc *ec2.EC2, clusterName string,instanceArn string, taskFilter string) {
	pageNum := 0
	var task_ch=make(chan []Task)
	var task_wg sync.WaitGroup

	list_params := &ecs.ListTasksInput{
		Cluster: aws.String(clusterName),
		ContainerInstance: aws.String(instanceArn),
	}
	err := svc.ListTasksPages(list_params,
		func(page *ecs.ListTasksOutput, lastPage bool) bool {
			pageNum++
			for _, taskArn := range page.TaskArns {
				task_wg.Add(1)
				go t.describeTasks(svc, ec2_svc, taskArn, clusterName, taskFilter, &task_wg, &task_ch)
			}
			return pageNum > 0
		})

	if err != nil {
		fmt.Println(err.Error())
		return
	}
	go func() {
		task_wg.Wait()
		close(task_ch)
	}()
	for m := range task_ch {
//		taskList = append(taskList, Task{m[index].Name, m[index].Arn, m[index].ClusterArn, m[index].ClusterName, m[index].ContainerInstanceArn, m[index].ContainerEc2Id, m[index].ContainerEc2Ip}) //,hostList})
		t.TaskList = append(t.TaskList,m[0])
	}
}

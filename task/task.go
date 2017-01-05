package task

import (
	"fmt"
	"github.com/aws/aws-sdk-go/service/ecs"
	"github.com/aws/aws-sdk-go/aws"
	"strings"
	"github.com/aws/aws-sdk-go/service/ec2"
	"sync"
)

type Task struct {
	Name string
	Arn string
	ClusterArn string
	ClusterName string
	ContainerInstanceArn string
	ContainerEc2Id string
}

var ch=make(chan []Task)
var wg sync.WaitGroup

func getContainerInstanceId(svc *ecs.ECS,clusterName string, containerInstanceArn string) string {
	ec2list_params := &ecs.DescribeContainerInstancesInput{
		ContainerInstances: []*string{
			aws.String(containerInstanceArn),
		},
		Cluster: aws.String(clusterName),
	}
	ec2list_resp, err := svc.DescribeContainerInstances(ec2list_params)
	ec2id := ec2list_resp.ContainerInstances[0].Ec2InstanceId

	if err != nil {
		fmt.Println(err.Error())
		return ""
	}

	return *ec2id
}

func getContainerInstanceIpAddress(svc *ecs.ECS,ec2_svc *ec2.EC2, clusterName string, containerInstanceArn string) string {
	ec2list_params := &ecs.DescribeContainerInstancesInput{
		ContainerInstances: []*string{
			aws.String(containerInstanceArn),
		},
		Cluster: aws.String(clusterName),
	}
	ec2list_resp, err := svc.DescribeContainerInstances(ec2list_params)
	ec2id := ec2list_resp.ContainerInstances[0].Ec2InstanceId

	if err != nil {
		fmt.Println(err.Error())
		return ""
	}

	return *ec2id
}

func (t *Task) describeTasks(svc *ecs.ECS, ec2_svc *ec2.EC2, taskArn *string, clusterName string, taskFilter string) {
	defer wg.Done()
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
		clusterArn := *taskDesc_resp.Tasks[0].ClusterArn
		containerInstanceArn := *taskDesc_resp.Tasks[0].ContainerInstanceArn

		containerInstanceId := getContainerInstanceId(svc, clusterName, containerInstanceArn)
		taskList:=[]Task{}
		taskList = append(taskList, Task{taskName, *taskArn, clusterArn, clusterName, containerInstanceArn, containerInstanceId})
		ch <- taskList
	}
}

func (t *Task) GetTaskInfo (svc *ecs.ECS, ec2_svc *ec2.EC2, clusterName string,taskFilter string) []Task {
	list_params := &ecs.ListTasksInput{
		Cluster: aws.String(clusterName),
	}
	pageNum := 0
	taskList:=[]Task{}
	err := svc.ListTasksPages(list_params,
		func(page *ecs.ListTasksOutput, lastPage bool) bool {
			pageNum++
			for _, taskArn := range page.TaskArns {
/*
				tasklist_params := &ecs.DescribeTasksInput{
					Tasks: []*string{
						aws.String(*taskArn),
					},
					Cluster: aws.String(clusterName),
				}
				taskDesc_resp, err := svc.DescribeTasks(tasklist_params)
				if err != nil {
					fmt.Println(err.Error())
					return false
				}
				taskName := *taskDesc_resp.Tasks[0].Containers[0].Name
				if strings.Contains(taskName,taskFilter) {
					clusterArn := *taskDesc_resp.Tasks[0].ClusterArn
					containerInstanceArn := *taskDesc_resp.Tasks[0].ContainerInstanceArn

					containerInstanceId := getContainerInstanceId(svc, clusterName, containerInstanceArn)
					taskList = append(taskList, Task{taskName, *taskArn, clusterArn, clusterName, containerInstanceArn, containerInstanceId})
				}
*/
				wg.Add(1)
				go t.describeTasks(svc, ec2_svc, taskArn, clusterName, taskFilter)
			}
			return pageNum > 0
		})

	if err != nil {
		fmt.Println(err.Error())
		return nil
	}
	go func() {
		wg.Wait()
		close(ch)
	}()
	index:=0
	for m := range ch {
		taskList = append(taskList, Task{m[index].Name, m[index].Arn, m[index].ClusterArn, m[index].ClusterName, m[index].ContainerInstanceArn, m[index].ContainerEc2Id}) //,hostList})
	}
	return taskList
}

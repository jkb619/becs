package task

import (
	"fmt"
	"github.com/aws/aws-sdk-go/service/ecs"
	"github.com/aws/aws-sdk-go/aws"
	"strings"
)

type Task struct {
	Name string
	Arn string
	ClusterArn string
	ContainerInstanceArn string
	ContainerEc2Id string
}

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

func GetTaskInfo (svc *ecs.ECS, clusterName string,taskFilter string) []Task {
	list_params := &ecs.ListTasksInput{
		Cluster: aws.String(clusterName),
	}
	pageNum := 0
	taskList := []Task{}
	err := svc.ListTasksPages(list_params,
		func(page *ecs.ListTasksOutput, lastPage bool) bool {
			pageNum++
			for _, taskArn := range page.TaskArns {
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
					taskList = append(taskList, Task{taskName, *taskArn, clusterArn, containerInstanceArn, containerInstanceId})
				}
			}
			return pageNum > 0
		})

	if err != nil {
		fmt.Println(err.Error())
		return nil
	}

	return taskList
}

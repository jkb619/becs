package task

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
}

type Tasks struct {
	TaskList []Task
}

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
	if strings.Contains(taskName,taskFilter) || strings.Contains(*taskArn,taskFilter){
		taskList:=[]Task{}
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
		t.TaskList = append(t.TaskList,m[0])
	}
}

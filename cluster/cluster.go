package cluster

import (
	"github.com/aws/aws-sdk-go/service/ecs"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/aws"
	"fmt"
//	"becs/host"
	"becs/task"
)

type Cluster struct {
	Arn string
	Name string
	TaskList []task.Task
//	HostList []host.Host
}

type Clusters struct {
	ClusterList []Cluster
}

func (c *Clusters) GetClusterInfo(svc *ecs.ECS) {
	list_params := &ecs.ListClustersInput{
	}
	pageNum := 0
	err := svc.ListClustersPages(list_params,
		func(page *ecs.ListClustersOutput, lastPage bool) bool {
			pageNum++
			for _, arn := range page.ClusterArns {
				describe_params := &ecs.DescribeClustersInput{
					Clusters: []*string{
						aws.String(*arn),
					},
				}
				name,_ := svc.DescribeClusters(describe_params)
//				hostList := host.GetHostInfo(svc,*name.Clusters[0].ClusterName)
				taskList := task.GetTaskInfo(svc,*name.Clusters[0].ClusterName)
				c.ClusterList=append(c.ClusterList,Cluster{*arn,*name.Clusters[0].ClusterName,taskList})//,hostList})
			}
			return pageNum > 0
		})

	if err != nil {
		fmt.Println(err.Error())
		return
	}
}

func Cluster_list() {
	clust := new(Clusters)
	sess, err := session.NewSession(&aws.Config{Region: aws.String("us-east-1")})
	if err != nil {
		fmt.Println("failed to create session,", err)
		return
	}
	svc := ecs.New(sess)
	clust.GetClusterInfo(svc)
	for _, element := range clust.ClusterList {
		fmt.Println(element.Name)
		for _,taskElement := range element.TaskList {
//			fmt.Println("-----",taskElement.Arn)//," : ",taskElement.Ec2Id)
			fmt.Println("-----",taskElement.Name)
//			fmt.Println("-----",taskElement.ClusterArn)
			fmt.Println("----------",taskElement.ContainerInstanceArn," : ",taskElement.ContainerEc2Id)
		}
	}
}
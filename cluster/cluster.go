package cluster

import (
	"github.com/aws/aws-sdk-go/service/ecs"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/aws"
	"fmt"
//	"becs/host"
	"becs/task"
	"strings"
	"github.com/aws/aws-sdk-go/service/ec2"
	"sync"
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

var ch=make(chan []task.Task)
var wg sync.WaitGroup

func (c *Clusters) describeClusters(svc *ecs.ECS, ec2_svc *ec2.EC2, clusterArn *string, taskFilter string) {
	defer wg.Done()
	describe_params := &ecs.DescribeClustersInput{
		Clusters: []*string{
			aws.String(*clusterArn),
		},
	}
	name, _ := svc.DescribeClusters(describe_params)
	localTask:= new(task.Task)
	taskList := localTask.GetTaskInfo(svc, ec2_svc, *name.Clusters[0].ClusterName, taskFilter)
	ch <- taskList
}

func (c *Clusters) GetClusterInfo(svc *ecs.ECS,ec2_svc *ec2.EC2,clusterFilter string, taskFilter string) {
	list_params := &ecs.ListClustersInput{}

	pageNum := 0
	err := svc.ListClustersPages(list_params,
		func(page *ecs.ListClustersOutput, lastPage bool) bool {
			pageNum++
			for _, arn := range page.ClusterArns {
				/*
				if strings.Contains(*arn,clusterFilter) {
						describe_params := &ecs.DescribeClustersInput{
						Clusters: []*string{
							aws.String(*arn),
						},
					}
					name, _ := svc.DescribeClusters(describe_params)
					//				hostList := host.GetHostInfo(svc,*name.Clusters[0].ClusterName)
					taskList := task.GetTaskInfo(svc,ec2_svc,*name.Clusters[0].ClusterName,taskFilter)
*/
				if strings.Contains(*arn,clusterFilter) {
					wg.Add(1)
					go c.describeClusters(svc, ec2_svc, arn, taskFilter)
				}

				//c.ClusterList = append(c.ClusterList, Cluster{*arn, *name.Clusters[0].ClusterName, taskList}) //,hostList})
			}
			return pageNum > 0
		})

	if err != nil {
		fmt.Println(err.Error())
		return
	}
	go func() {
		wg.Wait()
		close(ch)
	}()
	index:=0
	for m := range ch {
		c.ClusterList = append(c.ClusterList, Cluster{m[index].ClusterArn, m[index].ClusterName, m}) //,hostList})
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
	ec2_svc := ec2.New(sess)
	clusterFilter:="deploytest"
	taskFilter:=""

	clust.GetClusterInfo(svc,ec2_svc,clusterFilter,taskFilter)
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
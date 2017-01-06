package cluster

import (
	"github.com/aws/aws-sdk-go/service/ecs"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/aws"
	"fmt"
	"strings"
	"becs/host"
	"sync"
	"github.com/aws/aws-sdk-go/service/ec2"
	"strconv"
)

type QueryLevel uint8

const (
	LevelCluster=QueryLevel(iota)
	LevelHost
	LevelTask
)

func (s QueryLevel) String() string {
	var name = []string{"cluster","host","task"}
	var i = uint8(s)
	switch {
	case i <= uint8(LevelTask):
		return name[i]
	default:
		return strconv.Itoa(int(i))
	}
}

type Cluster struct {
	Arn string
	Name string
//	TaskList []task.Task
	Hosts host.Hosts
}

type Clusters struct {
	ClusterList []Cluster
}

func (c *Clusters) getClusterGoroutine(svc *ecs.ECS,clusterArn string, wg *sync.WaitGroup, ch *(chan []Cluster)) {
	defer (*wg).Done()
	describe_params := &ecs.DescribeClustersInput{
		Clusters: []*string{
			aws.String(clusterArn),
		},
	}
	clusterDescription, _ := svc.DescribeClusters(describe_params)
	clusterList:=[]Cluster{}
	clusterList=append(clusterList,Cluster{clusterArn,*clusterDescription.Clusters[0].ClusterName,host.Hosts{}})
	*ch<-clusterList
}
func (c *Clusters) getClusterInfo(svc *ecs.ECS,clusterFilter string) {
	var cluster_ch = make(chan []Cluster)
	var cluster_wg sync.WaitGroup
	pageNum := 0
	list_params := &ecs.ListClustersInput{}

	err := svc.ListClustersPages(list_params,
		func(page *ecs.ListClustersOutput, lastPage bool) bool {
			pageNum++
			for _, clusterArn := range page.ClusterArns {
				if strings.Contains(*clusterArn, clusterFilter) {
					cluster_wg.Add(1)
					go c.getClusterGoroutine(svc,*clusterArn,&cluster_wg,&cluster_ch)
				}
			}
			return pageNum > 0
		})
	if err != nil {
		fmt.Println(err.Error())
		return
	}
	go func() {
		cluster_wg.Wait()
		close(cluster_ch)
	}()
	for m := range cluster_ch {
		c.ClusterList=append(c.ClusterList,m[0])
	}
}

func Cluster_list(clusterFilter string, hostFilter string, taskFilter string, level QueryLevel) {
	clusters := new(Clusters)
	sess, err := session.NewSession(&aws.Config{Region: aws.String("us-east-1")})
	if err != nil {
		fmt.Println("failed to create session,", err)
		return
	}
	svc := ecs.New(sess)
	ec2_svc := ec2.New(sess)

	clusters.getClusterInfo(svc,clusterFilter)
	if level > LevelCluster {
		for i := 0; i < len(clusters.ClusterList); i++ {
			clusters.ClusterList[i].Hosts.GetHostInfo(svc, ec2_svc, clusters.ClusterList[i].Name)
			if level > LevelHost {
				for j :=0;j<len(clusters.ClusterList[i].Hosts.HostList);j++ {
					//if (clusters.ClusterList[i].Hosts.HostList[j].)
					clusters.ClusterList[i].Hosts.HostList[j].Tasks.GetTaskInfo(svc,ec2_svc,clusters.ClusterList[i].Name, clusters.ClusterList[i].Hosts.HostList[j].Arn, taskFilter)
				}
			}
		}
	}

	for _, cluster := range clusters.ClusterList {
		fmt.Println(cluster.Name," : ",cluster.Arn)
		if (level > LevelCluster) {
			for _, host := range cluster.Hosts.HostList {
				fmt.Println("-----"," : ", host.Ec2Id, " : ", host.Ec2Ip)
				if (level > LevelHost) {
					for _, taskElement := range host.Tasks.TaskList {
						fmt.Println("----------", taskElement.Name, " : ", taskElement.Arn)
					}
				}
			}
		}
	}
}
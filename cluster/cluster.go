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

func (c *Clusters) List(clusterFilter string, hostFilter string, taskFilter string, level QueryLevel) {
	sess, err := session.NewSession(&aws.Config{Region: aws.String("us-east-1")})
	if err != nil {
		fmt.Println("failed to create session,", err)
		return
	}
	svc := ecs.New(sess)
	ec2_svc := ec2.New(sess)

	c.getClusterInfo(svc,clusterFilter)
	if level > LevelCluster {
		for i := 0; i < len(c.ClusterList); i++ {
			c.ClusterList[i].Hosts.GetHostInfo(svc, ec2_svc, c.ClusterList[i].Name, hostFilter)
			if level > LevelHost {
				for j :=0;j<len(c.ClusterList[i].Hosts.HostList);j++ {
					c.ClusterList[i].Hosts.HostList[j].Tasks.GetTaskInfo(svc,ec2_svc,c.ClusterList[i].Name, c.ClusterList[i].Hosts.HostList[j].Arn, taskFilter)
				}
			}
		}
	}

	displayArray:=[]string{}
	for _, cluster := range c.ClusterList {
		switch level {
		case LevelCluster:
			displayArray=append(displayArray,cluster.Name+" : "+cluster.Arn)
		case LevelHost:
			if len(cluster.Hosts.HostList) > 0 {
				displayArray=append(displayArray,cluster.Name+" : "+cluster.Arn)
				for _, hostLoop := range cluster.Hosts.HostList {
					displayArray=append(displayArray,"-----"+hostLoop.Ec2Id+" : "+hostLoop.Ec2Ip)
				}
			}
		case LevelTask:
			printClusterHeader:=true
			printHostHeader:=true
			for _, hostLoop := range cluster.Hosts.HostList {
				for _, taskElement := range hostLoop.Tasks.TaskList {
					if printClusterHeader {
						fmt.Println(cluster.Name," : ",cluster.Arn)
						printClusterHeader=false
					}
					if printHostHeader {
						fmt.Println("-----", hostLoop.Ec2Id, " : ", hostLoop.Ec2Ip)
						printHostHeader=false
					}
					fmt.Println("----------",taskElement.Name," : ",taskElement.Arn)
				}
			}
		}
	}
}
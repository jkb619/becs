package cluster

import (
	"github.com/aws/aws-sdk-go/service/ecs"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/aws"
	"fmt"
	"becs/host"
)

type Cluster struct {
	Arn string
	Name string
	HostList []host.Host
}

type Clusters struct {
	ClusterList []Cluster
}

func (c *Clusters) GetClusterInfo(svc *ecs.ECS) {
	list_params := &ecs.ListClustersInput{
		//Clusters: []*string{
		//aws.String("String"), // Required
		// More values...
		//},
	}
	pageNum := 0
	err := svc.ListClustersPages(list_params,
		func(page *ecs.ListClustersOutput, lastPage bool) bool {
			pageNum++
			for _, arn := range page.ClusterArns {
				describe_params := &ecs.DescribeClustersInput{
					Clusters: []*string{
						aws.String(*arn), // Required
						// More values...
					},
				}
				name,_ := svc.DescribeClusters(describe_params)
				hostList := host.GetHostInfo(svc,*name.Clusters[0].ClusterName)
				c.ClusterList=append(c.ClusterList,Cluster{*arn,*name.Clusters[0].ClusterName,hostList})
			}
			return pageNum > 0
		})

	if err != nil {
		fmt.Println(err.Error())
		return
	}
}

func Cluster_list() {
	clusters := new(Clusters)
	sess, err := session.NewSession(&aws.Config{Region: aws.String("us-east-1")})
	if err != nil {
		fmt.Println("failed to create session,", err)
		return
	}
	svc := ecs.New(sess)
	clusters.GetClusterInfo(svc)
	for _, element := range clusters.ClusterList {
		fmt.Println(element.Name)
//		fmt.Println(element.Arn)
		for _,hostElement := range element.HostList {
			fmt.Println("-----",hostElement.Arn)
		}
	}
}
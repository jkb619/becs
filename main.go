package main

import (
	"github.com/aws/aws-sdk-go/aws/session"
	"fmt"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/ecs"
)

func main() {
	sess, err := session.NewSession(&aws.Config{Region: aws.String("us-east-1")})
	if err != nil {
		fmt.Println("failed to create session,", err)
		return
	}

	//Specify profile to load for the session's config
	//sess, err := session.NewSessionWithOptions(session.Options{
	//	Profile: "default",
	//})

	svc := ecs.New(sess)

	params := &ecs.ListClustersInput{
		//Clusters: []*string{
			//aws.String("String"), // Required
			// More values...
		//},
	}

	pageNum := 0
	clusterList := make([]string,10)
	err2 := svc.ListClustersPages(params,
		func(page *ecs.ListClustersOutput, lastPage bool) bool {
			pageNum++
			for _,name := range page.ClusterArns {
				clusterList=append(clusterList,*name)
			}
			//fmt.Println(page)
			return pageNum > 0
		})

	if err2 != nil {
		// Print the error, cast err to awserr.Error to get the Code and
		// Message from an error.
		fmt.Println(err2.Error())
		return
	}

	// Pretty-print the response data.
	for _,element := range clusterList {
		fmt.Println(element)
	}

	//fmt.Println(resp)
}

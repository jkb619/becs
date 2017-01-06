package main

import (
	"flag"
	"fmt"
	"os"
	"becs/cluster"
	"strings"
)

func main() {
	listCommand := flag.NewFlagSet("list",flag.ExitOnError)
	levelFlag := listCommand.String("level","task","what level to delve: cluster/task (defaults to task)")
	clusterFilterFlag := listCommand.String("cluster","","cluster substring to match")
	hostFilterFlag := listCommand.String("host","","host substring to match")
	taskFilterFlag := listCommand.String("task","","task substring to match")

//	sshCommand := flag.NewFlagSet("list",flag.ExitOnError)
//	sshClusterFilterFlag := sshCommand.String("cluster","","cluster substring to match")
//	sshTaskFilterFlag := sshCommand.String("task","","task substring to match")
//	userFlag := sshCommand.String("user","","user to login as")
//	passwordFlag := sshCommand.String("password","","password for user")
//	toSendFlag := sshCommand.String("send", "", "what to send via ssh")

	if len(os.Args) == 1 {
		fmt.Println("usage: becs <command> [<args>]")
		fmt.Println("where <command> is :")
		fmt.Println("list,ssh,scp")
		os.Exit(2)
	}
	level:=cluster.LevelCluster
	switch os.Args[1] {
	case "list":
		listCommand.Parse(os.Args[2:])
		if !strings.Contains(*levelFlag,"cluster") &&
			!strings.Contains(*levelFlag,"task") {
			fmt.Println("-level must be either 'cluster','host', or 'task'")
			os.Exit(2)
		} else {
			switch *levelFlag {
			case "cluster":
				level=cluster.LevelCluster
			case "host":
				level=cluster.LevelHost
			case "task":
				level=cluster.LevelTask
			}
		}
//	case "ssh":
//		sshCommand.Parse(os.Args[2:])
	default:
		fmt.Printf("%q is invalid.\n",os.Args[1])
		os.Exit(2)
	}

	if listCommand.Parsed() {
		//fmt.Println("list + ",*clusterFilterFlag," + ",*taskFilterFlag)
		cluster.Cluster_list(*clusterFilterFlag,*hostFilterFlag,*taskFilterFlag,level)
	}
//	if sshCommand.Parsed() {
//	}
	//cluster.Cluster_list()
}


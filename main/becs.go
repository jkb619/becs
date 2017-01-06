package main

import (
	"flag"
	"fmt"
	"os"
	"becs/cluster"
	"strings"
	"becs/ssh"
)

func main() {
	listCommand := flag.NewFlagSet("list",flag.ExitOnError)
	listLevelFlag := listCommand.String("level","task","what level to delve: cluster/task (defaults to task)")
	listClusterFilterFlag := listCommand.String("cluster","","cluster substring to match")
	listHostFilterFlag := listCommand.String("host","","host substring to match")
	listTaskFilterFlag := listCommand.String("task","","task substring to match")

	sshCommand := flag.NewFlagSet("list",flag.ExitOnError)
	sshMode := sshCommand.String("mode","multi","none/single/multi. 'none' is non-interactive,single is terminal,multi is desktop")
	sshClusterFilterFlag := sshCommand.String("cluster","","cluster substring to match")
	sshHostFilterFlag := sshCommand.String("host","","host substring to match")
	sshTaskFilterFlag := sshCommand.String("task","","task substring to match")
	sshUserFlag := sshCommand.String("user","ec2-user","user to login as")
	sshPasswordFlag := sshCommand.String("password","","password for user")
	sshToSendFlag := sshCommand.String("send", "", "what to send via ssh")

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
		if !strings.Contains(*listLevelFlag,"cluster") &&
			!strings.Contains(*listLevelFlag,"task") {
			fmt.Println("-level must be either 'cluster','host', or 'task'")
			os.Exit(2)
		} else {
			switch *listLevelFlag {
			case "cluster":
				level=cluster.LevelCluster
			case "host":
				level=cluster.LevelHost
			case "task":
				level=cluster.LevelTask
			}
		}
	case "ssh":
		sshCommand.Parse(os.Args[2:])
	default:
		fmt.Printf("%q is invalid.\n",os.Args[1])
		os.Exit(2)
	}

	if listCommand.Parsed() {
		clusters := new(cluster.Clusters)
		clusters.List(*listClusterFilterFlag,*listHostFilterFlag,*listTaskFilterFlag,level)
	}
	if sshCommand.Parsed() {
		clusters := new(cluster.Clusters)
		ecsssh.EcsSSH(clusters,sshMode,sshClusterFilterFlag,sshHostFilterFlag,sshTaskFilterFlag,sshUserFlag,sshPasswordFlag,sshToSendFlag)
	}
}


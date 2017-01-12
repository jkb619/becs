package main

import (
	"flag"
	"fmt"
	"os"
	"becs/cluster"
	"becs/ssh"
)

func main() {
	listCommand := flag.NewFlagSet("list",flag.ExitOnError)
	listLevelFlag := listCommand.String("level","task","what level to delve: cluster/host/task (defaults to task)")
	listClusterFilterFlag := listCommand.String("cluster","","cluster substring to match")
	listHostFilterFlag := listCommand.String("host","","host substring to match")
	listTaskFilterFlag := listCommand.String("task","","task substring to match")
	listVerbosity := listCommand.Bool("v",false,"verbose (true/false. Defaults false)")

	sshCommand := flag.NewFlagSet("ssh",flag.ExitOnError)
	sshTarget := sshCommand.String("target","task", "host/task. Identifies which elements to ssh to.")
	sshMode := sshCommand.String("mode","tmux","tmux / gui / batch. ")
	sshClusterFilterFlag := sshCommand.String("cluster","","cluster substring to match")
	sshHostFilterFlag := sshCommand.String("host","","host substring to match")
	sshTaskFilterFlag := sshCommand.String("task","","task substring to match")
	sshUserFlag := sshCommand.String("user","ec2-user","user to login as")
	//sshPasswordFlag := sshCommand.String("password","","password for user")
	sshCmdToRun := sshCommand.String("cmd", "", "what cmd to send via ssh")

	scpCommand := flag.NewFlagSet("scp",flag.ExitOnError)
	scpTarget := scpCommand.String("target","host", "host/task. Identifies which elements to ssh to. (default host. Task not working currently)")
	scpClusterFilterFlag := scpCommand.String("cluster","","cluster substring to match")
	scpHostFilterFlag := scpCommand.String("host","","host substring to match")
	scpTaskFilterFlag := scpCommand.String("task","","task substring to match")
	scpUserFlag := scpCommand.String("user","ec2-user","user to login as")
	//scpPasswordFlag := scpCommand.String("password","","password for user")
	scpFileToSend := scpCommand.String("file", "", "what file to send via scp")
	scpTargetDir := scpCommand.String("tdir", "/tmp", "what target directory to put the file")
	scpFileRunFlag := scpCommand.Bool("x",false,"run uploaded file on target true/false (default false)")
	scpDeleteFileAfterRun := scpCommand.Bool("d",false,"delete file after running")

	if len(os.Args) == 1 {
		fmt.Println("usage: becs <command> [<args>]")
		fmt.Println("where <command> is :")
		fmt.Println("list,ssh,scp")
		os.Exit(2)
	}

	level:=cluster.LevelCluster
	mode:=ecsssh.ModeTmux
	target:=ecsssh.TargetTask
	switch os.Args[1] {
	case "list":
		listCommand.Parse(os.Args[2:])
		switch *listLevelFlag {
		case "cluster":
			level=cluster.LevelCluster
		case "host":
			level=cluster.LevelHost
		case "task":
			level=cluster.LevelTask
		default:
			fmt.Println("-level must be either 'cluster','host', or 'task'")
			os.Exit(2)
		}
	case "ssh":
		sshCommand.Parse(os.Args[2:])
		switch *sshMode {
		case "tmux":
			mode=ecsssh.ModeTmux
		case "gui":
			mode=ecsssh.ModeGui
		case "batch":
			mode=ecsssh.ModeBatch
		default:
			fmt.Println("-mode must be either 'tmux','gui', or 'batch'")
			os.Exit(2)
		}

		switch *sshTarget {
		case "host":
			target=ecsssh.TargetHost
		case "task":
			target=ecsssh.TargetTask
		default:
			fmt.Println("-mode must be either 'tmux','gui', or 'batch'")
			os.Exit(2)
		}
	case "scp":
		scpCommand.Parse(os.Args[2:])
		switch *scpTarget {
		case "host":
			target=ecsssh.TargetHost
		case "task":
			target=ecsssh.TargetTask
		default:
			fmt.Println("-mode must be either 'tmux','gui', or 'batch'")
			os.Exit(2)
		}

	default:
		fmt.Printf("%q is invalid.\n",os.Args[1])
		os.Exit(2)
	}

	if listCommand.Parsed() {
		clusters := new(cluster.Clusters)
		clusters.List(*listClusterFilterFlag,*listHostFilterFlag,*listTaskFilterFlag,level,*listVerbosity)
	}
	if sshCommand.Parsed() {
		clusters := new(cluster.Clusters)
		ecsssh.EcsSSH(clusters,mode,target,sshClusterFilterFlag,sshHostFilterFlag,sshTaskFilterFlag,sshUserFlag,sshCmdToRun)
	}
	if scpCommand.Parsed() {
		clusters := new(cluster.Clusters)
		ecsssh.EcsSCP(clusters,target,scpClusterFilterFlag,scpHostFilterFlag,scpTaskFilterFlag,scpUserFlag,scpTargetDir,scpFileToSend,scpFileRunFlag,scpDeleteFileAfterRun)
	}
}


package ecsssh

import (
	"becs/cluster"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/aws"
	"fmt"
	"github.com/aws/aws-sdk-go/service/ecs"
	"github.com/aws/aws-sdk-go/service/ec2"
	"os/exec"
	"os"
	"strings"
)

func EcsSSH(c *cluster.Clusters,sshInteractive *bool, clusterFilter *string,hostFilter *string,taskFilter *string,user *string,password *string,toSend *string) {
	sess, err := session.NewSession(&aws.Config{Region: aws.String("us-east-1")})
	if err != nil {
		fmt.Println("failed to create session,", err)
		return
	}
	svc := ecs.New(sess)
	ec2_svc := ec2.New(sess)

	c.GetClusterInfo(svc,*clusterFilter)
	for i := 0; i < len(c.ClusterList); i++ {
		c.ClusterList[i].Hosts.GetHostInfo(svc, ec2_svc, c.ClusterList[i].Name, *hostFilter)
		for j :=0;j<len(c.ClusterList[i].Hosts.HostList);j++ {
			c.ClusterList[i].Hosts.HostList[j].Tasks.GetTaskInfo(svc,ec2_svc,c.ClusterList[i].Name, c.ClusterList[i].Hosts.HostList[j].Arn, *taskFilter)
		}
	}

	// check what the client is using for a gui to determine if we should limit to a single ssh call
	com1:=exec.Command("ps","-A")
	com2:=exec.Command("egrep","-i","\"gnome|kde|mate|cinnamon|lxde|xfce|jwm\"")
	com2.Stdin,_=com1.StdoutPipe()
	//com1.StdoutPipe()=com1.Stdout
	_=com2.Start()
	_=com1.Run()
	_=com2.Wait()
	//com2.Stdout=os.Stdout
	fmt.Println(com1.Output())

	for _, cluster := range c.ClusterList {
		for _, hostLoop := range cluster.Hosts.HostList {
			for _, taskElement := range hostLoop.Tasks.TaskList {
				//fmt.Println("ssh'ing to ", hostLoop.Ec2Ip, " and getting dockerIdName for", taskElement.Name)
				cmd:="docker ps |grep "+taskElement.Name+" | cut -d' ' -f1"
				sshOut,err:=exec.Command("ssh",*user+"@"+hostLoop.Ec2Ip,cmd).Output()
				if err!=nil {
					fmt.Printf("%v\n",err)
					os.Exit(2)
				}
				dockerId:=strings.TrimSpace(string(sshOut))

				if (*sshInteractive) {
					var sshSession *exec.Cmd
					terminal:="none"
					dockerCmd:="docker exec -it "+dockerId+" /bin/bash"
					cmdOut,_:=exec.Command("which","x-terminal-emulator").Output()
					if len(cmdOut)!=0 {
						//fmt.Println("ssh'ing to ", hostLoop.Ec2Ip, " with dockerIdName ",dockerId, " for", taskElement.Name)
						terminal="x-terminal-emulator"
					} else {
						cmdOut, err = exec.Command("which", "konsole").Output()
						if len(cmdOut) != 0 {
							terminal="konsole"
						} else {
							cmdOut, err = exec.Command("which", "xterm").Output()
							if len(cmdOut) != 0 {
								terminal="xterm"
							}
						}
					}
					if (terminal!="none") {
						sshSession = exec.Command(terminal, "-e", "ssh", "-tt", *user+"@"+hostLoop.Ec2Ip, dockerCmd)
					}

					sshSession.Stdout = os.Stdout
					sshSession.Stderr = os.Stderr
					sshSession.Stdin = os.Stdin
					errSession := sshSession.Run()
					if errSession != nil {
						fmt.Printf("errSession %v\n", errSession)
						os.Exit(2)
					}
				}
			}
		}

	}
}
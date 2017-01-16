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
	"runtime"
	"strconv"
	"time"
	"sync"
	"path/filepath"
	"github.com/mitchellh/go-homedir"
	"bufio"
)

type ModeType uint8

const (
	ModeTmux=ModeType(iota)
	ModeGui
	ModeBatch
)
func (s ModeType) String() string {
	var name = []string{"tmux","gui","batch"}
	var i = uint8(s)
	switch {
	case i <= uint8(ModeBatch):
		return name[i]
	default:
		return strconv.Itoa(int(i))
	}
}

type Target uint8

const (
	TargetHost=Target(iota)
	TargetTask
)
func (s Target) String() string {
	var name = []string{"host","task"}
	var i = uint8(s)
	switch {
	case i <= uint8(TargetTask):
		return name[i]
	default:
		return strconv.Itoa(int(i))
	}
}
func goEcsBatchSSH(clusterName string,user *string,Ec2Id string,Ec2Ip string,taskName string,dockerId string,toSend string,wg *sync.WaitGroup, ch *(chan []string)) {
	defer (*wg).Done()
	var sshOut []byte
	var err error
	var goBatchOutput=[]string{}
	if (dockerId != "") { //this is for a task
		sshOut, err = exec.Command("ssh",*user+"@"+Ec2Ip, "docker exec -t "+dockerId+" "+toSend).Output()
		if err != nil {
			fmt.Printf("%v\n", err)
			os.Exit(2)
		}
		goBatchOutput=append(goBatchOutput,"===== " + clusterName + ":" + Ec2Id + "(" + Ec2Ip + "):" + taskName + "(" + dockerId + ") ===== ")
	} else {
		sshOut, err = exec.Command("ssh",*user+"@"+Ec2Ip,toSend).Output()
		if err != nil {
			fmt.Printf("%v\n", err)
			os.Exit(2)
		}
		goBatchOutput=append(goBatchOutput,"===== " + clusterName + ":" + Ec2Id + "(" + Ec2Ip + ") ===== ")
	}
	goBatchOutput=append(goBatchOutput,"cmd: "+toSend)
	goBatchOutput=append(goBatchOutput,string(sshOut))
	*ch<-goBatchOutput
}

func modifyKnownHosts(ip string) {
	sshDir,_:=homedir.Dir()
	sshDir,_ = homedir.Expand(sshDir)
	sshDir = sshDir+"/.ssh"
	_ = exec.Command("ssh-keygen","-R",ip)
	sshKeyOut,_:= exec.Command("ssh-keyscan",ip).Output()
	if runtime.GOOS=="windows" {

	} else {
		f,_:=os.OpenFile(sshDir+"/known_hosts",os.O_RDWR|os.O_APPEND,0600)
		buffw:=bufio.NewWriter(f)
		defer f.Close()
		buffw.Write(sshKeyOut)
		buffw.Flush()
	}
}

func EcsSSH(c *cluster.Clusters,sshMode ModeType, sshTarget Target,clusterFilter *string,hostFilter *string,taskFilter *string,user *string,toSend *string) {
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
/*
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
*/
	var tmuxServer *exec.Cmd
	if (sshMode==ModeTmux) {
		cmdOut, _ := exec.Command("which", "tmux").Output()
		if len(cmdOut) == 0 {
			fmt.Println("Requires tmux to be installed.")
			os.Exit(2)
		}
		tmuxServer = exec.Command("tmux","new-session","-d","-s","becs")
		tmuxServerErr:=tmuxServer.Start()
		if tmuxServerErr !=nil {
			panic(err)
		}
		time.Sleep(100*time.Millisecond)  // give the tmux server time to start..otherwise panic/nil error when trying to run against it
	}
	var ch = make(chan []string)
	var wg sync.WaitGroup
	var batchOutput=[]string{}
	for _, cluster := range c.ClusterList {
		var sshOut []byte
		for _, hostLoop := range cluster.Hosts.HostList {
			modifyKnownHosts(hostLoop.Ec2Ip)
			sshOut=[]byte{}
			if (sshTarget==TargetTask) {
				for _, taskElement := range hostLoop.Tasks.TaskList {
					sshOut=[]byte{}
					cmd := "docker ps |grep " + taskElement.Name + " | cut -d' ' -f1"
					fmt.Println("ssh", *user+"@"+hostLoop.Ec2Ip, cmd)
					sshOut, err = exec.Command("ssh", *user+"@"+hostLoop.Ec2Ip, cmd).Output()
					if err != nil {
						fmt.Printf("%v\n", err)
						os.Exit(2)
					}
					dockerId := strings.TrimSpace(string(sshOut))
					dockerCmd := "docker exec -it " + dockerId + " /bin/bash"
					var sshSession *exec.Cmd
					extraArgs := []string{}
					var terminal string = "none"
					switch sshMode {
					case ModeGui:
						if runtime.GOOS == "windows" {
							//(windows10) sshSession = exec.Command("bash", "-c", "ssh", "-tt", *user+"@"+hostLoop.Ec2Ip, dockerCmd)
							args := append(extraArgs, []string{"/C","start","ssh","-tt", *user + "@" + hostLoop.Ec2Ip, dockerCmd}...)
							sshSession = exec.Command("cmd",args...)
							sshSession.Stdout = os.Stdout
							sshSession.Stderr = os.Stderr
							sshSession.Stdin = os.Stdin
							errSession := sshSession.Run()
							if errSession != nil {
								fmt.Printf("errSession %v\n", errSession)
								os.Exit(2)
							}
						} else {
							cmdOut, _ := exec.Command("which", "x-terminal-emulator").Output()
							if len(cmdOut) != 0 {
								terminal = "x-terminal-emulator"
							} else {
								cmdOut, err = exec.Command("which", "konsole").Output()
								if len(cmdOut) != 0 {
									terminal = "konsole"
								} else {
									cmdOut, err = exec.Command("which", "xterm").Output()
									if len(cmdOut) != 0 {
										terminal = "xterm"
									}
								}
							}

							if (terminal != "none") {
								args := append(extraArgs, []string{"-e", "ssh", "-tt", *user + "@" + hostLoop.Ec2Ip, dockerCmd}...)
								sshSession = exec.Command(terminal, args...)

								sshSession.Stdout = os.Stdout
								sshSession.Stderr = os.Stderr
								sshSession.Stdin = os.Stdin
								errSession := sshSession.Run()
								if errSession != nil {
									fmt.Printf("errSession %v\n", errSession)
									os.Exit(2)
								}
							} else { // single terminal...only allow one ssh session
								args := append(extraArgs, []string{"-tt", *user + "@" + hostLoop.Ec2Ip, dockerCmd}...)
								sshSession = exec.Command("ssh", args...)

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
					case ModeTmux:
						args := append(extraArgs, []string{"new-window", "-t", "becs", "-n", hostLoop.Ec2Id[2:7] + ":" + taskElement.Name[0:9], "ssh","-tt", *user + "@" + hostLoop.Ec2Ip, dockerCmd}...)
						tmuxSession := exec.Command("tmux", args...)
						tmuxErr := tmuxSession.Run()
						if tmuxErr != nil {
							panic(err)
						}

					case ModeBatch:
						//if runtime.GOOS == "windows" {
						//	sshOut, err = exec.Command("bash", "-c", "'ssh "+*user+"@"+hostLoop.Ec2Ip+" "+*toSend+"'").Output()
						//	if err != nil {
						//		fmt.Printf("%v\n", err)
						//		os.Exit(2)
						//	}
						//} else {
							wg.Add(1)
							go goEcsBatchSSH(cluster.Name,user,hostLoop.Ec2Id,hostLoop.Ec2Ip,taskElement.Name,dockerId,*toSend,&wg,&ch)
						//}
					}
				}
			} else { //just ssh to the Hosts
				for _, taskElement := range hostLoop.Tasks.TaskList {
					if (strings.Contains(taskElement.Name, *taskFilter)) {
						var sshSession *exec.Cmd
						extraArgs := []string{}
						var terminal string = "none"
						switch sshMode {
						case ModeGui:
							if runtime.GOOS == "windows" {
								args := append(extraArgs, []string{"/C","start","ssh","-tt", *user + "@" + hostLoop.Ec2Ip}...)
								fmt.Println(args)
								sshSession = exec.Command("cmd",args...)
								sshSession.Stdout = os.Stdout
								sshSession.Stderr = os.Stderr
								sshSession.Stdin = os.Stdin
								errSession := sshSession.Run()
								if errSession != nil {
									fmt.Printf("errSession %v\n", errSession)
									os.Exit(2)
								}
							} else {
								cmdOut, _ := exec.Command("which", "x-terminal-emulator").Output()
								if len(cmdOut) != 0 {
									terminal = "x-terminal-emulator"
								} else {
									cmdOut, err = exec.Command("which", "konsole").Output()
									if len(cmdOut) != 0 {
										terminal = "konsole"
									} else {
										cmdOut, err = exec.Command("which", "xterm").Output()
										if len(cmdOut) != 0 {
											terminal = "xterm"
										}
									}
								}

								if (terminal != "none") {
									args := append(extraArgs, []string{"-e", "ssh", "-t", *user + "@" + hostLoop.Ec2Ip}...)
									sshSession = exec.Command(terminal, args...)

									sshSession.Stdout = os.Stdout
									sshSession.Stderr = os.Stderr
									sshSession.Stdin = os.Stdin
									errSession := sshSession.Run()
									if errSession != nil {
										fmt.Printf("errSession %v\n", errSession)
										os.Exit(2)
									}
								} else { // single terminal...only allow one ssh session
									args := append(extraArgs, []string{"-tt", *user + "@" + hostLoop.Ec2Ip}...)
									sshSession = exec.Command("ssh", args...)

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
						case ModeTmux:
							args := append(extraArgs, []string{"new-window", "-t", "becs", "-n", hostLoop.Ec2Id[2:7], "ssh", "-tt", *user + "@" + hostLoop.Ec2Ip}...)
							tmuxSession := exec.Command("tmux", args...)
							tmuxErr := tmuxSession.Run()
							if tmuxErr != nil {
								panic(err)
							}

						case ModeBatch:
							wg.Add(1)
							go goEcsBatchSSH(cluster.Name,user,hostLoop.Ec2Id,hostLoop.Ec2Ip,taskElement.Name,"",*toSend,&wg,&ch)
						}
						break
					}
				}
			}
		}
	}
	if (sshMode==ModeBatch) {
		go func() {
			wg.Wait()
			close(ch)
		}()
		for m := range ch {
			batchOutput=append(batchOutput,m...)
		}
		for i := range batchOutput {
			fmt.Println(batchOutput[i])
		}
	}
	if (sshMode==ModeTmux) {
		tmuxRoot := exec.Command("tmux","attach-session","-t","becs")
		tmuxRoot.Stdin = os.Stdin
		tmuxRoot.Stdout = os.Stdout
		tmuxRoot.Stderr = os.Stderr
		fmt.Println("before starting root tmux session")
		tmuxRootErr:=tmuxRoot.Run()
		fmt.Println("started root tmux session")
		if tmuxRootErr !=nil {
			panic(err)
		}
	}
}

func goEcsBatchSCP(clusterName string,user *string,Ec2Id string,Ec2Ip string,taskName string,dockerId string,targetDir *string,toSend *string, runFlag *bool, deleteFlag *bool, wg *sync.WaitGroup, ch *(chan []string)) {
	defer (*wg).Done()
	var sshOut []byte
	var err error
	var goBatchOutput=[]string{}
	_,fileName:=filepath.Split(*toSend)
	if (dockerId != "") { //this is for a task
		hostFileRename:=dockerId+"-"+fileName
		sshOut, err = exec.Command("scp", *toSend,*user+"@"+Ec2Ip+":"+"/tmp/"+hostFileRename).Output()  //prepend the dockerID to prevent collisions for big batches
		if err != nil {
			fmt.Printf("%v\n", err)
			os.Exit(2)
		}
		sshOut, err = exec.Command("ssh", *user+"@"+Ec2Ip, "docker cp /tmp/"+hostFileRename+" "+dockerId+":"+*targetDir+"/"+fileName+" && rm -f /tmp/"+hostFileRename).Output() //prepend the dockerID on the host
		if err != nil {
			fmt.Printf("%v\n", err)
			os.Exit(2)
		}
		goBatchOutput=append(goBatchOutput,"===== " + clusterName + ":" + Ec2Id + "(" + Ec2Ip + "):" + taskName + "(" + dockerId + ") ===== ")
		goBatchOutput=append(goBatchOutput,"file: "+*toSend+ " put to "+Ec2Id + "(" + Ec2Ip + "):" + taskName + "(" + dockerId + ")"+" directory "+*targetDir)
	} else {
		sshOut, err = exec.Command("scp",*toSend, *user+"@"+Ec2Ip+":"+*targetDir).Output()
		if err != nil {
			fmt.Printf("%v\n", err)
			os.Exit(2)
		}
		goBatchOutput=append(goBatchOutput,"===== " + clusterName + ":" + Ec2Id + "(" + Ec2Ip + ") =====")
		goBatchOutput=append(goBatchOutput,"file: "+*toSend+ " put to directory "+*targetDir)
	}

	if (*runFlag) {
		if (dockerId != "") {
			sshOut, err = exec.Command("ssh", *user+"@"+Ec2Ip, "docker exec -t "+dockerId+" "+*targetDir+"/"+fileName).Output()
			if err != nil {
				fmt.Printf("%v\n", err)
				os.Exit(2)
			}
			goBatchOutput=append(goBatchOutput,"Output of run:")
			goBatchOutput=append(goBatchOutput,string(sshOut))
			if (*deleteFlag) {
				sshOut, err = exec.Command("ssh",*user+"@"+Ec2Ip, "docker exec -t "+dockerId+" rm -f "+*targetDir+"/"+fileName).Output() //delete file
				goBatchOutput=append(goBatchOutput,"Deleted "+*targetDir+"/"+fileName+" from "+dockerId)
			}
		}else {
			sshOut, err = exec.Command("ssh", *user+"@"+Ec2Ip, *targetDir+"/"+fileName).Output()
			if err != nil {
				fmt.Printf("%v\n", err)
				os.Exit(2)
			}
			goBatchOutput=append(goBatchOutput,"Output of run:")
			goBatchOutput=append(goBatchOutput,string(sshOut))
			if (*deleteFlag) {
				sshOut, err = exec.Command("ssh", *user+"@"+Ec2Ip,"rm -f "+*targetDir+"/"+fileName).Output() //delete file
				goBatchOutput=append(goBatchOutput,"Deleted "+*targetDir+"/"+fileName+" from "+Ec2Id)
			}
		}
	}
	*ch<-goBatchOutput
}

func EcsSCP(c *cluster.Clusters, sshTarget Target, clusterFilter *string,hostFilter *string,taskFilter *string,user *string,targetDir *string, toSend *string, runFlag *bool, deleteFlag *bool) {
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
	var ch = make(chan []string)
	var wg sync.WaitGroup
	var batchOutput=[]string{}
	for _, cluster := range c.ClusterList {
		var sshOut []byte
		for _, hostLoop := range cluster.Hosts.HostList {
			modifyKnownHosts(hostLoop.Ec2Ip)
			sshOut=[]byte{}
			if (sshTarget==TargetTask) {
				for _, taskElement := range hostLoop.Tasks.TaskList {
					sshOut=[]byte{}
					cmd := "docker ps |grep " + taskElement.Name + " | cut -d' ' -f1"
					sshOut, err = exec.Command("ssh",*user+"@"+hostLoop.Ec2Ip, cmd).Output()
					if err != nil {
						fmt.Printf("%v\n", err)
						os.Exit(2)
					}
					dockerId := strings.TrimSpace(string(sshOut))
					wg.Add(1)
					go goEcsBatchSCP(cluster.Name,user,hostLoop.Ec2Id,hostLoop.Ec2Ip,taskElement.Name,dockerId,targetDir,toSend,runFlag, deleteFlag, &wg,&ch)
				}
			} else { //just sftp to the Hosts
				for _, taskElement := range hostLoop.Tasks.TaskList {
					if (strings.Contains(taskElement.Name, *taskFilter)) {
						for _, taskElement := range hostLoop.Tasks.TaskList {  // copy the file to the host first
							if (strings.Contains(taskElement.Name, *taskFilter)) {
								wg.Add(1)
								go goEcsBatchSCP(cluster.Name, user, hostLoop.Ec2Id, hostLoop.Ec2Ip, taskElement.Name, "", targetDir, toSend, runFlag, deleteFlag, &wg, &ch)
								break
							}
						}
					}
				}
			}
		}
	}
	go func() {
		wg.Wait()
		close(ch)
	}()
	for m := range ch {
		batchOutput=append(batchOutput,m...)
	}
	for i := range batchOutput {
		fmt.Println(batchOutput[i])
	}
}
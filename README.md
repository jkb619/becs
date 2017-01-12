# becs
Batch ECS (ssh to multiple ecs containers)

*usage*

* becs 
  * list
     * -cluster string cluster substring to match
     * -host string host substring to match
     * -level string what level to delve: cluster/host/task (defaults to task) (default "task")
     * -task string task substring to match
     * -v bool verbose (dump arns as well as names)
 
  * ssh
    * -cluster string cluster substring to match
    * -cmd string what cmd to send via ssh
    * -host string host substring to match
    * -mode string tmux / gui / batch.  (default "tmux")
    * -target string host/task. Identifies which elements to ssh to. (default "task")
    * -task string task substring to match
    * -user string user to login as (default "ec2-user")

  * scp
    * -cluster string cluster substring to match
    * -host string host substring to match
    * -target string host/task. Identifies which elements to ssh to. (default "task")
    * -task string task substring to match
    * -user string user to login as (default "ec2-user")
    * -tdir string target directory for file (default "/tmp")
    * -x bool execute file/script (default false)
    * -d bool delete file/script after running (default false)    
    
* examples *

* *list all hosts and their tasks in cluster 'testcluster'*
  * becs list -cluster=testluster
* *list just hosts in container 'testcluster' running task 'task1'*
  * becs list -cluster=testcluster task=task1 -level=host
* *list everything (all clusters, hosts, and tasks), including arns*
  * becs list -v=true
    
* *ssh to all tasks containing the string 'mytask' in cluster 'testcluster'*
  * becs ssh -cluster=testcluster -task=mytask
* *ssh to all hosts containing tasks with the string 'mytask' in cluster 'testcluster'*
  * becs ssh -cluster=testcluster -task=mytask -target=hosts
* *ssh to all hosts containing tasks with the string 'mytask' in all environments using gui/desktop*
  * becs ssh -task=mytask -target=hosts -mode=gui
* *get a listing of all running process from all hosts in cluster 'testcluster'*
  * becs ssh -mode=batch -cluster=testcluster -target=host -cmd="ps -ef"
* *get a listing of all running process from all 'mytask's in cluster 'testcluster'*
  * becs ssh -mode=batch -cluster=testcluster -target=task -task=mytask -cmd="ps -ef"    
    
Windows (ubuntu-on-windows) not tested/worked on yet.
Darwin not tested at all.
# becs
Batch ECS (ssh to multiple ecs containers)

usage:
becs
  list
     -cluster string
       	cluster substring to match
     -host string
       	host substring to match
     -level string
       	what level to delve: cluster/task (defaults to task) (default "task")
     -task string
       	task substring to match
 
  ssh
    -cluster string
      	cluster substring to match
    -cmd string
      	what cmd to send via ssh
    -host string
      	host substring to match
    -mode string
      	tmux / gui / batch.  (default "tmux")
    -password string
      	password for user
    -target string
      	host/task. Identifies which elements to ssh to. (default "task")
    -task string
      	task substring to match
    -user string
      	user to login as (default "ec2-user")

  scp
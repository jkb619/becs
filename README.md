# becs
Batch ECS (ssh to multiple ecs containers)

becs
  container
    list
    ssh <container_blob> <cmd>
    scp <container_blob> <cmd>
  host
    list
    ssh <host_blob> <cmd>
    scp <host_blob> <cmd>
  cluster
    list <-v also sub hosts, -vv hosts+containers>
    ssh <cluster_blob> <cmd>
    scp <cluster_blob> <cmd>
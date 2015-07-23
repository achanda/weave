#! /bin/bash

. ./config.sh

C1=10.2.0.78
C2=10.2.0.34
NAME=seetwo.weave.local

check_attached() {
    assert_raises "proxy exec_on $HOST1 c2 $CHECK_ETHWE_UP"
    assert_dns_record $HOST1 c1 $NAME $C2
}

start_suite "Proxy restart reattaches networking to containers"

weave_on $HOST1 launch
proxy docker_on $HOST1 run -e WEAVE_CIDR=$C2/24 -di --name=c2 --restart=always -h $NAME $SMALL_IMAGE /bin/sh
proxy docker_on $HOST1 run -e WEAVE_CIDR=$C1/24 -di --name=c1 --restart=always          $DNS_IMAGE   /bin/sh

proxy docker_on $HOST1 restart c2
check_attached

# Kill outside of Docker so Docker will restart it
run_on $HOST1 sudo kill $(docker_on $HOST1 inspect --format='{{.State.Pid}}' c2)
check_attached

# Restart docker itself
# - disabled since the commands are different between our Vagrant VMs and GCE
#run_on $HOST1 sudo systemctl restart docker # for systemd
#run_on $HOST1 sudo service docker restart # for upstart
#weave_on $HOST1 launch
#check_attached

end_suite

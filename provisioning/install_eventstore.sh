#! /bin/bash

echo ">>> Installing EventStore"

# Fixing timezone issue
timedatectl set-timezone $2

# install eventstore
wget -q http://download.geteventstore.com/binaries/EventStore-OSS-Ubuntu-v3.3.1.tar.gz
tar xf EventStore-OSS-Ubuntu-v3.3.1.tar.gz
mv "EventStore-OSS-Ubuntu-14.04-v3.3.1" /opt/eventstore

mkdir -p /var/eventstore/data
mkdir -p /var/eventstore/logs

echo "#!/bin/bash

case \$1 in
start)
    pushd .
    cd /opt/eventstore
    ./run-node.sh --log /var/eventstore/logs --db /var/eventstore/data --ext-ip=$1 --run-projections=All > /dev/null &
    popd
    ;;
stop)
    sudo pkill eventstored
    ;;
esac
" > /etc/init.d/eventstore

chmod +x /etc/init.d/eventstore
update-rc.d eventstore defaults
service eventstore start